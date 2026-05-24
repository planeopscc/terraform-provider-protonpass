// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package passcli_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/planeopscc/terraform-provider-protonpass/internal/passcli"
)

func TestNewExecRunnerWithEnv_InjectsVar(t *testing.T) {
	runner := passcli.NewExecRunnerWithEnv("printenv", 5*time.Second,
		[]string{"PROTON_PASS_AGENT_REASON=audit-test"})
	stdout, _, err := runner.Run(context.Background(), "PROTON_PASS_AGENT_REASON")
	if err != nil {
		t.Fatalf("printenv failed: %v", err)
	}
	got := strings.TrimSpace(string(stdout))
	if got != "audit-test" {
		t.Errorf("expected PROTON_PASS_AGENT_REASON=audit-test, got %q", got)
	}
}

func TestNewExecRunnerWithEnv_ConfigOverridesProcessEnv(t *testing.T) {
	t.Setenv("PROTON_PASS_AGENT_REASON", "from-env")
	runner := passcli.NewExecRunnerWithEnv("printenv", 5*time.Second,
		[]string{"PROTON_PASS_AGENT_REASON=from-config"})
	stdout, _, err := runner.Run(context.Background(), "PROTON_PASS_AGENT_REASON")
	if err != nil {
		t.Fatalf("printenv failed: %v", err)
	}
	got := strings.TrimSpace(string(stdout))
	if got != "from-config" {
		t.Errorf("expected provider config to win, got %q", got)
	}
}

func TestNewExecRunner_InheritsProcessEnv(t *testing.T) {
	t.Setenv("PROTON_PASS_AGENT_REASON", "inherited")
	runner := passcli.NewExecRunner("printenv", 5*time.Second)
	stdout, _, err := runner.Run(context.Background(), "PROTON_PASS_AGENT_REASON")
	if err != nil {
		t.Fatalf("printenv failed: %v", err)
	}
	got := strings.TrimSpace(string(stdout))
	if got != "inherited" {
		t.Errorf("expected inherited value, got %q", got)
	}
}
