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

func buildVaultMemberState(t *testing.T, r *VaultMemberResource, shareID, email, role, memberShareID string) tfsdk.State {
	t.Helper()
	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)

	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"share_id":        tftypes.String,
			"email":           tftypes.String,
			"role":            tftypes.String,
			"member_share_id": tftypes.String,
		},
	}
	raw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"share_id":        tftypes.NewValue(tftypes.String, shareID),
		"email":           tftypes.NewValue(tftypes.String, email),
		"role":            tftypes.NewValue(tftypes.String, role),
		"member_share_id": tftypes.NewValue(tftypes.String, memberShareID),
	})
	return tfsdk.State{Schema: schemaResp.Schema, Raw: raw}
}

// TestVaultMemberRead_NotFound_RemovesFromState verifies that a member no
// longer present in the vault's member list is removed from state silently.
func TestVaultMemberRead_NotFound_RemovesFromState(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"vault member": {Stdout: []byte(`[]`)},
	})
	r := &VaultMemberResource{client: passcli.NewClient(runner)}
	state := buildVaultMemberState(t, r, "share-abc-123", "user@example.com", "viewer", "member-gone")

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

// TestVaultMemberDelete_NotFound_IsIdempotent verifies that deleting a member
// that is already gone does not produce an error diagnostic.
func TestVaultMemberDelete_NotFound_IsIdempotent(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"vault member": {Err: &passcli.CLIError{ExitCode: 1, Stderr: "member not found"}},
	})
	r := &VaultMemberResource{client: passcli.NewClient(runner)}
	state := buildVaultMemberState(t, r, "share-abc-123", "user@example.com", "viewer", "member-001")

	req := resource.DeleteRequest{State: state}
	resp := &resource.DeleteResponse{}
	r.Delete(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no error for already-deleted member, got: %v", resp.Diagnostics)
	}
}

// TestVaultMemberUpdate_RoleChange verifies that changing the role triggers
// an UpdateVaultMemberRole CLI call.
func TestVaultMemberUpdate_RoleChange(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"vault member": {Stdout: nil},
	})
	r := &VaultMemberResource{client: passcli.NewClient(runner)}

	schemaResp := &resource.SchemaResponse{}
	r.Schema(context.Background(), resource.SchemaRequest{}, schemaResp)

	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"share_id":        tftypes.String,
			"email":           tftypes.String,
			"role":            tftypes.String,
			"member_share_id": tftypes.String,
		},
	}
	rawState := tftypes.NewValue(objType, map[string]tftypes.Value{
		"share_id":        tftypes.NewValue(tftypes.String, "share-abc-123"),
		"email":           tftypes.NewValue(tftypes.String, "user@example.com"),
		"role":            tftypes.NewValue(tftypes.String, "viewer"),
		"member_share_id": tftypes.NewValue(tftypes.String, "member-001"),
	})
	rawPlan := tftypes.NewValue(objType, map[string]tftypes.Value{
		"share_id":        tftypes.NewValue(tftypes.String, "share-abc-123"),
		"email":           tftypes.NewValue(tftypes.String, "user@example.com"),
		"role":            tftypes.NewValue(tftypes.String, "editor"), // changed
		"member_share_id": tftypes.NewValue(tftypes.String, "member-001"),
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

	found := false
	for _, call := range runner.Calls {
		if len(call.Args) >= 2 && call.Args[0] == "vault" && call.Args[1] == "member" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected vault member update CLI call when role changes, but none was made")
	}
}
