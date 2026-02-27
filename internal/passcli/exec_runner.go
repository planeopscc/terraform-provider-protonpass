// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package passcli

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

// ExecRunner runs pass-cli commands via os/exec.
type ExecRunner struct {
	cliPath string
	timeout time.Duration
}

// NewExecRunner creates a runner that shells out to the pass-cli binary.
func NewExecRunner(cliPath string, timeout time.Duration) *ExecRunner {
	return &ExecRunner{cliPath: cliPath, timeout: timeout}
}

// Run executes pass-cli with the given arguments.
func (r *ExecRunner) Run(ctx context.Context, args ...string) ([]byte, []byte, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	redacted := RedactArgs(args)
	_ = redacted // Available for debug logging.

	cmd := exec.CommandContext(ctx, r.cliPath, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		exitCode := -1
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
		return stdout.Bytes(), stderr.Bytes(), &CLIError{
			Args:     args,
			ExitCode: exitCode,
			Stderr:   stderr.String(),
			Err:      fmt.Errorf("pass-cli %v exited with code %d: %s", redacted, exitCode, stderr.String()),
		}
	}
	return stdout.Bytes(), stderr.Bytes(), nil
}
