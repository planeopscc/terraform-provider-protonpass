terraform {
  required_providers {
    protonpass = { source = "planeopscc/protonpass" }
  }
}
provider "protonpass" {}

resource "protonpass_vault" "test" {
  name = "Delete-Lifecycle-Test"
}

# All items with destroy_permanently = true
resource "protonpass_item" "login" {
  share_id            = protonpass_vault.test.share_id
  type                = "login"
  title               = "Test Login"
  username            = "admin"
  password            = "secret123"
  destroy_permanently = true
}

resource "protonpass_item" "note" {
  share_id            = protonpass_vault.test.share_id
  type                = "note"
  title               = "Test Note"
  note                = "This is a test note."
  destroy_permanently = true
}

resource "protonpass_item" "cc" {
  share_id            = protonpass_vault.test.share_id
  type                = "credit-card"
  title               = "Test CC"
  cardholder_name     = "Test User"
  number              = "4111111111111111"
  verification_number = "123"
  expiration_date     = "2030-12"
  destroy_permanently = true
}

resource "protonpass_item" "wifi" {
  share_id            = protonpass_vault.test.share_id
  type                = "wifi"
  title               = "Test WiFi"
  ssid                = "TestNet"
  password            = "wifipass"
  security            = "WPA2"
  destroy_permanently = true
}

resource "protonpass_item" "ssh" {
  share_id            = protonpass_vault.test.share_id
  type                = "ssh-key"
  title               = "Test SSH"
  generate            = true
  destroy_permanently = true
}

resource "protonpass_item" "identity" {
  share_id            = protonpass_vault.test.share_id
  type                = "identity"
  title               = "Test Identity"
  full_name           = "John Doe"
  email               = "john@example.com"
  destroy_permanently = true
}

output "share_id" {
  value = protonpass_vault.test.share_id
}

resource "protonpass_alias" "test_alias" {
  share_id = protonpass_vault.test.share_id
  prefix   = "test-terraform-alias"
}
