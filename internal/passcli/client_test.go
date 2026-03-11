// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package passcli_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/planeopscc/terraform-provider-protonpass/internal/passcli"
	"github.com/planeopscc/terraform-provider-protonpass/internal/testutil"
)

func fixturesDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "testutil", "fixtures")
}

func loadFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(fixturesDir(), name))
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", name, err)
	}
	return data
}

func TestListVaults(t *testing.T) {
	fixture := loadFixture(t, "vault_list.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"vault list": {Stdout: fixture},
	})
	client := passcli.NewClient(runner)
	vaults, err := client.ListVaults(t.Context())
	if err != nil {
		t.Fatalf("ListVaults returned error: %v", err)
	}
	if len(vaults) != 2 {
		t.Fatalf("expected 2 vaults, got %d", len(vaults))
	}
	if vaults[0].ShareID != "share-abc-123" {
		t.Errorf("expected share ID 'share-abc-123', got %q", vaults[0].ShareID)
	}
}

func TestCreateVault(t *testing.T) {
	fixture := loadFixture(t, "vault_list.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"vault create": {Stdout: nil},
		"vault list":   {Stdout: fixture},
	})
	client := passcli.NewClient(runner)
	vault, err := client.CreateVault(t.Context(), "Personal")
	if err != nil {
		t.Fatalf("CreateVault returned error: %v", err)
	}
	if vault.Name != "Personal" {
		t.Errorf("expected vault name 'Personal', got %q", vault.Name)
	}
}

func TestReadVault_Found(t *testing.T) {
	fixture := loadFixture(t, "vault_list.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"vault list": {Stdout: fixture},
	})
	client := passcli.NewClient(runner)
	vault, err := client.ReadVault(t.Context(), "share-def-456")
	if err != nil {
		t.Fatalf("ReadVault returned error: %v", err)
	}
	if vault.Name != "Work" {
		t.Errorf("expected vault name 'Work', got %q", vault.Name)
	}
}

func TestReadVault_NotFound(t *testing.T) {
	fixture := loadFixture(t, "vault_list.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"vault list": {Stdout: fixture},
	})
	client := passcli.NewClient(runner)
	_, err := client.ReadVault(t.Context(), "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing vault")
	}
	if !passcli.IsNotFound(err) {
		t.Errorf("expected not-found error, got: %v", err)
	}
}

func TestHealthCheck_Success(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"test": {Stdout: []byte("Connection successful\n")},
	})
	client := passcli.NewClient(runner)
	if err := client.HealthCheck(t.Context()); err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}
}

func TestHealthCheck_AuthError(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"test": {Err: &passcli.CLIError{ExitCode: 1, Stderr: "unauthorized"}},
	})
	client := passcli.NewClient(runner)
	err := client.HealthCheck(t.Context())
	if err == nil {
		t.Fatal("expected error for unauthenticated session")
	}
	if !passcli.IsAuthError(err) {
		t.Errorf("expected AuthError, got: %v", err)
	}
}

func TestReadItem(t *testing.T) {
	fixture := loadFixture(t, "item_login_read.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item view": {Stdout: fixture},
	})
	client := passcli.NewClient(runner)
	item, err := client.ReadItem(t.Context(), "item-login-001", "share-abc-123")
	if err != nil {
		t.Fatalf("ReadItem returned error: %v", err)
	}
	if item.Title != "Database Credentials" {
		t.Errorf("expected title 'Database Credentials', got %q", item.Title)
	}
}

func TestUpdateItem(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item update": {Stdout: []byte("")},
	})
	client := passcli.NewClient(runner)
	fields := map[string]string{"title": "New Title", "password": "new-secret"}
	err := client.UpdateItem(t.Context(), "item-001", "share-001", fields)
	if err != nil {
		t.Fatalf("UpdateItem returned error: %v", err)
	}
	hasField := false
	for _, a := range runner.Calls[0].Args {
		if a == "--field" {
			hasField = true
			break
		}
	}
	if !hasField {
		t.Errorf("expected --field flag in args: %v", runner.Calls[0].Args)
	}
}

func TestDeleteItem(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item trash": {Stdout: []byte("")},
	})
	client := passcli.NewClient(runner)
	err := client.DeleteItem(t.Context(), "item-001", "share-001", false)
	if err != nil {
		t.Fatalf("DeleteItem returned error: %v", err)
	}
}

func TestCreateItemLogin(t *testing.T) {
	listFixture := loadFixture(t, "item_list_multi.json")
	viewFixture := loadFixture(t, "item_login_read.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item create": {Stdout: []byte("item-login-001")},
		"item list":   {Stdout: listFixture},
		"item view":   {Stdout: viewFixture},
	})
	client := passcli.NewClient(runner)
	item, err := client.CreateItemLogin(t.Context(), "share-abc-123", "Database Credentials", "admin", "secret", "https://db.local", nil)
	if err != nil {
		t.Fatalf("CreateItemLogin returned error: %v", err)
	}
	if item.Title != "Database Credentials" {
		t.Errorf("expected 'Database Credentials', got %q", item.Title)
	}
}

func TestCreateItemNote(t *testing.T) {
	listFixture := loadFixture(t, "item_list_multi.json")
	viewFixture := loadFixture(t, "item_login_read.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item create": {Stdout: []byte("item-note-001")},
		"item list":   {Stdout: listFixture},
		"item view":   {Stdout: viewFixture},
	})
	client := passcli.NewClient(runner)
	item, err := client.CreateItemNote(t.Context(), "share-abc-123", "My Note", "secret")
	if err != nil {
		t.Fatalf("CreateItemNote returned error: %v", err)
	}
	// Note: item_login_read.json has title "Database Credentials"
	if item.Title != "Database Credentials" {
		t.Errorf("expected 'Database Credentials', got %q", item.Title)
	}
}

func TestCreateItemCreditCard(t *testing.T) {
	listFixture := loadFixture(t, "item_list_multi.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item create": {Stdout: []byte("")},
		"item list":   {Stdout: listFixture},
	})
	client := passcli.NewClient(runner)
	item, err := client.CreateItemCreditCard(t.Context(), "share-abc-123", "My Visa", "John", "4111", "123", "2027-12", "5678")
	if err != nil {
		t.Fatalf("CreateItemCreditCard returned error: %v", err)
	}
	if item.Title != "My Visa" {
		t.Errorf("expected 'My Visa', got %q", item.Title)
	}
}

func TestCreateItemWiFi(t *testing.T) {
	listFixture := loadFixture(t, "item_list_multi.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item create": {Stdout: []byte("")},
		"item list":   {Stdout: listFixture},
	})
	client := passcli.NewClient(runner)
	item, err := client.CreateItemWiFi(t.Context(), "share-abc-123", "Office WiFi", "Net", "pw", "wpa2")
	if err != nil {
		t.Fatalf("CreateItemWiFi returned error: %v", err)
	}
	if item.Title != "Office WiFi" {
		t.Errorf("expected 'Office WiFi', got %q", item.Title)
	}
}

func TestCreateItemSSHKey(t *testing.T) {
	listFixture := loadFixture(t, "item_list_multi.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item create": {Stdout: []byte("")},
		"item list":   {Stdout: listFixture},
	})
	client := passcli.NewClient(runner)
	item, err := client.CreateItemSSHKey(t.Context(), "share-abc-123", "Deploy Key", "ed25519", "deploy@server")
	if err != nil {
		t.Fatalf("CreateItemSSHKey returned error: %v", err)
	}
	if item.Title != "Deploy Key" {
		t.Errorf("expected 'Deploy Key', got %q", item.Title)
	}
}

func TestCreateItemIdentity(t *testing.T) {
	listFixture := loadFixture(t, "item_list_multi.json")
	viewFixture := loadFixture(t, "item_login_read.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item create": {Stdout: []byte("item-login-001")},
		"item list":   {Stdout: listFixture},
		"item view":   {Stdout: viewFixture},
	})
	client := passcli.NewClient(runner)
	tmpl := []byte(`{"title":"My Identity","full_name":"John Doe"}`)
	item, err := client.CreateItemIdentity(t.Context(), "share-abc-123", tmpl)
	if err != nil {
		t.Fatalf("CreateItemIdentity returned error: %v", err)
	}
	// Note: item_login_read.json has title "Database Credentials"
	if item.Title != "Database Credentials" {
		t.Errorf("expected 'Database Credentials', got %q", item.Title)
	}
}

func TestListItemsInVault(t *testing.T) {
	fixture := loadFixture(t, "item_list_multi.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item list": {Stdout: fixture},
	})
	client := passcli.NewClient(runner)
	items, err := client.ListItemsInVault(t.Context(), "share-abc-123")
	if err != nil {
		t.Fatalf("ListItemsInVault returned error: %v", err)
	}
	if len(items) != 6 {
		t.Fatalf("expected 6 items, got %d", len(items))
	}
}

// --- Vault Member Tests ---

func TestListVaultMembers(t *testing.T) {
	fixture := loadFixture(t, "vault_member_list.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"vault member": {Stdout: fixture},
	})
	client := passcli.NewClient(runner)
	members, err := client.ListVaultMembers(t.Context(), "share-abc-123")
	if err != nil {
		t.Fatalf("ListVaultMembers returned error: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
	if members[0].Email != "user@example.com" {
		t.Errorf("expected user@example.com, got %q", members[0].Email)
	}
}

func TestAddVaultMember(t *testing.T) {
	fixture := loadFixture(t, "vault_member_list.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"vault share":  {Stdout: []byte("")},
		"vault member": {Stdout: fixture},
	})
	client := passcli.NewClient(runner)
	member, err := client.AddVaultMember(t.Context(), "share-abc-123", "user@example.com", "admin")
	if err != nil {
		t.Fatalf("AddVaultMember returned error: %v", err)
	}
	if member.MemberShareID != "member-001" {
		t.Errorf("expected member-001, got %q", member.MemberShareID)
	}
}

func TestReadVaultMember_Found(t *testing.T) {
	fixture := loadFixture(t, "vault_member_list.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"vault member": {Stdout: fixture},
	})
	client := passcli.NewClient(runner)
	member, err := client.ReadVaultMember(t.Context(), "share-abc-123", "member-002", "")
	if err != nil {
		t.Fatalf("ReadVaultMember returned error: %v", err)
	}
	if member.Role != "readonly" {
		t.Errorf("expected readonly role, got %q", member.Role)
	}
}

func TestRemoveVaultMember(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"vault member": {Stdout: []byte("")},
	})
	client := passcli.NewClient(runner)
	err := client.RemoveVaultMember(t.Context(), "share-abc-123", "member-002")
	if err != nil {
		t.Fatalf("RemoveVaultMember returned error: %v", err)
	}
}

func TestUpdateVaultMemberRole(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"vault member": {Stdout: []byte("")},
	})
	client := passcli.NewClient(runner)
	err := client.UpdateVaultMemberRole(t.Context(), "share-abc-123", "member-001", "readonly")
	if err != nil {
		t.Fatalf("UpdateVaultMemberRole returned error: %v", err)
	}
}

// --- Alias & TOTP Tests ---

func TestCreateAlias(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item alias": {Stdout: []byte(`{"id":"alias-001","alias":"prefix.uuid@slmail.me"}`)},
	})
	client := passcli.NewClient(runner)
	alias, err := client.CreateAlias(t.Context(), "share-abc-123", "prefix")
	if err != nil {
		t.Fatalf("CreateAlias returned error: %v", err)
	}
	if alias.ID != "alias-001" {
		t.Errorf("expected alias-001, got %q", alias.ID)
	}
}

func TestDeleteAlias(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item trash": {Stdout: []byte("")},
	})
	client := passcli.NewClient(runner)
	// DeleteAlias is basically an item trash, but specific to the alias.
	// Since we mapped it to DeleteItem, we test DeleteItem here or a dedicated alias delete method if it existed.
	// Wait, we use `DeleteItem` for aliases in alias_resource.go.
	// Let's just make sure there's no specific DeleteAlias we missed in Client.
	// client.go does not have DeleteAlias, it uses DeleteItem.
	err := client.DeleteItem(t.Context(), "alias-001", "share-abc-123", false)
	if err != nil {
		t.Fatalf("DeleteItem (for alias) returned error: %v", err)
	}
}

func TestGetItemTOTP_Object(t *testing.T) {
	fixture := loadFixture(t, "item_totp.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item totp": {Stdout: fixture},
	})
	client := passcli.NewClient(runner)
	totp, err := client.GetItemTOTP(t.Context(), "item-001", "share-001")
	if err != nil {
		t.Fatalf("GetItemTOTP returned error: %v", err)
	}
	if totp.Code != "123456" {
		t.Errorf("expected 123456, got %q", totp.Code)
	}
}

func TestGetItemTOTP_Array(t *testing.T) {
	fixture := loadFixture(t, "item_totp_array.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item totp": {Stdout: fixture},
	})
	client := passcli.NewClient(runner)
	totp, err := client.GetItemTOTP(t.Context(), "item-001", "share-001")
	if err != nil {
		t.Fatalf("GetItemTOTP returned error: %v", err)
	}
	if totp.Code != "654321" {
		t.Errorf("expected 654321, got %q", totp.Code)
	}
}
