package vmm

import (
	"context"

	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
)

type stepCreateDisk struct {
	diskPath   string
	diskFormat string
	diskSize   string
	resultKey  string
}

func (step *stepCreateDisk) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(Driver)

	ui.Say("Creating image " + step.diskPath)
	if err := driver.CreateDisk(step.diskFormat+":"+step.diskPath, step.diskSize); err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (step *stepCreateDisk) Cleanup(state multistep.StateBag) {}
