package vmm

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/helper/multistep"
	"github.com/hashicorp/packer/packer"
	"github.com/hashicorp/packer/template/interpolate"
)

type stepBootCommand struct {
	bootWait    time.Duration
	bootCommand string
	ctx         interpolate.Context
}

func (step *stepBootCommand) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packer.Ui)
	driver := state.Get("driver").(Driver)

	if step.bootWait > 0 {
		ui.Say(fmt.Sprintf("Waiting %s for boot...", step.bootWait.String()))
		select {
		case <-time.After(step.bootWait):
			break
		case <-ctx.Done():
			return multistep.ActionHalt
		}
	}

	ui.Say("Typing the boot command...")
	command, err := interpolate.Render(step.bootCommand, &step.ctx)
	if err != nil {
		state.Put("error", fmt.Errorf("Error preparing boot command: %s", err))
		return multistep.ActionHalt
	}

	seq, err := bootcommand.GenerateExpressionSequence(command)
	if err != nil {
		state.Put("error", fmt.Errorf("Error generating boot command: %s", err))
		return multistep.ActionHalt
	}

	if err := seq.Do(ctx, driver); err != nil {
		state.Put("error", fmt.Errorf("Error running boot command: %s", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (step *stepBootCommand) Cleanup(state multistep.StateBag) {}
