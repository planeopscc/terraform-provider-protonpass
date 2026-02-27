data "protonpass_item_identity" "example" {
  item_id  = "item_id_here"
  share_id = "share_id_here"
}

output "full_name" {
  value = data.protonpass_item_identity.example.full_name
}
