package vmm

import (
	"errors"
	"io"
	"os"
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
	logPath   string
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
	driver.ui.Message("Logging console output to " + driver.logPath)
	logFile, err := os.Create(driver.logPath)
	if err != nil {
		return err
	}

	args = append([]string{driver.vmctlPath, "start", name}, args...)
	driver.ui.Message("Executing " + driver.doasPath + " " + strings.Join(args, " "))

	cmd := exec.Command(driver.doasPath, args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	// Create an stdin pipe that is used to issue commands.
	if driver.tty, err = cmd.StdinPipe(); err != nil {
		return err
	}

	// Write the console output to the log file.
	go func() {
		defer stdout.Close()
		defer logFile.Close()

		_, _ = io.Copy(logFile, stdout)
	}()

	// Start up the VM.
	if err := cmd.Start(); err != nil {
		return err
	}

	// Give the VM a bit of time to start up.
	time.Sleep(3 * time.Second)
	return nil
}

// Stop the VM and close down the input pipe.
func (driver *VmmDriver) Stop(name string) error {
	driver.ui.Message("Executing " + driver.doasPath + " " + driver.vmctlPath + " stop " + name)

	err := exec.Command(driver.doasPath, driver.vmctlPath, "stop", name).Run()
	if err != nil {
		return err
	}

	if driver.tty != nil {
		return driver.tty.Close()
	}

	return nil
}

// SendKey sends a key press.
func (driver *VmmDriver) SendKey(key rune, action bootcommand.KeyAction) error {
	data := []byte{byte(key)}

	_, err := driver.tty.Write(data); err != nil {
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
		if _, err := driver.tty.Write(data); err != nil {
			return err
		}
	}

	return nil
}

// Flush doesn't do anything here.
func (driver *VmmDriver) Flush() error {
	return nil
}
