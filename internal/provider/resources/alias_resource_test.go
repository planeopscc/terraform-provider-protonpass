// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/planeopscc/terraform-provider-protonpass/internal/passcli"
	"github.com/planeopscc/terraform-provider-protonpass/internal/testutil"
)

// buildAliasState constructs a tfsdk.State for alias_resource from plain Go values.
func buildAliasState(t *testing.T, r *AliasResource, id, shareID, prefix, alias string) tfsdk.State {
	t.Helper()
	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)

	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"id":                  tftypes.String,
			"share_id":            tftypes.String,
			"prefix":              tftypes.String,
			"alias":               tftypes.String,
			"destroy_permanently": tftypes.Bool,
		},
	}
	raw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"id":                  tftypes.NewValue(tftypes.String, id),
		"share_id":            tftypes.NewValue(tftypes.String, shareID),
		"prefix":              tftypes.NewValue(tftypes.String, prefix),
		"alias":               tftypes.NewValue(tftypes.String, alias),
		"destroy_permanently": tftypes.NewValue(tftypes.Bool, nil), // optional, unset
	})
	return tfsdk.State{Schema: schemaResp.Schema, Raw: raw}
}

// TestAliasRead_NotFound_RemovesFromState verifies that a not-found error removes
// the resource from state without adding diagnostics.
func TestAliasRead_NotFound_RemovesFromState(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item view": {Err: &passcli.CLIError{ExitCode: 1, Stderr: "Could not find item"}},
	})
	r := &AliasResource{client: passcli.NewClient(runner)}
	state := buildAliasState(t, r, "alias-001", "share-001", "mypfx", "mypfx@passmail.net")

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error on not-found, got: %v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Error("expected state to be removed (null) after not-found")
	}
}

// TestAliasRead_TransientError_AddsWarning verifies that a non-not-found error
// produces a warning diagnostic and retains the existing state.
func TestAliasRead_TransientError_AddsWarning(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item view": {Err: &passcli.CLIError{ExitCode: 422, Stderr: "rate limited"}},
	})
	r := &AliasResource{client: passcli.NewClient(runner)}
	state := buildAliasState(t, r, "alias-001", "share-001", "mypfx", "mypfx@passmail.net")

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error (warning only), got error: %v", resp.Diagnostics)
	}
	hasWarning := false
	for _, d := range resp.Diagnostics {
		if d.Severity() == diag.SeverityWarning {
			hasWarning = true
		}
	}
	if !hasWarning {
		t.Error("expected a warning diagnostic for transient read error, got none")
	}
}
