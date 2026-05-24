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

func TestRedactArgs_PasswordConcatenated(t *testing.T) {
	args := []string{"item", "create", "--password=supersecret"}
	redacted := passcli.RedactArgs(args)
	if redacted[2] != "--password=[REDACTED]" {
		t.Errorf("expected '--password=[REDACTED]', got %q", redacted[2])
	}
}

func TestRedactArgs_NoteConcatenated(t *testing.T) {
	args := []string{"item", "create", "--note=top secret content"}
	redacted := passcli.RedactArgs(args)
	if redacted[2] != "--note=[REDACTED]" {
		t.Errorf("expected '--note=[REDACTED]', got %q", redacted[2])
	}
}

func TestRedactArgs_PinConcatenated(t *testing.T) {
	args := []string{"item", "create", "--pin=1234"}
	redacted := passcli.RedactArgs(args)
	if redacted[2] != "--pin=[REDACTED]" {
		t.Errorf("expected '--pin=[REDACTED]', got %q", redacted[2])
	}
}

func TestRedactArgs_NumberConcatenated(t *testing.T) {
	args := []string{"item", "create", "--number=4111222233334444"}
	redacted := passcli.RedactArgs(args)
	if redacted[2] != "--number=[REDACTED]" {
		t.Errorf("expected '--number=[REDACTED]', got %q", redacted[2])
	}
}

func TestRedactArgs_FieldPassword(t *testing.T) {
	args := []string{"item", "edit", "--field", "password=newsecret"}
	redacted := passcli.RedactArgs(args)
	if redacted[3] != "password=[REDACTED]" {
		t.Errorf("expected 'password=[REDACTED]', got %q", redacted[3])
	}
}

func TestRedactArgs_FieldNote(t *testing.T) {
	args := []string{"item", "edit", "--field", "note=sensitive content"}
	redacted := passcli.RedactArgs(args)
	if redacted[3] != "note=[REDACTED]" {
		t.Errorf("expected 'note=[REDACTED]', got %q", redacted[3])
	}
}

func TestRedactArgs_FieldSSN(t *testing.T) {
	args := []string{"item", "edit", "--field", "social_security_number=123-45-6789"}
	redacted := passcli.RedactArgs(args)
	if redacted[3] != "social_security_number=[REDACTED]" {
		t.Errorf("expected 'social_security_number=[REDACTED]', got %q", redacted[3])
	}
}

func TestRedactArgs_FieldPassportNumber(t *testing.T) {
	args := []string{"item", "edit", "--field", "passport_number=AB123456"}
	redacted := passcli.RedactArgs(args)
	if redacted[3] != "passport_number=[REDACTED]" {
		t.Errorf("expected 'passport_number=[REDACTED]', got %q", redacted[3])
	}
}

func TestRedactArgs_FieldLicenseNumber(t *testing.T) {
	args := []string{"item", "edit", "--field", "license_number=DL-987654"}
	redacted := passcli.RedactArgs(args)
	if redacted[3] != "license_number=[REDACTED]" {
		t.Errorf("expected 'license_number=[REDACTED]', got %q", redacted[3])
	}
}

func TestRedactArgs_FieldVerificationNumber(t *testing.T) {
	args := []string{"item", "edit", "--field", "verification_number=123"}
	redacted := passcli.RedactArgs(args)
	if redacted[3] != "verification_number=[REDACTED]" {
		t.Errorf("expected 'verification_number=[REDACTED]', got %q", redacted[3])
	}
}

func TestRedactArgs_FieldNonSensitive(t *testing.T) {
	args := []string{"item", "edit", "--field", "title=My Login"}
	redacted := passcli.RedactArgs(args)
	if redacted[3] != "title=My Login" {
		t.Errorf("expected 'title=My Login', got %q", redacted[3])
	}
}

func TestRedactArgs_MultipleFields(t *testing.T) {
	args := []string{
		"item", "edit",
		"--field", "title=My Login",
		"--field", "password=secret123",
		"--field", "username=user",
	}
	redacted := passcli.RedactArgs(args)
	if redacted[3] != "title=My Login" {
		t.Errorf("expected 'title=My Login', got %q", redacted[3])
	}
	if redacted[5] != "password=[REDACTED]" {
		t.Errorf("expected 'password=[REDACTED]', got %q", redacted[5])
	}
	if redacted[7] != "username=user" {
		t.Errorf("expected 'username=user', got %q", redacted[7])
	}
}

func TestRedactArgs_PreservesStructuralArgs(t *testing.T) {
	args := []string{"item", "create", "--share-id=abc123", "--type=login", "--title=My Item"}
	redacted := passcli.RedactArgs(args)
	for i, arg := range args {
		if redacted[i] != arg {
			t.Errorf("index %d: expected %q, got %q", i, arg, redacted[i])
		}
	}
}
