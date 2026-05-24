# Example 1: Create a Login Item
# Use password_wo + password_wo_version to avoid storing the password in state.
# Increment password_wo_version to trigger a password rotation.
resource "protonpass_item" "example_login" {
  share_id            = protonpass_vault.example.share_id
  type                = "login"
  title               = "My Login"
  username            = "user"
  password_wo         = "secret"
  password_wo_version = 1
  urls                = ["https://example.com"]
  destroy_permanently = true
}

# Example 2: Create a Note Item
resource "protonpass_item" "example_note" {
  share_id = protonpass_vault.example.share_id
  type     = "note"
  title    = "My Note"
  note_wo  = "This is a super secret note."
}

# Example 3: Create a Credit Card Item
# Note: number, verification_number, and pin are Sensitive but stored in state.
# Treat your Terraform state file as a secret when using credit card items.
resource "protonpass_item" "example_card" {
  share_id        = protonpass_vault.example.share_id
  type            = "credit-card"
  title           = "Company Card"
  cardholder_name = "Jane Doe"
  number          = "4111222233334444"
  expiration_date = "2030-12"
  pin             = "1234"
}

# Example 4: Create a WiFi Item
# Use password_wo to avoid storing the WiFi password in state.
resource "protonpass_item" "example_wifi" {
  share_id            = protonpass_vault.example.share_id
  type                = "wifi"
  title               = "Office Guest WiFi"
  ssid                = "Guest-Net"
  password_wo         = "guest-password"
  password_wo_version = 1
  security            = "WPA2"
}

# Example 5: Generate an SSH Key
resource "protonpass_item" "example_ssh_gen" {
  share_id = protonpass_vault.example.share_id
  type     = "ssh-key"
  title    = "Generated Deploy Key"
  generate = true # This will let the provider generate the key pair
}

# Example 6: Create an Identity Item
resource "protonpass_item" "example_identity" {
  share_id  = protonpass_vault.example.share_id
  type      = "identity"
  title     = "Jane Doe Identity"
  full_name = "Jane Doe"
  email     = "jane.doe@example.com"
}
