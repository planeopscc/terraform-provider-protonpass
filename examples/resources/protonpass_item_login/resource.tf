resource "protonpass_item_login" "example" {
  share_id            = protonpass_vault.example.share_id
  title               = "My Login"
  username            = "user"
  password_wo         = "secret"
  password_wo_version = 1
}
