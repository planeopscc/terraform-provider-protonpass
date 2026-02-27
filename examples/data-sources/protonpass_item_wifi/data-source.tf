data "protonpass_item_wifi" "example" {
  item_id  = "item_id_here"
  share_id = "share_id_here"
}

output "ssid" {
  value = data.protonpass_item_wifi.example.ssid
}
