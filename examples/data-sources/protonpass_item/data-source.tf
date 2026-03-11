# Example 1: Retrieve a Login Item
data "protonpass_item" "github" {
  share_id = protonpass_vault.example.share_id
  name     = "github"
  type     = "login"
}

output "github_username" {
  value = data.protonpass_item.github.username
}

# Example 2: Retrieve a generated SSH Public Key
data "protonpass_item" "deploy_key" {
  share_id = protonpass_vault.example.share_id
  name     = "Generated Deploy Key"
  type     = "ssh-key"
}

output "ssh_public_key" {
  value = data.protonpass_item.deploy_key.public_key
}

# Example 3: Retrieve Identity Info
data "protonpass_item" "identity" {
  share_id = protonpass_vault.example.share_id
  name     = "Jane Doe Identity"
  type     = "identity"
}

output "identity_email" {
  value = data.protonpass_item.identity.email
}
