resource "protonpass_alias" "example" {
  share_id            = protonpass_vault.example.share_id
  prefix              = "netflix_signup"
  destroy_permanently = true
}
