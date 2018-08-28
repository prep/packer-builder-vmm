package vmm

import (
	"errors"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/packer/common"
	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/helper/config"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

// BuilderId references the unique ID of this builder.
const BuilderId = "prep.vmm"

// Config describes the configuration of this builder.
type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	common.HTTPConfig      `mapstructure:",squash"`
	common.ISOConfig       `mapstructure:",squash"`
	bootcommand.BootConfig `mapstructure:",squash"`

	VMName    string `mapstructure:"vm_name"`
	BiosFile  string `mapstructure:"bios_file"`
	DiskSize  string `mapstructure:"disk_size"`
	MemSize   string `mapstructure:"mem_size"`
	OutputDir string `mapstructure:"output_directory"`

	ctx interpolate.Context
}

// Builder is responsible for create a machine and generating an image.
type Builder struct {
	config Config
	runner multistep.Runner
}

// Prepare the build configuration parameters.
func (builder *Builder) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(&builder.config, &config.DecodeOpts{
		Interpolate:        true,
		InterpolateContext: &builder.config.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
				"vmm_args",
			},
		},
	}, raws...)
	if err != nil {
		return nil, err
	}

	if builder.config.VMName == "" {
		builder.config.VMName = "packer-" + builder.config.PackerBuildName
	}

	if builder.config.DiskSize == "" {
		builder.config.DiskSize = "10G"
	}

	if builder.config.MemSize == "" {
		builder.config.MemSize = "512M"
	}

	if builder.config.OutputDir == "" {
		builder.config.OutputDir = "output-vmm"
	}

	errs := &packer.MultiError{Errors: builder.config.BootConfig.Prepare(&builder.config.ctx)}
	errs = packer.MultiErrorAppend(errs, builder.config.HTTPConfig.Prepare(&builder.config.ctx)...)
	warnings, isoErrs := builder.config.ISOConfig.Prepare(&builder.config.ctx)
	errs = packer.MultiErrorAppend(errs, isoErrs...)

	if errs != nil && len(errs.Errors) > 0 {
		return warnings, errs
	}

	return warnings, nil
}

// Run a packer build and returns a packer.Artifact representing a raw image.
func (builder *Builder) Run(ui packer.Ui, hook packer.Hook, cache packer.Cache) (packer.Artifact, error) {
	driver, err := builder.newDriver(ui)
	if err != nil {
		return nil, err
	}

	var steps []multistep.Step

	// Download the ISO.
	steps = append(steps, &common.StepDownload{
		Checksum:     builder.config.ISOChecksum,
		ChecksumType: builder.config.ISOChecksumType,
		Description:  "ISO",
		ResultKey:    "iso_path",
		TargetPath:   builder.config.TargetPath,
		Url:          builder.config.ISOUrls,
		Extension:    builder.config.TargetExtension,
	})

	// Create output directory.
	steps = append(steps, &stepOutputDir{
		outputPath: builder.config.OutputDir,
		force:      builder.config.PackerForce,
	})

	diskPath := filepath.Join(builder.config.OutputDir, builder.config.VMName) + ".raw"

	// Create the disk.
	steps = append(steps, &stepCreateDisk{
		diskPath: diskPath,
		diskSize: builder.config.DiskSize,
	})

	// Set up the HTTP server.
	steps = append(steps, &common.StepHTTPServer{
		HTTPDir:     builder.config.HTTPDir,
		HTTPPortMin: builder.config.HTTPPortMin,
		HTTPPortMax: builder.config.HTTPPortMax,
	})

	// Start the VM.
	steps = append(steps, &stepRun{
		vmName:   builder.config.VMName,
		diskPath: diskPath,
		memSize:  builder.config.MemSize,
	})

	// Execute the boot command sequence.
	steps = append(steps, &stepBootCommand{
		bootWait:    builder.config.BootWait,
		bootCommand: builder.config.FlatBootCommand(),
		ctx:         builder.config.ctx,
	})

	// Set up a state bag.
	state := new(multistep.BasicStateBag)
	state.Put("ui", ui)
	state.Put("hook", hook)
	state.Put("cache", cache)
	state.Put("config", &builder.config)
	state.Put("driver", driver)

	// Run the jewels.
	builder.runner = common.NewRunnerWithPauseFn(steps, builder.config.PackerConfig, ui, state)
	builder.runner.Run(state)

	// If we were interrupted or cancelled, then just exit.
	if _, ok := state.GetOk(multistep.StateCancelled); ok {
		return nil, errors.New("Build was cancelled.")
	}

	if _, ok := state.GetOk(multistep.StateHalted); ok {
		return nil, errors.New("Build was halted.")
	}

	artifact := &Artifact{
		directory: builder.config.OutputDir,
		files:     []string{diskPath},
	}

	return artifact, nil
}

// Cancel the build process.
func (builder *Builder) Cancel() {
	builder.runner.Cancel()
}

func (builder *Builder) newDriver(ui packer.Ui) (Driver, error) {
	doasPath, err := exec.LookPath("doas")
	if err != nil {
		return nil, err
	}

	vmctlPath, err := exec.LookPath("vmctl")
	if err != nil {
		return nil, err
	}

	return &VmmDriver{doasPath: doasPath, vmctlPath: vmctlPath, ui: ui}, nil
}
