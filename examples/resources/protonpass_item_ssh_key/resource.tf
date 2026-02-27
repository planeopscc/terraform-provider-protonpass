resource "protonpass_item_ssh_key" "example" {
  share_id = protonpass_vault.example.share_id
  title    = "Key"
  key_type = "ed25519"
}
