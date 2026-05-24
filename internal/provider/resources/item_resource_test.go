// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/planeopscc/terraform-provider-protonpass/internal/passcli"
)

// TestMapItemToModel_ImportNullType verifies that mapItemToModel sets data.Type
// from item.Type when the model's Type is null — which is the state after ImportState,
// where only share_id and item_id are populated.
func TestMapItemToModel_ImportNullType(t *testing.T) {
	r := &ItemResource{}

	tests := []struct {
		name        string
		item        passcli.ItemJSON
		wantType    string
		checkFields func(t *testing.T, data ItemResourceModel)
	}{
		{
			name: "login",
			item: passcli.ItemJSON{
				ItemID:   "item-001",
				ShareID:  "share-abc",
				Type:     "login",
				Title:    "My Login",
				Username: "alice",
				Password: "s3cret",
				URLs:     []string{"https://example.com"},
			},
			wantType: "login",
			checkFields: func(t *testing.T, data ItemResourceModel) {
				t.Helper()
				if data.Username.ValueString() != "alice" {
					t.Errorf("username: expected 'alice', got %q", data.Username.ValueString())
				}
				if data.Password.ValueString() != "s3cret" {
					t.Errorf("password: expected 's3cret', got %q", data.Password.ValueString())
				}
			},
		},
		{
			name: "note",
			item: passcli.ItemJSON{
				ItemID:  "item-002",
				ShareID: "share-abc",
				Type:    "note",
				Title:   "My Note",
				Note:    "secret content",
			},
			wantType: "note",
			checkFields: func(t *testing.T, data ItemResourceModel) {
				t.Helper()
				if data.Note.ValueString() != "secret content" {
					t.Errorf("note: expected 'secret content', got %q", data.Note.ValueString())
				}
			},
		},
		{
			name: "credit-card",
			item: passcli.ItemJSON{
				ItemID:             "item-003",
				ShareID:            "share-abc",
				Type:               "credit-card",
				Title:              "My Card",
				CardholderName:     "Jane Doe",
				Number:             "4111222233334444",
				VerificationNumber: "123",
				ExpirationDate:     "2030-12",
				PIN:                "9999",
			},
			wantType: "credit-card",
			checkFields: func(t *testing.T, data ItemResourceModel) {
				t.Helper()
				if data.CardholderName.ValueString() != "Jane Doe" {
					t.Errorf("cardholder_name: expected 'Jane Doe', got %q", data.CardholderName.ValueString())
				}
				if data.Number.ValueString() != "4111222233334444" {
					t.Errorf("number: expected '4111222233334444', got %q", data.Number.ValueString())
				}
				if data.VerificationNumber.ValueString() != "123" {
					t.Errorf("verification_number: expected '123', got %q", data.VerificationNumber.ValueString())
				}
				if data.ExpirationDate.ValueString() != "2030-12" {
					t.Errorf("expiration_date: expected '2030-12', got %q", data.ExpirationDate.ValueString())
				}
				if data.PIN.ValueString() != "9999" {
					t.Errorf("pin: expected '9999', got %q", data.PIN.ValueString())
				}
			},
		},
		{
			name: "wifi",
			item: passcli.ItemJSON{
				ItemID:   "item-004",
				ShareID:  "share-abc",
				Type:     "wifi",
				Title:    "Office WiFi",
				SSID:     "Corp-Net",
				Password: "wifi-secret",
				Security: "WPA2",
			},
			wantType: "wifi",
			checkFields: func(t *testing.T, data ItemResourceModel) {
				t.Helper()
				if data.SSID.ValueString() != "Corp-Net" {
					t.Errorf("ssid: expected 'Corp-Net', got %q", data.SSID.ValueString())
				}
				if data.Security.ValueString() != "WPA2" {
					t.Errorf("security: expected 'WPA2', got %q", data.Security.ValueString())
				}
				if data.Password.ValueString() != "wifi-secret" {
					t.Errorf("password: expected 'wifi-secret', got %q", data.Password.ValueString())
				}
			},
		},
		{
			name: "ssh-key",
			item: passcli.ItemJSON{
				ItemID:     "item-005",
				ShareID:    "share-abc",
				Type:       "ssh-key",
				Title:      "Deploy Key",
				PrivateKey: "-----BEGIN OPENSSH PRIVATE KEY-----",
				PublicKey:  "ssh-ed25519 AAAA...",
			},
			wantType: "ssh-key",
			checkFields: func(t *testing.T, data ItemResourceModel) {
				t.Helper()
				if data.PrivateKey.ValueString() != "-----BEGIN OPENSSH PRIVATE KEY-----" {
					t.Errorf("private_key: unexpected value %q", data.PrivateKey.ValueString())
				}
				if data.PublicKey.ValueString() != "ssh-ed25519 AAAA..." {
					t.Errorf("public_key: unexpected value %q", data.PublicKey.ValueString())
				}
			},
		},
		{
			name: "identity",
			item: passcli.ItemJSON{
				ItemID:    "item-006",
				ShareID:   "share-abc",
				Type:      "identity",
				Title:     "Jane Doe",
				FullName:  "Jane Doe",
				FirstName: "Jane",
				LastName:  "Doe",
				Email:     "jane@example.com",
			},
			wantType: "identity",
			checkFields: func(t *testing.T, data ItemResourceModel) {
				t.Helper()
				if data.FullName.ValueString() != "Jane Doe" {
					t.Errorf("full_name: expected 'Jane Doe', got %q", data.FullName.ValueString())
				}
				if data.FirstName.ValueString() != "Jane" {
					t.Errorf("first_name: expected 'Jane', got %q", data.FirstName.ValueString())
				}
				if data.Email.ValueString() != "jane@example.com" {
					t.Errorf("email: expected 'jane@example.com', got %q", data.Email.ValueString())
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := ItemResourceModel{
				Type: types.StringNull(), // simulates state after ImportState
			}
			diags := r.mapItemToModel(context.Background(), &tt.item, &data)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %v", diags)
			}
			if data.Type.ValueString() != tt.wantType {
				t.Errorf("type: expected %q, got %q", tt.wantType, data.Type.ValueString())
			}
			tt.checkFields(t, data)
		})
	}
}
