package vmm

import (
	"context"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepRun struct {
	vmName     string
	diskPath   string
	diskFormat string
	memSize    string
}

func (step *stepRun) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(Driver)
	isoPath := state.Get("iso_path").(string)

	args := []string{"-c", "-L", "-d", isoPath, "-d", step.diskFormat + ":" + step.diskPath, "-m", step.memSize}

	ui.Say("Starting VMM instance " + step.vmName)
	if err := driver.Start(step.vmName, args...); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (step *stepRun) Cleanup(state multistep.StateBag) {
	driver := state.Get("driver").(Driver)
	driver.Stop(step.vmName)
}
