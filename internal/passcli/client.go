// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package passcli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Client wraps a Runner to provide high-level Proton Pass operations.
type Client struct {
	runner Runner
}

// NewClient creates a new Client.
func NewClient(runner Runner) *Client {
	return &Client{runner: runner}
}

// HealthCheck verifies that pass-cli is available and a session is active.
func (c *Client) HealthCheck(ctx context.Context) error {
	stdout, _, err := c.runner.Run(ctx, "test")
	if err != nil {
		return err
	}
	if !strings.Contains(string(stdout), "Connection successful") {
		return fmt.Errorf("unexpected output from pass-cli test: %s", string(stdout))
	}
	return nil
}

// --- Vault operations ---

// ListVaults returns all vaults.
func (c *Client) ListVaults(ctx context.Context) ([]VaultJSON, error) {
	stdout, _, err := c.runner.Run(ctx, "vault", "list", "--output=json")
	if err != nil {
		return nil, err
	}
	var resp struct {
		Vaults []VaultJSON `json:"vaults"`
	}
	if err := json.Unmarshal(stdout, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse vault list JSON: %w", err)
	}
	return resp.Vaults, nil
}

// CreateVault creates a new vault and returns its details.
func (c *Client) CreateVault(ctx context.Context, name string) (*VaultJSON, error) {
	_, _, err := c.runner.Run(ctx, "vault", "create", "--name="+name)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault: %w", err)
	}
	vaults, err := c.ListVaults(ctx)
	if err != nil {
		return nil, err
	}
	for _, v := range vaults {
		if v.Name == name {
			return &v, nil
		}
	}
	return nil, &CLIError{Stderr: fmt.Sprintf("vault %q not found after creation", name)}
}

// ReadVault reads a vault by share ID.
func (c *Client) ReadVault(ctx context.Context, shareID string) (*VaultJSON, error) {
	vaults, err := c.ListVaults(ctx)
	if err != nil {
		return nil, err
	}
	for _, v := range vaults {
		if v.ShareID == shareID {
			return &v, nil
		}
	}
	return nil, &CLIError{Stderr: fmt.Sprintf("vault with share_id %q not found", shareID)}
}

// UpdateVault renames a vault.
func (c *Client) UpdateVault(ctx context.Context, shareID, newName string) error {
	_, _, err := c.runner.Run(ctx, "vault", "update", "--share-id="+shareID, "--name="+newName)
	if err != nil {
		return fmt.Errorf("failed to rename vault: %w", err)
	}
	return nil
}

// DeleteVault deletes a vault.
func (c *Client) DeleteVault(ctx context.Context, shareID string) error {
	_, _, err := c.runner.Run(ctx, "vault", "delete", "--share-id="+shareID)
	if err != nil {
		return fmt.Errorf("failed to delete vault: %w", err)
	}
	return nil
}

// --- Vault Member operations ---

// AddVaultMember invites a user to a vault.
func (c *Client) AddVaultMember(ctx context.Context, shareID, email, role string) (*VaultMemberJSON, error) {
	_, _, err := c.runner.Run(ctx, "vault", "share", "--share-id="+shareID, "--role="+role, email)
	if err != nil {
		if strings.Contains(err.Error(), "InvalidValue") || strings.Contains(err.Error(), "NotAllowed") {
			tflog.Warn(ctx, "API returned error during vault share, but invite was likely sent", map[string]interface{}{"error": err.Error()})
		} else {
			return nil, fmt.Errorf("failed to add member %q to vault %q: %w", email, shareID, err)
		}
	}
	// Since `vault share` may not output standard JSON directly or we need the full updated list,
	// let's fetch the member list and filter.
	members, err := c.ListVaultMembers(ctx, shareID)
	if err == nil {
		for _, m := range members {
			if strings.EqualFold(m.Email, email) {
				return &m, nil
			}
		}
	}

	// Workaround: newly invited users who haven't accepted don't appear in the list.
	// We return a synthetic ID so Terraform can track the resource state.
	return &VaultMemberJSON{
		MemberShareID: "pending-" + email,
		Email:         email,
		Role:          role,
	}, nil
}

// ListVaultMembers gets all members of a vault.
func (c *Client) ListVaultMembers(ctx context.Context, shareID string) ([]VaultMemberJSON, error) {
	stdout, _, err := c.runner.Run(ctx, "vault", "member", "list", "--share-id="+shareID, "--output=json")
	if err != nil {
		return nil, fmt.Errorf("failed to list members for vault %q: %w", shareID, err)
	}

	var members []VaultMemberJSON
	if err := json.Unmarshal(stdout, &members); err != nil {
		return nil, fmt.Errorf("failed to parse vault members JSON: %w", err)
	}

	return members, nil
}

// ReadVaultMember finds a specific member by their share ID or email.
func (c *Client) ReadVaultMember(ctx context.Context, shareID, memberShareID, email string) (*VaultMemberJSON, error) {
	members, err := c.ListVaultMembers(ctx, shareID)
	if err != nil {
		return nil, err
	}

	for _, m := range members {
		if m.MemberShareID == memberShareID || (email != "" && strings.EqualFold(m.Email, email)) {
			return &m, nil
		}
	}

	if strings.HasPrefix(memberShareID, "pending-") {
		// Retain the synthetic state if the user is still pending
		return &VaultMemberJSON{
			MemberShareID: memberShareID,
			Email:         strings.TrimPrefix(memberShareID, "pending-"),
			Role:          "unknown", // We might not know the exact role, but keeping state is paramount
		}, nil
	}

	return nil, fmt.Errorf("member %q not found in vault %q", memberShareID, shareID)
}

// RemoveVaultMember removes a user from a vault.
func (c *Client) RemoveVaultMember(ctx context.Context, shareID, memberShareID string) error {
	if strings.HasPrefix(memberShareID, "pending-") {
		// We cannot delete a pending invite via the CLI currently
		tflog.Warn(ctx, "Skipping deletion of pending vault invite", map[string]interface{}{
			"member_share_id": memberShareID,
		})
		return nil
	}

	_, _, err := c.runner.Run(ctx, "vault", "member", "remove", "--share-id="+shareID, "--member-share-id="+memberShareID)
	if err != nil {
		return fmt.Errorf("failed to remove vault member %q: %w", memberShareID, err)
	}
	return nil
}

// UpdateVaultMemberRole updates a user's role on a vault.
func (c *Client) UpdateVaultMemberRole(ctx context.Context, shareID, memberShareID, role string) error {
	if strings.HasPrefix(memberShareID, "pending-") {
		tflog.Warn(ctx, "Skipping role update for pending vault invite", map[string]interface{}{
			"member_share_id": memberShareID,
		})
		return nil
	}

	_, _, err := c.runner.Run(ctx, "vault", "member", "update", "--share-id="+shareID, "--member-share-id="+memberShareID, "--role="+role)
	if err != nil {
		return fmt.Errorf("failed to update vault member %q role to %q: %w", memberShareID, role, err)
	}
	return nil
}

// GetItemTOTP retrieves the current TOTP code for an item.
func (c *Client) GetItemTOTP(ctx context.Context, itemID, shareID string) (*ItemTOTPJSON, error) {
	args := []string{"item", "totp",
		"--item-id=" + itemID,
		"--share-id=" + shareID,
		"--output=json",
	}

	stdout, _, err := c.runner.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get TOTP for item %q: %w", itemID, err)
	}

	// pass-cli item totp might return an array of codes if there are multiple TOTP fields,
	// or a single object. We will try unmarshaling as an array first.
	var totpArray []ItemTOTPJSON
	if err := json.Unmarshal(stdout, &totpArray); err == nil && len(totpArray) > 0 {
		return &totpArray[0], nil
	}

	// Fallback to single object
	var totpResp ItemTOTPJSON
	if err := json.Unmarshal(stdout, &totpResp); err != nil {
		return nil, fmt.Errorf("failed to parse TOTP response JSON: %w", err)
	}

	return &totpResp, nil
}

// CreateAlias creates a new Hide-My-Email alias.
func (c *Client) CreateAlias(ctx context.Context, shareID, prefix string) (*AliasCreateJSON, error) {
	args := []string{"item", "alias", "create",
		"--share-id=" + shareID,
		"--prefix=" + prefix,
		"--output=json",
	}

	stdout, _, err := c.runner.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create alias %q: %w", prefix, err)
	}

	var aliasResp AliasCreateJSON
	if err := json.Unmarshal(stdout, &aliasResp); err != nil {
		return nil, fmt.Errorf("failed to parse alias response JSON: %w", err)
	}

	return &aliasResp, nil
}

// --- Item operations ---

// ListItemsInVault lists all items in a vault.
func (c *Client) ListItemsInVault(ctx context.Context, shareID string) ([]ItemJSON, error) {
	stdout, _, err := c.runner.Run(ctx, "item", "list", "--share-id="+shareID, "--output=json")
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}
	var resp struct {
		Items []ItemRawJSON `json:"items"`
	}
	if err := json.Unmarshal(stdout, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse items JSON: %w", err)
	}
	result := make([]ItemJSON, len(resp.Items))
	for i, raw := range resp.Items {
		result[i] = FlattenItem(raw)
	}
	return result, nil
}

// ListTrashedItems lists all trashed items in a vault.
func (c *Client) ListTrashedItems(ctx context.Context, shareID string) ([]ItemJSON, error) {
	stdout, _, err := c.runner.Run(ctx, "item", "list", "--share-id="+shareID, "--filter-state=trashed", "--output=json")
	if err != nil {
		return nil, fmt.Errorf("failed to list trashed items: %w", err)
	}
	var resp struct {
		Items []ItemRawJSON `json:"items"`
	}
	if err := json.Unmarshal(stdout, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse trashed items JSON: %w", err)
	}
	result := make([]ItemJSON, len(resp.Items))
	for i, raw := range resp.Items {
		result[i] = FlattenItem(raw)
	}
	return result, nil
}

// RestoreItem restores an item from the trash.
func (c *Client) RestoreItem(ctx context.Context, itemID, shareID string) error {
	_, _, err := c.runner.Run(ctx, "item", "untrash", "--item-id="+itemID, "--share-id="+shareID)
	if err != nil {
		return fmt.Errorf("failed to untrash item %q: %w", itemID, err)
	}
	return nil
}

// GetItem returns the full raw JSON for a single item.
// It can use either itemID or title (title is less reliable if duplicates exist, but useful right after creation).
func (c *Client) GetItem(ctx context.Context, itemID, title, shareID string) (*ItemRawJSON, error) {
	var args []string
	if itemID != "" {
		uri := fmt.Sprintf("pass://%s/%s", shareID, itemID)
		args = []string{"item", "view", "--output=json", "--", uri}
	} else if title != "" {
		args = []string{"item", "view", "--output=json", "--share-id=" + shareID, "--item-title=" + title}
	} else {
		return nil, fmt.Errorf("must provide either itemID or title")
	}

	stdout, _, err := c.runner.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	var resp ItemViewResponse
	if err := json.Unmarshal(stdout, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse item view JSON: %w (output: %q)", err, string(stdout))
	}
	return &resp.Item, nil
}

// findItemByTitle finds a newly created item in the vault by title.
func (c *Client) findItemByTitle(ctx context.Context, shareID, title string) (*ItemJSON, error) {
	items, err := c.ListItemsInVault(ctx, shareID)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		if item.Title == title {
			return &item, nil
		}
	}
	return nil, &CLIError{Stderr: fmt.Sprintf("item %q not found in vault %q after creation", title, shareID)}
}

// WriteTempFile writes data to a temporary file and returns the path.
// The caller is responsible for removing the file when done.
func WriteTempFile(pattern string, data []byte) (string, error) {
	tmpFile, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(data); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	return tmpFile.Name(), nil
}

// CreateItemLogin creates a new login item.
func (c *Client) CreateItemLogin(ctx context.Context, shareID, title, username, password, email string, urls []string) (*ItemJSON, error) {
	args := []string{"item", "create", "login",
		"--share-id=" + shareID,
		"--title=" + title,
	}
	if username != "" {
		args = append(args, "--username="+username)
	}
	if password != "" {
		args = append(args, "--password="+password)
	}
	if email != "" {
		args = append(args, "--email="+email)
	}
	for _, u := range urls {
		args = append(args, "--url="+u)
	}
	_, _, err := c.runner.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create login item: %w", err)
	}
	return c.findItemByTitle(ctx, shareID, title)
}

// ReadItem reads an item by ID and share ID. Works for all item types.
func (c *Client) ReadItem(ctx context.Context, itemID, shareID string) (*ItemJSON, error) {
	uri := fmt.Sprintf("pass://%s/%s", shareID, itemID)
	stdout, _, err := c.runner.Run(ctx, "item", "view", "--output=json", "--", uri)
	if err != nil {
		return nil, fmt.Errorf("failed to read item %q: %w", itemID, err)
	}

	// item view wraps in {"item": {...}, "attachments": [...]}
	var viewResp struct {
		Item ItemRawJSON `json:"item"`
	}
	if err := json.Unmarshal(stdout, &viewResp); err != nil {
		// Try parsing as a direct item.
		var raw ItemRawJSON
		if err2 := json.Unmarshal(stdout, &raw); err2 != nil {
			return nil, fmt.Errorf("failed to parse item view JSON: %w", err)
		}
		flat := FlattenItem(raw)
		return &flat, nil
	}
	flat := FlattenItem(viewResp.Item)
	return &flat, nil
}

// UpdateItem updates an existing item using --field key=value pairs. Works for all item types.
func (c *Client) UpdateItem(ctx context.Context, itemID, shareID string, fields map[string]string) error {
	args := []string{"item", "update",
		"--share-id=" + shareID,
		"--",
		itemID,
	}
	for k, v := range fields {
		args = append(args, "--field", fmt.Sprintf("%s=%s", k, v))
	}

	_, _, err := c.runner.Run(ctx, args...)
	if err != nil {
		return fmt.Errorf("failed to update item %q: %w", itemID, err)
	}
	return nil
}

// DeleteItem moves an item to the trash or deletes it permanently.
func (c *Client) DeleteItem(ctx context.Context, itemID, shareID string, destroyPermanently bool) error {
	cmd := "trash"
	if destroyPermanently {
		cmd = "delete"
	}
	_, _, err := c.runner.Run(ctx, "item", cmd, "--item-id="+itemID, "--share-id="+shareID)
	if err != nil {
		return fmt.Errorf("failed to %s item %q: %w", cmd, itemID, err)
	}
	return nil
}

// --- Item Note operations ---

// CreateItemNote creates a new note item.
func (c *Client) CreateItemNote(ctx context.Context, shareID, title, note string) (*ItemJSON, error) {
	args := []string{"item", "create", "note",
		"--share-id=" + shareID,
		"--title=" + title,
	}
	if note != "" {
		args = append(args, "--note="+note)
	}
	stdout, _, err := c.runner.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create note item: %w", err)
	}
	itemID := strings.TrimSpace(string(stdout))
	if itemID == "" {
		return nil, fmt.Errorf("failed to get item ID after create note")
	}
	return c.ReadItem(ctx, itemID, shareID)
}

// --- Item Credit Card operations ---

// CreateItemCreditCard creates a new credit card item.
func (c *Client) CreateItemCreditCard(ctx context.Context, shareID, title, cardholderName, cardNumber, cvv, expirationDate, pin string) (*ItemJSON, error) {
	args := []string{"item", "create", "credit-card",
		"--share-id=" + shareID,
		"--title=" + title,
	}
	if cardholderName != "" {
		args = append(args, "--cardholder-name="+cardholderName)
	}
	if cardNumber != "" {
		args = append(args, "--number="+cardNumber)
	}
	if cvv != "" {
		args = append(args, "--cvv="+cvv)
	}
	if expirationDate != "" {
		args = append(args, "--expiration-date="+expirationDate)
	}
	if pin != "" {
		args = append(args, "--pin="+pin)
	}
	_, _, err := c.runner.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create credit card item: %w", err)
	}
	return c.findItemByTitle(ctx, shareID, title)
}

// --- Item WiFi operations ---

// CreateItemWiFi creates a new WiFi item.
func (c *Client) CreateItemWiFi(ctx context.Context, shareID, title, ssid, password, security string) (*ItemJSON, error) {
	args := []string{"item", "create", "wifi",
		"--share-id=" + shareID,
		"--title=" + title,
	}
	if ssid != "" {
		args = append(args, "--ssid="+ssid)
	}
	if password != "" {
		args = append(args, "--password="+password)
	}
	if security != "" {
		args = append(args, "--security="+security)
	}
	_, _, err := c.runner.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create WiFi item: %w", err)
	}
	return c.findItemByTitle(ctx, shareID, title)
}

// --- Item Identity operations ---

// CreateItemIdentity creates a new identity item from a JSON template.
func (c *Client) CreateItemIdentity(ctx context.Context, shareID string, templateJSON []byte) (*ItemJSON, error) {
	tmpFile, err := WriteTempFile("protonpass-identity-*.json", templateJSON)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile)

	// Extract title from template.
	var tmpl map[string]string
	if err := json.Unmarshal(templateJSON, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse template JSON: %w", err)
	}

	stdout, _, err := c.runner.Run(ctx, "item", "create", "identity",
		"--share-id="+shareID,
		"--from-template="+tmpFile,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity item: %w", err)
	}

	itemID := strings.TrimSpace(string(stdout))
	if itemID == "" {
		return nil, fmt.Errorf("failed to parse identity creation response: empty output")
	}

	return c.ReadItem(ctx, itemID, shareID)
}

// --- Item SSH Key operations ---

// CreateItemSSHKey creates a new SSH key item.
func (c *Client) CreateItemSSHKey(ctx context.Context, shareID, title, keyType, comment string) (*ItemJSON, error) {
	args := []string{"item", "create", "ssh-key", "generate",
		"--share-id=" + shareID,
		"--title=" + title,
	}
	if keyType != "" {
		args = append(args, "--key-type="+keyType)
	}
	if comment != "" {
		args = append(args, "--comment="+comment)
	}
	_, _, err := c.runner.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to generate SSH key item: %w", err)
	}
	return c.findItemByTitle(ctx, shareID, title)
}

// CreateItemSSHKeyImport imports an SSH key from a private key file.
func (c *Client) CreateItemSSHKeyImport(ctx context.Context, shareID, title, privateKeyPath string) (*ItemJSON, error) {
	_, _, err := c.runner.Run(ctx, "item", "create", "ssh-key", "import",
		"--share-id="+shareID,
		"--title="+title,
		"--from-private-key="+privateKeyPath,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to import SSH key item: %w", err)
	}
	return c.findItemByTitle(ctx, shareID, title)
}
