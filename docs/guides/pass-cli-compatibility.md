---
page_title: "pass-cli Compatibility"
description: |-
  Tested and recommended versions of pass-cli for use with the protonpass
  provider, plus environment variables for CI/CD authentication.
---

# pass-cli Compatibility

The `protonpass` provider delegates all Proton Pass operations to the
`pass-cli` binary. This guide documents which versions have been tested,
which are recommended, and how to authenticate in CI environments.

## Version Matrix

| Version | Status | Notes |
|---|---|---|
| **v1.5.2** | Tested locally | Minimum version confirmed to work with this provider |
| **v1.6.1 – v1.10.0** | Untested | Later 1.x releases; no breaking changes expected but not verified |
| **v2.0.0 – v2.0.3** | Recommended | First stable 2.x series; introduces PAT authentication |
| **v2.1.0** | Latest upstream | Most recent tagged release at time of writing |

> **Tested** means a real `pass-cli` session was used to exercise the
> provider during development. **Untested** means the version is tagged
> upstream but has not been verified against this provider.

### Minimum Tested Version: v1.5.2

The provider was developed and verified against `pass-cli` v1.5.2. All
core operations (vault create/read/delete, item create/read/update/delete)
were confirmed working against this version.

### Recommended Version: v2.0.3+

The 2.x line adds Personal Access Token (PAT) support, which is the
recommended authentication method for CI/CD pipelines. Use v2.0.3 or
later for new installations.

To check your installed version:

```shell
pass-cli --version
```

To install or upgrade:

```shell
pip install --upgrade pass-cli
```

## Verifying Your Session

The provider calls `pass-cli test` as a health check at the start of each
Terraform operation. It considers the session valid if the command exits
with code 0.

```shell
pass-cli test
```

A successful response exits 0 regardless of the exact stdout message.
A non-zero exit code indicates the session is not active.

## Authentication in CI/CD

### Interactive Session (v1.x and v2.x)

For local development and simple CI setups, authenticate once with:

```shell
pass-cli login
pass-cli test
```

The session is stored locally and reused by subsequent `pass-cli` calls.

### Personal Access Token (v2.x only)

For CI/CD pipelines, use a Personal Access Token (PAT) to avoid
interactive login. PATs grant scoped access to specific vaults and items.

Generate a token in your Proton Pass account settings, then either pass
it as a flag:

```shell
pass-cli login --personal-access-token "$PROTON_PASS_PERSONAL_ACCESS_TOKEN"
```

Or set the environment variable and let `pass-cli` pick it up
automatically at login:

```shell
export PROTON_PASS_PERSONAL_ACCESS_TOKEN="<your-token>"
pass-cli login
pass-cli test
```

> Do not hardcode the token value. Inject it via your CI secret store
> (GitHub Actions secrets, HashiCorp Vault, AWS Secrets Manager, etc.).

## Environment Variables Reference

The following environment variables are documented by `pass-cli` for
authentication. They apply to the `pass-cli login` command.

| Variable | Purpose | Version |
|---|---|---|
| `PROTON_PASS_PERSONAL_ACCESS_TOKEN` | Token for PAT-based login | v2.x+ |
| `PROTON_PASS_PASSWORD` | Account password for interactive login | v1.x+ |
| `PROTON_PASS_PASSWORD_FILE` | Path to a file containing the password | v1.x+ |
| `PROTON_PASS_TOTP` | TOTP code for two-factor authentication | v1.x+ |
| `PROTON_PASS_TOTP_FILE` | Path to a file containing the TOTP code | v1.x+ |
| `PROTON_PASS_EXTRA_PASSWORD` | Pass-specific extra password | v1.x+ |
| `PROTON_PASS_EXTRA_PASSWORD_FILE` | Path to a file containing the extra password | v1.x+ |

> **Security note**: never set `PROTON_PASS_PASSWORD` or
> `PROTON_PASS_PERSONAL_ACCESS_TOKEN` in files committed to source control.
> Always inject them from a secret store at runtime.

The `protonpass` Terraform provider itself reads none of these variables
directly. They are consumed by `pass-cli` during session setup, before
Terraform runs.

## Example: GitHub Actions

```yaml
- name: Set up pass-cli session
  run: |
    pip install pass-cli
    pass-cli login
    pass-cli test
  env:
    PROTON_PASS_PERSONAL_ACCESS_TOKEN: ${{ secrets.PROTON_PASS_PAT }}

- name: Terraform apply
  run: terraform apply -auto-approve
```

## Upstream Resources

- [pass-cli releases](https://github.com/protonpass/pass-cli/tags)
- [pass-cli documentation](https://protonpass.github.io/pass-cli/)
