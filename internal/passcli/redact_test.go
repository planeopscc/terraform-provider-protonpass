// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package passcli_test

import (
	"testing"

	"github.com/planeopscc/terraform-provider-protonpass/internal/passcli"
)

func TestRedactArgs_Password(t *testing.T) {
	args := []string{"item", "create", "--password", "secret"}
	redacted := passcli.RedactArgs(args)
	if redacted[3] != "[REDACTED]" {
		t.Errorf("expected '[REDACTED]', got %q", redacted[3])
	}
}

func TestRedactArgs_NoSensitive(t *testing.T) {
	args := []string{"vault", "list"}
	redacted := passcli.RedactArgs(args)
	if len(redacted) != 2 {
		t.Errorf("expected 2 args, got %d", len(redacted))
	}
}
