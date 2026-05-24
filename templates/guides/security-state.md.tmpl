---
page_title: "Security: Terraform State and Secrets"
description: |-
  How sensitive attributes, write-only attributes, and state backends
  affect the security of secrets managed by the protonpass provider.
---

# Security: Terraform State and Secrets

The `protonpass` provider manages credentials, notes, and other secrets stored
in Proton Pass. This guide explains how Terraform handles sensitive data and
what you need to do to avoid storing secrets in plain text outside of Proton
Pass.

## Sensitive vs. Write-Only Attributes

Terraform distinguishes two mechanisms for protecting secret values:

| Mechanism | What it does | Stored in state? |
|---|---|---|
| `Sensitive` | Redacts the value in CLI output and plan diffs | **Yes** — in plain text |
| `WriteOnly` | Never reads or stores the value; passed directly to the provider | **No** |

> **Key point**: `Sensitive` only affects display. Anyone with access to the
> `.tfstate` file or the remote backend can read those values in plain text.

`WriteOnly` attributes (marked `[Write-only]` in the schema) require
**Terraform >= 1.11**. They are the safest way to provide secrets: the value
is consumed once at apply time and is never persisted.

## Attribute Reference

### `protonpass_item`

The table below lists every attribute that carries secret data, whether a
write-only alternative exists, and whether that alternative avoids state
storage.

| Attribute | Item type | Sensitive | Write-only alternative | Stored in state? |
|---|---|---|---|---|
| `password` | login, wifi | Yes | `password_wo` / `password_wo_version` | **Yes** — use `password_wo` instead |
| `note` | note | Yes | `note_wo` / `note_wo_version` | **Yes** — use `note_wo` instead |
| `ssn` | identity | Yes | `ssn_wo` / `ssn_wo_version` | **Yes** — use `ssn_wo` instead |
| `passport_number` | identity | Yes | `passport_number_wo` / `passport_number_wo_version` | **Yes** — use `passport_number_wo` instead |
| `license_number` | identity | Yes | `license_number_wo` / `license_number_wo_version` | **Yes** — use `license_number_wo` instead |
| `number` | credit-card | Yes | None | **Always in state** |
| `verification_number` | credit-card | Yes | None | **Always in state** |
| `pin` | credit-card | Yes | None | **Always in state** |
| `private_key` | ssh-key | Yes | None | **Always in state** |

**Recommendation**: always use the `_wo` variant when it exists. Use the plain
`Sensitive` variant only when you need Terraform to read the current value back
(e.g., to reference it in another resource).

### `protonpass_item` — Items with No Write-Only Alternative

`credit-card` and `ssh-key` items always store secret values in state:

- `number`, `verification_number`, `pin` — credit card fields
- `private_key` — SSH private key

If you use these item types, **treat your state file as a secret**. Apply the
backend and access-control recommendations below.

## Protecting the State Backend

### Use a Remote Backend with Encryption at Rest

Do not rely on a local `.tfstate` file for workspaces that contain credit card
or SSH key items. Use a remote backend that encrypts state at rest:

- **Terraform Cloud / HCP Terraform** — encryption at rest and access control
  built in
- **AWS S3 + KMS** — enable server-side encryption with a KMS key
- **Google Cloud Storage** — enable CMEK encryption
- **Azure Blob Storage** — enable storage service encryption

Example (S3 backend):

```hcl
terraform {
  backend "s3" {
    bucket         = "my-tfstate"
    key            = "protonpass/terraform.tfstate"
    region         = "eu-west-1"
    encrypt        = true
    kms_key_id     = "arn:aws:kms:eu-west-1:123456789012:key/..."
  }
}
```

### Restrict Access to the State

State access should follow the principle of least privilege:

- Grant read/write access only to the CI/CD role that runs `terraform apply`
- Grant read-only access to roles that only run `terraform plan`
- Do not grant state access to developers unless required for debugging
- Enable audit logging on the state backend

### Enable State Locking

Always enable state locking to prevent concurrent writes that could expose
partial state during a conflict resolution:

- S3: use a DynamoDB lock table
- Terraform Cloud: locking is automatic
- GCS: object versioning + locking is automatic

## CI/CD Recommendations

### Do Not Log Plans That Contain Secrets

`terraform plan` output includes `(sensitive value)` placeholders, but
complete plan JSON (`-json` flag, or `terraform show -json planfile`) can
expose Sensitive values depending on the Terraform version and tool.

- Avoid piping `terraform plan -json` output to log aggregators
- If plan output is stored as a CI artifact, restrict artifact access
- Do not print environment variables that contain secrets in CI logs

### Use Short-Lived Credentials for `pass-cli`

The provider calls `pass-cli` using the session established by `pass-cli login`.
In CI, prefer ephemeral sessions:

1. Inject credentials via environment variables supported by `pass-cli`
2. Revoke the session after the pipeline completes
3. Do not cache `pass-cli` session tokens in persistent CI storage

### Rotate Secrets Without Touching State

For `login` and `note` items that use write-only attributes, rotate secrets
without re-reading state:

```hcl
resource "protonpass_item" "app_db" {
  share_id            = protonpass_vault.ops.share_id
  type                = "login"
  title               = "App DB"
  username            = "app"
  password_wo         = var.db_password  # never stored in state
  password_wo_version = var.db_password_version  # increment to rotate
}
```

Increment `password_wo_version` (or any `*_wo_version` counter) to trigger an
update without Terraform comparing old and new secret values.

## Summary

| Risk | Mitigation |
|---|---|
| Secrets in local `.tfstate` | Use a remote backend with encryption at rest |
| Secrets visible in plain text | Prefer `_wo` write-only attributes |
| Credit card / SSH key always in state | Restrict backend access; encrypt at rest |
| CI logs exposing secrets | Avoid logging plan JSON; restrict artifacts |
| Stale `pass-cli` session in CI | Use ephemeral sessions; revoke after apply |
