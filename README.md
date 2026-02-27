# Terraform Provider for Proton Pass

Unofficial Terraform Provider to manage items and vaults via the Proton Pass CLI (`pass-cli`). 

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.11 (Required for `write-only` attribute support)
- [Go](https://golang.org/doc/install) >= 1.24 (For building the provider locally)
- [pass-cli](https://github.com/ProtonMail/pass-cli) >= 1.0 (Must be installed and authenticated via `pass-cli login`)

## Installation

This provider relies on the local environment having an authenticated active `pass-cli` session.

1. Install the CLI: `pip install pass-cli`
2. Authenticate: `pass-cli login`
3. Verify session: `pass-cli test`

To install the provider manually for local development:
```shell
make install
```

## Example Usage

### Setting up the Provider
```hcl
terraform {
  required_providers {
    protonpass = {
      source = "planeopscc/protonpass"
    }
  }
}

provider "protonpass" {}
```

### Creating Resources
```hcl
# Create a Vault
resource "protonpass_vault" "my_vault" {
  name = "My Secure Vault"
}

# Create a Login Item
resource "protonpass_item_login" "my_login" {
  share_id            = protonpass_vault.my_vault.share_id
  title               = "GitHub Account"
  username            = "my_user"
  password_wo         = "super_secret_password"
  password_wo_version = 1
  urls                = ["https://github.com/login"]
}
```

### Rotating Passwords
Values like `password_wo` inside the Login Item or `note_wo` in Note Items use the Terraform `>= 1.11` write-only attribute type. This means they are passed to Proton Pass but are **never stored in your local state**. To trigger a password rotation, simply update the `password_wo` field and increment `password_wo_version`.

### Importing Existing Items
Because items in Proton Pass require a context to lookup, you must provide both the Share ID (Vault ID) and the Item ID using a composite format (`share_id:item_id`). 

```shell
# Import a vault using its share_id
terraform import protonpass_vault.mega_vault "share_id"

# Import a login item using its composite share_id and item_id
terraform import protonpass_item_login.full_login "share_id:item_id"
```

## Developing the Provider

To compile the provider, run `make install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.
To generate or update documentation, run `make generate`.

```shell
make generate
make lint
make test
```
