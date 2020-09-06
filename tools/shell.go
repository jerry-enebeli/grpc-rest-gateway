package tools

import (
	"bytes"
	"os/exec"
)

// Shell runs shell commands and return the output of the command and the command error
func Shell(shell, command string) (output string, err error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(shell, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()

	return stdout.String(), err
}
