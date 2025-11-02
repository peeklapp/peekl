package utils

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type ExecutionErrorDetails struct {
	Command  string
	ExitCode int
	Stderr   string
}

func (e ExecutionErrorDetails) Error() string {
	return fmt.Sprintf("Error during execution of the following command : %s", e.Command)
}

type ExecutionOutput struct {
	Stdout       string
	ErrorDetails ExecutionErrorDetails
}

func Execute(command string, args ...string) *ExecutionOutput {
	// Declare output variables
	var execOutput ExecutionOutput

	// Prepare command
	var stdoutBuff bytes.Buffer
	var stderrBuff bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Stdout = &stdoutBuff
	cmd.Stderr = &stderrBuff

	// Execute command
	err := cmd.Run()
	execOutput.Stdout = stdoutBuff.String()

	// Process error if any
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			execOutput.ErrorDetails = ExecutionErrorDetails{
				Command:  fmt.Sprintf("%s %s", command, strings.Join(args, " ")),
				Stderr:   stderrBuff.String(),
				ExitCode: exitError.ExitCode(),
			}
		}
	}

	return &execOutput
}
