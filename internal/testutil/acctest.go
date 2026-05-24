// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package testutil

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/planeopscc/terraform-provider-protonpass/internal/passcli"
)

// SkipIfNotAcc skips t unless TF_ACC is set to a non-empty value.
// All acceptance tests must call this at the top of the test function.
func SkipIfNotAcc(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance tests skipped: set TF_ACC=1 to run")
	}
}

// AccCLIPath returns the path to the pass-cli binary for acceptance tests.
// Uses PROTONPASS_ACC_CLI_PATH if set, otherwise defaults to "pass-cli".
func AccCLIPath() string {
	if p := os.Getenv("PROTONPASS_ACC_CLI_PATH"); p != "" {
		return p
	}
	return "pass-cli"
}

// NewAccClient creates a real passcli.Client backed by the live pass-cli binary.
// Skips t if the binary is not found (PATH lookup fails).
func NewAccClient(t *testing.T) *passcli.Client {
	t.Helper()
	cliPath := AccCLIPath()
	if _, err := exec.LookPath(cliPath); err != nil {
		t.Skipf("pass-cli not found at %q: %v — install it and authenticate with 'pass-cli login'", cliPath, err)
	}
	runner := passcli.NewExecRunner(cliPath, 30*time.Second)
	return passcli.NewClient(runner)
}
