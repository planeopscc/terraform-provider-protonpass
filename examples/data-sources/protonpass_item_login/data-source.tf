data "protonpass_item_login" "example" {
  item_id  = "item_id_here"
  share_id = "share_id_here"
}

output "username" {
  value = data.protonpass_item_login.example.username
}
