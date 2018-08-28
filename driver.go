package vmm

import (
	"errors"
	"io"
	"os/exec"
	"strings"
	"time"

	"github.com/hashicorp/packer/common/bootcommand"
	"github.com/hashicorp/packer/packer"
)

var errNoTTY = errors.New("unable to determine instance tty")

// Driver describes a VMM driver.
type Driver interface {
	bootcommand.BCDriver

	CreateDisk(path, size string) error
	Start(name string, args ...string) error
	Stop(name string) error
}

// VmmDriver manages a vmm instance.
type VmmDriver struct {
	doasPath  string
	vmctlPath string
	tty       io.WriteCloser
	ui        packer.Ui
}

// CreateDisk creates a new disk image.
func (driver *VmmDriver) CreateDisk(path, size string) error {
	args := []string{"create", path, "-s", size}
	driver.ui.Message("Executing " + driver.vmctlPath + " " + strings.Join(args, " "))

	return exec.Command(driver.vmctlPath, args...).Run()
}

// Start the VM and create a pipe to insert commands into the VM.
func (driver *VmmDriver) Start(name string, args ...string) error {
	args = append([]string{driver.vmctlPath, "start", name}, args...)
	driver.ui.Message("Executing " + driver.doasPath + " " + strings.Join(args, " "))

	// Start up the VM.
	if err := exec.Command(driver.doasPath, args...).Run(); err != nil {
		return err
	}

	// Give the VM a bit of time to come up.
	time.Sleep(3 * time.Second)

	// Ask for the path of the pseudo TTY.
	devicePath, err := driver.devicePath(name)
	if err != nil {
		return err
	}

	driver.ui.Message("Executing " + driver.doasPath + " tee -a " + devicePath)

	// Get a write-only socket to the pseudo TTY.
	cmd := exec.Command(driver.doasPath, "tee", "-a", devicePath)
	if driver.tty, err = cmd.StdinPipe(); err != nil {
		return err
	}

	return cmd.Start()
}

// devicePath returns the path to the pseudo TTY device.
func (driver *VmmDriver) devicePath(name string) (string, error) {
	output, err := exec.Command(driver.doasPath, driver.vmctlPath, "status", name).Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return "", errNoTTY
	}

	fields := strings.Fields(lines[0])
	values := strings.Fields(lines[1])

	if len(fields) != len(values) {
		return "", errNoTTY
	}

	for i, field := range fields {
		if field == "TTY" {
			return "/dev/" + values[i], nil
		}
	}

	return "", errNoTTY
}

// Stop the VM and close down the input pipe.
func (driver *VmmDriver) Stop(name string) error {
	driver.ui.Message("Executing " + driver.doasPath + " " + driver.vmctlPath + " stop " + name)

	err := exec.Command(driver.doasPath, driver.vmctlPath, "stop", name).Run()
	if err != nil {
		return err
	}

	return driver.tty.Close()
}

// SendKey sends a key press.
func (driver *VmmDriver) SendKey(key rune, action bootcommand.KeyAction) error {
	_, err := driver.tty.Write([]byte{byte(key)})
	return err
}

// SendSpecial sends a special character.
func (driver *VmmDriver) SendSpecial(special string, action bootcommand.KeyAction) error {
	var data []byte
	switch special {
	case "enter":
		data = []byte("\n")
	}

	if len(data) != 0 {
		_, err := driver.tty.Write([]byte{'\n'})
		return err
	}

	return nil
}

// Flush doesn't do anything here.
func (driver *VmmDriver) Flush() error {
	return nil
}
