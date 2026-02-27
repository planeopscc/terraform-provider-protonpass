// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package passcli_test

import (
	"context"
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
	vaults, err := client.ListVaults(context.Background())
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
	vault, err := client.CreateVault(context.Background(), "Personal")
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
	vault, err := client.ReadVault(context.Background(), "share-def-456")
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
	_, err := client.ReadVault(context.Background(), "nonexistent")
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
	if err := client.HealthCheck(context.Background()); err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}
}

func TestHealthCheck_AuthError(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"test": {Err: &passcli.CLIError{ExitCode: 1, Stderr: "unauthorized"}},
	})
	client := passcli.NewClient(runner)
	err := client.HealthCheck(context.Background())
	if err == nil {
		t.Fatal("expected error for unauthenticated session")
	}
	if !passcli.IsAuthError(err) {
		t.Errorf("expected AuthError, got: %v", err)
	}
}

func TestReadItemLogin(t *testing.T) {
	fixture := loadFixture(t, "item_login_read.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item view": {Stdout: fixture},
	})
	client := passcli.NewClient(runner)
	item, err := client.ReadItemLogin(context.Background(), "item-login-001", "share-abc-123")
	if err != nil {
		t.Fatalf("ReadItemLogin returned error: %v", err)
	}
	if item.Title != "Database Credentials" {
		t.Errorf("expected title 'Database Credentials', got %q", item.Title)
	}
}

func TestUpdateItemLogin(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item update": {Stdout: []byte("")},
	})
	client := passcli.NewClient(runner)
	fields := map[string]string{"title": "New Title", "password": "new-secret"}
	err := client.UpdateItemLogin(context.Background(), "item-001", "share-001", fields)
	if err != nil {
		t.Fatalf("UpdateItemLogin returned error: %v", err)
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

func TestDeleteItemLogin(t *testing.T) {
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item delete": {Stdout: []byte("")},
	})
	client := passcli.NewClient(runner)
	err := client.DeleteItemLogin(context.Background(), "item-001", "share-001")
	if err != nil {
		t.Fatalf("DeleteItemLogin returned error: %v", err)
	}
}

func TestCreateItemNote(t *testing.T) {
	listFixture := loadFixture(t, "item_list_multi.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item create": {Stdout: []byte("")},
		"item list":   {Stdout: listFixture},
	})
	client := passcli.NewClient(runner)
	item, err := client.CreateItemNote(context.Background(), "share-abc-123", "My Note", "secret")
	if err != nil {
		t.Fatalf("CreateItemNote returned error: %v", err)
	}
	if item.Title != "My Note" {
		t.Errorf("expected 'My Note', got %q", item.Title)
	}
}

func TestCreateItemCreditCard(t *testing.T) {
	listFixture := loadFixture(t, "item_list_multi.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item create": {Stdout: []byte("")},
		"item list":   {Stdout: listFixture},
	})
	client := passcli.NewClient(runner)
	item, err := client.CreateItemCreditCard(context.Background(), "share-abc-123", "My Visa", "John", "4111", "123", "2027-12", "5678")
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
	item, err := client.CreateItemWiFi(context.Background(), "share-abc-123", "Office WiFi", "Net", "pw", "wpa2")
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
	item, err := client.CreateItemSSHKey(context.Background(), "share-abc-123", "Deploy Key", "ed25519", "deploy@server")
	if err != nil {
		t.Fatalf("CreateItemSSHKey returned error: %v", err)
	}
	if item.Title != "Deploy Key" {
		t.Errorf("expected 'Deploy Key', got %q", item.Title)
	}
}

func TestCreateItemIdentity(t *testing.T) {
	listFixture := loadFixture(t, "item_list_multi.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item create": {Stdout: []byte("")},
		"item list":   {Stdout: listFixture},
	})
	client := passcli.NewClient(runner)
	tmpl := []byte(`{"title":"My Identity","full_name":"John Doe"}`)
	item, err := client.CreateItemIdentity(context.Background(), "share-abc-123", tmpl)
	if err != nil {
		t.Fatalf("CreateItemIdentity returned error: %v", err)
	}
	if item.Title != "My Identity" {
		t.Errorf("expected 'My Identity', got %q", item.Title)
	}
}

func TestListItemsInVault(t *testing.T) {
	fixture := loadFixture(t, "item_list_multi.json")
	runner := testutil.NewFakeRunner(map[string]testutil.FakeResponse{
		"item list": {Stdout: fixture},
	})
	client := passcli.NewClient(runner)
	items, err := client.ListItemsInVault(context.Background(), "share-abc-123")
	if err != nil {
		t.Fatalf("ListItemsInVault returned error: %v", err)
	}
	if len(items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(items))
	}
}
