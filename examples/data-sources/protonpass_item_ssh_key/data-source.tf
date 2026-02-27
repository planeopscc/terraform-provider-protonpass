data "protonpass_item_ssh_key" "example" {
  item_id  = "item_id_here"
  share_id = "share_id_here"
}

output "public_key" {
  value = data.protonpass_item_ssh_key.example.public_key
}
