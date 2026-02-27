// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package passcli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	stdout, _, err := c.runner.Run(ctx, "vault", "list", "--output", "json")
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
	_, _, err := c.runner.Run(ctx, "vault", "create", "--name", name)
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
	_, _, err := c.runner.Run(ctx, "vault", "update", "--share-id="+shareID, "--name", newName)
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

// --- Item operations ---

// ListItemsInVault lists all items in a vault.
func (c *Client) ListItemsInVault(ctx context.Context, shareID string) ([]ItemLoginJSON, error) {
	stdout, _, err := c.runner.Run(ctx, "item", "list", "--share-id="+shareID, "--output", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}
	var resp struct {
		Items []ItemRawJSON `json:"items"`
	}
	if err := json.Unmarshal(stdout, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse items JSON: %w", err)
	}
	result := make([]ItemLoginJSON, len(resp.Items))
	for i, raw := range resp.Items {
		result[i] = FlattenItem(raw)
	}
	return result, nil
}

// GetItem returns the full raw JSON for a single item.
// It can use either itemID or title (title is less reliable if duplicates exist, but useful right after creation).
func (c *Client) GetItem(ctx context.Context, itemID, title, shareID string) (*ItemRawJSON, error) {
	args := []string{"item", "view", "--output", "json", "--share-id", shareID}
	if itemID != "" {
		args = append(args, "--item-id", itemID)
	} else if title != "" {
		args = append(args, "--item-title", title)
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
func (c *Client) findItemByTitle(ctx context.Context, shareID, title string) (*ItemLoginJSON, error) {
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

// writeTempFile writes data to a temporary file and returns the path.
func writeTempFile(prefix string, data []byte) (string, error) {
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, prefix)
	if err := os.WriteFile(tmpFile, data, 0600); err != nil {
		return "", fmt.Errorf("failed to write temp file: %w", err)
	}
	return tmpFile, nil
}

// CreateItemLogin creates a new login item.
func (c *Client) CreateItemLogin(ctx context.Context, shareID, title, username, password, email string, urls []string) (*ItemLoginJSON, error) {
	args := []string{"item", "create", "login",
		"--share-id", shareID,
		"--title", title,
	}
	if username != "" {
		args = append(args, "--username", username)
	}
	if password != "" {
		args = append(args, "--password", password)
	}
	if email != "" {
		args = append(args, "--email", email)
	}
	for _, u := range urls {
		args = append(args, "--url", u)
	}
	_, _, err := c.runner.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create login item: %w", err)
	}
	return c.findItemByTitle(ctx, shareID, title)
}

// ReadItemLogin reads a login item by item ID and share ID.
func (c *Client) ReadItemLogin(ctx context.Context, itemID, shareID string) (*ItemLoginJSON, error) {
	stdout, _, err := c.runner.Run(ctx, "item", "view", "--item-id="+itemID, "--share-id="+shareID, "--output", "json")
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

// UpdateItemLogin updates an existing login item using --field key=value pairs.
func (c *Client) UpdateItemLogin(ctx context.Context, itemID, shareID string, fields map[string]string) error {
	args := []string{"item", "update",
		"--item-id=" + itemID,
		"--share-id=" + shareID,
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

// DeleteItemLogin deletes a login item.
func (c *Client) DeleteItemLogin(ctx context.Context, itemID, shareID string) error {
	_, _, err := c.runner.Run(ctx, "item", "delete", "--item-id="+itemID, "--share-id="+shareID)
	if err != nil {
		return fmt.Errorf("failed to delete item %q: %w", itemID, err)
	}
	return nil
}

// --- Item Note operations ---

// CreateItemNote creates a new note item.
func (c *Client) CreateItemNote(ctx context.Context, shareID, title, note string) (*ItemLoginJSON, error) {
	args := []string{"item", "create", "note",
		"--share-id", shareID,
		"--title", title,
	}
	if note != "" {
		args = append(args, "--note", note)
	}
	_, _, err := c.runner.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create note item: %w", err)
	}
	return c.findItemByTitle(ctx, shareID, title)
}

// --- Item Credit Card operations ---

// CreateItemCreditCard creates a new credit card item.
func (c *Client) CreateItemCreditCard(ctx context.Context, shareID, title, cardholderName, cardNumber, cvv, expirationDate, pin string) (*ItemLoginJSON, error) {
	args := []string{"item", "create", "credit-card",
		"--share-id", shareID,
		"--title", title,
	}
	if cardholderName != "" {
		args = append(args, "--cardholder-name", cardholderName)
	}
	if cardNumber != "" {
		args = append(args, "--number", cardNumber)
	}
	if cvv != "" {
		args = append(args, "--cvv", cvv)
	}
	if expirationDate != "" {
		args = append(args, "--expiration-date", expirationDate)
	}
	if pin != "" {
		args = append(args, "--pin", pin)
	}
	_, _, err := c.runner.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create credit card item: %w", err)
	}
	return c.findItemByTitle(ctx, shareID, title)
}

// --- Item WiFi operations ---

// CreateItemWiFi creates a new WiFi item.
func (c *Client) CreateItemWiFi(ctx context.Context, shareID, title, ssid, password, security string) (*ItemLoginJSON, error) {
	args := []string{"item", "create", "wifi",
		"--share-id", shareID,
		"--title", title,
	}
	if ssid != "" {
		args = append(args, "--ssid", ssid)
	}
	if password != "" {
		args = append(args, "--password", password)
	}
	if security != "" {
		args = append(args, "--security", security)
	}
	_, _, err := c.runner.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create WiFi item: %w", err)
	}
	return c.findItemByTitle(ctx, shareID, title)
}

// --- Item Identity operations ---

// CreateItemIdentity creates a new identity item from a JSON template.
func (c *Client) CreateItemIdentity(ctx context.Context, shareID string, templateJSON []byte) (*ItemLoginJSON, error) {
	tmpFile, err := writeTempFile("protonpass-identity-*.json", templateJSON)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile)

	// Extract title from template for lookup.
	var tmpl map[string]string
	if err := json.Unmarshal(templateJSON, &tmpl); err != nil {
		return nil, fmt.Errorf("failed to parse template JSON: %w", err)
	}
	title := tmpl["title"]

	_, _, err = c.runner.Run(ctx, "item", "create", "identity",
		"--share-id", shareID,
		"--from-template", tmpFile,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create identity item: %w", err)
	}
	return c.findItemByTitle(ctx, shareID, title)
}

// --- Item SSH Key operations ---

// CreateItemSSHKey creates a new SSH key item.
func (c *Client) CreateItemSSHKey(ctx context.Context, shareID, title, keyType, comment string) (*ItemLoginJSON, error) {
	args := []string{"item", "create", "ssh-key", "generate",
		"--share-id", shareID,
		"--title", title,
	}
	if keyType != "" {
		args = append(args, "--key-type", keyType)
	}
	if comment != "" {
		args = append(args, "--comment", comment)
	}
	_, _, err := c.runner.Run(ctx, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH key item: %w", err)
	}
	return c.findItemByTitle(ctx, shareID, title)
}
