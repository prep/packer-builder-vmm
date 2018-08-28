package vmm

import (
	"fmt"
	"os"
)

// Artifact contains references to the things that a build has created.
type Artifact struct {
	directory string
	files     []string
}

// BuilderId returns the ID of the builder that is used to create this artifact.
func (artifact *Artifact) BuilderId() string {
	return BuilderId
}

// Files returns a list of files in this artifact.
func (artifact *Artifact) Files() []string {
	return artifact.files
}

// Id of this artifact.
func (artifact *Artifact) Id() string {
	return "VMM"
}

func (artifact *Artifact) String() string {
	return fmt.Sprintf("VM files in directory: %s", artifact.files)
}

// State returns builder-specific state information.
func (artifact *Artifact) State(name string) interface{} {
	return nil
}

// Destroy this artifact.
func (artifact *Artifact) Destroy() error {
	return os.RemoveAll(artifact.directory)
}
