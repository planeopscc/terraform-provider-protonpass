// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package testutil

import (
	"context"
	"fmt"
	"strings"

	"github.com/planeopscc/terraform-provider-protonpass/internal/passcli"
)

// FakeResponse defines what a fake runner returns for a command.
type FakeResponse struct {
	Stdout []byte
	Err    error
}

// FakeCall records a CLI invocation.
type FakeCall struct {
	Args []string
}

// FakeRunner is an in-memory Runner for testing.
type FakeRunner struct {
	responses map[string]FakeResponse
	Calls     []FakeCall
}

// NewFakeRunner creates a FakeRunner with command->response mappings.
// Keys are matched by prefix, e.g. "vault list" matches args ["vault", "list", "--output", "json"].
func NewFakeRunner(responses map[string]FakeResponse) *FakeRunner {
	return &FakeRunner{responses: responses}
}

func (f *FakeRunner) Run(ctx context.Context, args ...string) ([]byte, []byte, error) {
	f.Calls = append(f.Calls, FakeCall{Args: args})
	key := strings.Join(args[:min(len(args), 2)], " ")
	if resp, ok := f.responses[key]; ok {
		return resp.Stdout, nil, resp.Err
	}
	return nil, nil, &passcli.CLIError{
		Args:     args,
		ExitCode: 1,
		Stderr:   fmt.Sprintf("no fake response for %q", key),
	}
}
