package vmm

import "github.com/hashicorp/packer/helper/multistep"

func isCancelled(state multistep.StateBag) bool {
	_, cancelled := state.GetOk(multistep.StateCancelled)
	return cancelled
}

func isHalted(state multistep.StateBag) bool {
	_, halted := state.GetOk(multistep.StateHalted)
	return halted
}
