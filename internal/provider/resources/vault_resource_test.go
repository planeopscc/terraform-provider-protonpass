// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/planeopscc/terraform-provider-protonpass/internal/passcli"
	"github.com/planeopscc/terraform-provider-protonpass/internal/testutil"
)

func buildVaultState(t *testing.T, r *VaultResource, shareID, vaultID, name string) tfsdk.State {
	t.Helper()
	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)

	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"share_id": tftypes.String,
			"vault_id": tftypes.String,
			"name":     tftypes.String,
		},
	}
	raw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"share_id": tftypes.NewValue(tftypes.String, shareID),
		"vault_id": tftypes.NewValue(tftypes.String, vaultID),
		"name":     tftypes.NewValue(tftypes.String, name),
	})
	return tfsdk.State{Schema: schemaResp.Schema, Raw: raw}
}

// TestVaultRead_NotFound_RemovesFromState verifies that a vault no longer present
// in the remote API is silently removed from state without an error diagnostic.
func TestVaultRead_NotFound_RemovesFromState(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"vault list": {Stdout: []byte(`{"vaults": []}`)},
	})
	r := &VaultResource{client: passcli.NewClient(runner)}
	state := buildVaultState(t, r, "share-gone", "vault-gone", "Gone")

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

// TestVaultRead_Error_AddsDiagnostic verifies that a CLI error during Read
// surfaces as an error diagnostic and does not silently drop state.
func TestVaultRead_Error_AddsDiagnostic(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"vault list": {Err: &passcli.CLIError{ExitCode: 1, Stderr: "internal server error"}},
	})
	r := &VaultResource{client: passcli.NewClient(runner)}
	state := buildVaultState(t, r, "share-abc-123", "vault-abc-123", "Personal")

	req := resource.ReadRequest{State: state}
	resp := &resource.ReadResponse{State: state}
	r.Read(context.Background(), req, resp)

	if !resp.Diagnostics.HasError() {
		t.Error("expected an error diagnostic on CLI failure, got none")
	}
}

// TestVaultUpdate_SameNameSkipsAPICall verifies that renaming a vault to its
// current name does not trigger a vault update CLI call.
func TestVaultUpdate_SameNameSkipsAPICall(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{})
	r := &VaultResource{client: passcli.NewClient(runner)}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)

	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"share_id": tftypes.String,
			"vault_id": tftypes.String,
			"name":     tftypes.String,
		},
	}
	rawState := tftypes.NewValue(objType, map[string]tftypes.Value{
		"share_id": tftypes.NewValue(tftypes.String, "share-abc-123"),
		"vault_id": tftypes.NewValue(tftypes.String, "vault-abc-123"),
		"name":     tftypes.NewValue(tftypes.String, "Personal"),
	})
	rawPlan := tftypes.NewValue(objType, map[string]tftypes.Value{
		"share_id": tftypes.NewValue(tftypes.String, "share-abc-123"),
		"vault_id": tftypes.NewValue(tftypes.String, "vault-abc-123"),
		"name":     tftypes.NewValue(tftypes.String, "Personal"), // unchanged
	})

	req := resource.UpdateRequest{
		Plan:  tfsdk.Plan{Schema: schemaResp.Schema, Raw: rawPlan},
		State: tfsdk.State{Schema: schemaResp.Schema, Raw: rawState},
	}
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: schemaResp.Schema, Raw: rawState}}
	r.Update(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("unexpected error: %v", resp.Diagnostics)
	}
	for _, call := range runner.Calls {
		if len(call.Args) >= 2 && call.Args[0] == "vault" && call.Args[1] == "update" {
			t.Error("expected no vault update CLI call when name is unchanged, but one was made")
		}
	}
}
