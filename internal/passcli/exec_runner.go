// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package passcli

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// ExecRunner runs pass-cli commands via os/exec.
type ExecRunner struct {
	cliPath  string
	timeout  time.Duration
	extraEnv []string // additional env vars prepended to subprocess env
}

// NewExecRunner creates a runner that shells out to the pass-cli binary.
func NewExecRunner(cliPath string, timeout time.Duration) *ExecRunner {
	return &ExecRunner{cliPath: cliPath, timeout: timeout}
}

// NewExecRunnerWithEnv creates a runner that injects extra environment variables
// into each subprocess. Values in extraEnv take priority over any same-key
// variables already present in the process environment (first match wins in
// the subprocess's getenv).
func NewExecRunnerWithEnv(cliPath string, timeout time.Duration, extraEnv []string) *ExecRunner {
	return &ExecRunner{cliPath: cliPath, timeout: timeout, extraEnv: extraEnv}
}

// Run executes pass-cli with the given arguments.
func (r *ExecRunner) Run(ctx context.Context, args ...string) ([]byte, []byte, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	redacted := RedactArgs(args)
	_ = redacted // Available for debug logging.

	cmd := exec.CommandContext(ctx, r.cliPath, args...)
	if len(r.extraEnv) > 0 {
		// Build a set of keys we are injecting so we can remove any same-key
		// entries from the process environment. extraEnv values are appended
		// last, making them the sole occurrence and therefore authoritative.
		override := make(map[string]struct{}, len(r.extraEnv))
		for _, e := range r.extraEnv {
			if k, _, ok := strings.Cut(e, "="); ok {
				override[k] = struct{}{}
			}
		}
		base := os.Environ()
		env := make([]string, 0, len(base)+len(r.extraEnv))
		for _, e := range base {
			if k, _, ok := strings.Cut(e, "="); ok {
				if _, skip := override[k]; skip {
					continue
				}
			}
			env = append(env, e)
		}
		cmd.Env = append(env, r.extraEnv...)
	}
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
