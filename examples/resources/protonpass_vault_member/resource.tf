resource "protonpass_vault_member" "example" {
  share_id = protonpass_vault.example.share_id
  email    = "someone@example.com"
  role     = "viewer"
}
