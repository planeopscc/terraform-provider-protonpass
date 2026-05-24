// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package passcli_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/planeopscc/terraform-provider-protonpass/internal/testutil"
)

// TestAcc_VaultLifecycle creates, reads, and deletes a vault against a live
// pass-cli session. It requires TF_ACC=1 and an authenticated pass-cli session.
func TestAcc_VaultLifecycle(t *testing.T) {
	testutil.SkipIfNotAcc(t)
	client := testutil.NewAccClient(t)
	ctx := t.Context()

	if err := client.HealthCheck(ctx); err != nil {
		t.Skipf("pass-cli session not active: %v — run 'pass-cli login' first", err)
	}

	name := fmt.Sprintf("tf-acc-vault-%d", time.Now().UnixMilli())

	vault, err := client.CreateVault(ctx, name)
	if err != nil {
		t.Fatalf("CreateVault(%q): %v", name, err)
	}
	t.Cleanup(func() {
		if err := client.DeleteVault(context.Background(), vault.ShareID); err != nil {
			t.Logf("cleanup: DeleteVault(%q) failed: %v", vault.ShareID, err)
		}
	})

	if vault.Name != name {
		t.Errorf("vault name: expected %q, got %q", name, vault.Name)
	}
	if vault.ShareID == "" {
		t.Error("vault.ShareID must not be empty after creation")
	}

	found, err := client.ReadVault(ctx, vault.ShareID)
	if err != nil {
		t.Fatalf("ReadVault(%q): %v", vault.ShareID, err)
	}
	if found.Name != name {
		t.Errorf("ReadVault name: expected %q, got %q", name, found.Name)
	}
}
