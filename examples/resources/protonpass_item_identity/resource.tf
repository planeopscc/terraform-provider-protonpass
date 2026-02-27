resource "protonpass_item_identity" "example" {
  share_id  = protonpass_vault.example.share_id
  title     = "Identity"
  full_name = "John Doe"
  email     = "john@example.com"
}
