package main

import (
	"log"

	"github.com/prep/vmm"

	"github.com/hashicorp/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		log.Fatal(err)
	}

	server.RegisterBuilder(new(vmm.Builder))
	server.Serve()
}
