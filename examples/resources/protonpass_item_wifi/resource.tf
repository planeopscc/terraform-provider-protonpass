resource "protonpass_item_wifi" "example" {
  share_id            = protonpass_vault.example.share_id
  title               = "WiFi"
  ssid                = "Net"
  password_wo         = "pw"
  password_wo_version = 1
  security            = "wpa3"
}
