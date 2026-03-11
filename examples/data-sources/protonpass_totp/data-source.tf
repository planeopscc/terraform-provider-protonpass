data "protonpass_totp" "example" {
  share_id = "your-vault-share-id"
  item_id  = "your-item-id"
}

output "my_totp_code" {
  value     = data.protonpass_totp.example.code
  sensitive = true
}
