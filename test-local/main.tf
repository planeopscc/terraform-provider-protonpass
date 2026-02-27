terraform {
  required_providers {
    protonpass = {
      source = "planeopscc/protonpass"
    }
  }
}

provider "protonpass" {}

# ==========================================
# 1. VAULTS
# ==========================================
resource "protonpass_vault" "mega_vault" {
  name = "TF_Mega_Test_Vault"
}

# ==========================================
# 2. ITEM: LOGIN
# ==========================================
resource "protonpass_item_login" "full_login" {
  share_id            = protonpass_vault.mega_vault.share_id
  title               = "Ultimate Login Test"
  username            = "test_user_99"
  email               = "test99@example.com"
  password_wo         = "SuperSecretPassword123!"
  password_wo_version = 1
  urls                = ["https://example.com/login", "https://app.example.com"]
}

# ==========================================
# 3. ITEM: NOTE
# ==========================================
resource "protonpass_item_note" "full_note" {
  share_id        = protonpass_vault.mega_vault.share_id
  title           = "Massive Secure Note"
  note_wo         = "Line 1\nLine 2\nLine 3\nVery long secure text..."
  note_wo_version = 1
}

# ==========================================
# 4. ITEM: CREDIT CARD
# ==========================================
resource "protonpass_item_credit_card" "full_cc" {
  share_id               = protonpass_vault.mega_vault.share_id
  title                  = "Corporate Platinum"
  cardholder_name        = "Alice B. Smith"
  card_number_wo         = "4000123456789010"
  card_number_wo_version = 1
  cvv_wo                 = "456"
  cvv_wo_version         = 1
  expiration_date        = "2030-12"
  pin_wo                 = "9876"
  pin_wo_version         = 1
}

# ==========================================
# 5. ITEM: WIFI (Multiple types)
# ==========================================
resource "protonpass_item_wifi" "wpa3_wifi" {
  share_id            = protonpass_vault.mega_vault.share_id
  title               = "HQ WPA3 Network"
  ssid                = "HQ-Secure-5G"
  password_wo         = "hq_pass_123"
  password_wo_version = 1
  security            = "wpa3"
}

resource "protonpass_item_wifi" "open_wifi" {
  share_id = protonpass_vault.mega_vault.share_id
  title    = "Public Cafe"
  ssid     = "Cafe-Free-WiFi"
  security = "none" # No password for open network
}

# ==========================================
# 6. ITEM: SSH KEY (Multiple types)
# ==========================================
resource "protonpass_item_ssh_key" "ed25519_key" {
  share_id = protonpass_vault.mega_vault.share_id
  title    = "Modern ED25519 Key"
  key_type = "ed25519"
  comment  = "alice@modern-laptop"
}

resource "protonpass_item_ssh_key" "rsa_key" {
  share_id = protonpass_vault.mega_vault.share_id
  title    = "Legacy RSA Key"
  key_type = "rsa4096"
  comment  = "legacy_system_access"
}

# ==========================================
# 7. ITEM: IDENTITY (The big one)
# ==========================================
resource "protonpass_item_identity" "full_identity" {
  share_id = protonpass_vault.mega_vault.share_id
  title    = "Alice's Complete Identity"

  # Basic Info
  first_name   = "Alice"
  middle_name  = "Jane"
  last_name    = "Doe"
  full_name    = "Alice Jane Doe"
  birthdate    = "1990-01-01"
  gender       = "Female"
  phone_number = "+1-555-0100"
  email        = "alice.j.doe@example.com"

  # Address
  organization       = "Acme Corporation"
  street_address     = "123 Innovation Way"
  city               = "Techville"
  zip_or_postal_code = "90210"
  country_or_region  = "USA"

  # Work
  company           = "Acme Corp"
  job_title         = "Senior Engineer"
  work_email        = "alice@acme.corp"
  work_phone_number = "+1-555-0199"

  # Secure Identifiers (Write-Only)
  ssn_wo                     = "000-11-2222"
  ssn_wo_version             = 1
  passport_number_wo         = "A12345678"
  passport_number_wo_version = 1
  license_number_wo          = "DL-999-888"
  license_number_wo_version  = 1

  # Socials
  website = "https://alicedoe.dev"
}

# ==========================================
# 8. DATA SOURCES VERIFICATION
# ==========================================

data "protonpass_items" "vault_list" {
  share_id = protonpass_vault.mega_vault.share_id
  depends_on = [
    protonpass_item_login.full_login,
    protonpass_item_note.full_note,
    protonpass_item_credit_card.full_cc,
    protonpass_item_wifi.wpa3_wifi,
    protonpass_item_wifi.open_wifi,
    protonpass_item_ssh_key.ed25519_key,
    protonpass_item_ssh_key.rsa_key,
    protonpass_item_identity.full_identity
  ]
}

data "protonpass_item_identity" "read_identity" {
  item_id  = protonpass_item_identity.full_identity.item_id
  share_id = protonpass_vault.mega_vault.share_id
}

data "protonpass_item_credit_card" "read_cc" {
  item_id  = protonpass_item_credit_card.full_cc.item_id
  share_id = protonpass_vault.mega_vault.share_id
}

# ==========================================
# 9. OUTPUTS
# ==========================================
output "total_items_created" {
  value = length(data.protonpass_items.vault_list.items)
}

output "identity_full_name" {
  value = data.protonpass_item_identity.read_identity.full_name
}

output "identity_job_title" {
  value = data.protonpass_item_identity.read_identity.job_title
}

output "cc_cardholder" {
  value = data.protonpass_item_credit_card.read_cc.cardholder_name
}

output "cc_expiration" {
  value = data.protonpass_item_credit_card.read_cc.expiration_date
}
