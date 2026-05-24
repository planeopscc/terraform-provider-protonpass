# Terraform Provider for Proton Pass

Unofficial Terraform Provider to manage items and vaults via the Proton Pass CLI (`pass-cli`). 

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.11 (Required for `write-only` attribute support)
- [Go](https://golang.org/doc/install) >= 1.25.5 (For building the provider locally)
- [pass-cli](https://github.com/protonpass/pass-cli) >= 1.5.2 (tested; v2.0.3+ recommended — see [pass-cli compatibility guide](docs/guides/pass-cli-compatibility.md))

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
resource "protonpass_item" "my_login" {
  share_id            = protonpass_vault.my_vault.share_id
  type                = "login"
  title               = "GitHub Account"
  username            = "my_user"
  password_wo         = "super_secret_password"
  password_wo_version = 1
  urls                = ["https://github.com/login"]
}
```

### Sensitive attributes vs. Write-only attributes

> **Warning**: Attributes marked `Sensitive` (e.g. `password`, `note`, `ssn`, `number`, `pin`) are redacted in CLI output but are **still stored in the Terraform state file**. Anyone with access to your state (local `.tfstate`, remote backend) can read these values in plain text.
>
> Prefer the `_wo` write-only equivalents (`password_wo`, `note_wo`, `ssn_wo`, `passport_number_wo`, `license_number_wo`) when they exist — these are **never written to state**. Fields like `number` (card number), `verification_number` (CVV), and `pin` have no write-only equivalent; treat your state file as a secret for resources that use them.

### Rotating Passwords
Values like `password_wo` inside the Login Item or `note_wo` in Note Items use the Terraform `>= 1.11` write-only attribute type. This means they are passed to Proton Pass but are **never stored in your local state**. To trigger a password rotation, update the `password_wo` field and increment `password_wo_version`.

### Importing Existing Items
Because items in Proton Pass require a context to lookup, you must provide both the Share ID (Vault ID) and the Item ID using a composite format (`share_id:item_id`). 

```shell
# Import a vault using its share_id
terraform import protonpass_vault.mega_vault "share_id"

# Import a login item using its composite share_id and item_id
terraform import protonpass_item.full_login "share_id:item_id"
```

## Developing the Provider

To compile the provider, run `make install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.
To generate or update documentation, run `make generate`.

```shell
make generate
make lint
make test
```
