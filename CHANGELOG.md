## 1.2.0 (2026-05-25)

### Added
- `protonpass_item` data source now exposes all identity fields: `first_name`,
  `middle_name`, `last_name`, `birthdate`, `gender`, `organization`,
  `street_address`, `zip_or_postal_code`, `city`, `state_or_province`,
  `country_or_region` (#34).
- Unit tests for `vault_resource` and `vault_member_resource` covering
  not-found Read, error diagnostics, idempotent Delete, and Update paths (#36).

### Fixed
- `VaultMemberResource.Delete` is now idempotent: not-found errors are silently
  ignored, consistent with `AliasResource.Delete` (#29).
- `passcli.IsNotFound` is now used consistently for all not-found checks in
  `alias_resource` and `vault_member_resource`, replacing ad-hoc
  `strings.Contains` calls (#28).
- `ItemResource.Create` emits a warning diagnostic when a trashed item is
  restored instead of a new one being created, making the implicit behavior
  visible to operators (#35).
- `ItemDataSource.URLs` field changed from `[]types.String` to `types.List`,
  consistent with `ItemResource` (#33).
- `items_datasource` result slice initialised as empty (not nil) to guarantee
  a consistent `[]` in state when no items match (#31).
- `generate` attribute on `ssh-key` items now requires resource replacement
  when changed: `generate = true` and `generate = false` invoke different
  creation paths (`CreateItemSSHKey` vs `CreateItemSSHKeyImport`) (#32).
- `destroy_permanently` (alias + item) and `generate` (item) now default to
  `false` in state, eliminating null/false inconsistency (#30).
- `type` attribute on `protonpass_item` resource now validated with
  `stringvalidator.OneOf`, surfacing invalid types at plan time instead of CLI
  execution time (#38).

### Security
- `ItemResource` CRUD methods now call `tflog.MaskFieldValuesWithFieldKeys` for
  all sensitive field keys (`password`, `note`, `number`, `verification_number`,
  `pin`, `private_key`, `ssn`, `passport_number`, `license_number`), preventing
  accidental leakage into debug logs (#37).

## 1.1.0 (2026-05-25)

### Added
- `agent_reason` optional provider attribute: injects `PROTON_PASS_AGENT_REASON`
  into each pass-cli subprocess for agent token workflows (pass-cli v2.1.0+).
  Provider config takes priority over any same-key process environment variable.
- Acceptance test infrastructure (`TestAcc_` prefix, `TF_ACC` guard, `testutil`
  helpers) with a vault lifecycle test against a live pass-cli session.
- Documentation guides: pass-cli compatibility matrix, CI authentication
  patterns, and Terraform state security.

### Fixed
- `pass-cli item update` now uses `--item-id=<id>` (named flag) instead of
  a positional argument after `--`, which was silently ignored by the CLI.
- Health check now relies solely on the pass-cli exit code. Previously, a
  valid session was incorrectly rejected if stdout did not contain the exact
  string `"Connection successful"`.
- Secret arguments (`--field password=…`) are now fully redacted in error
  messages, covering both concatenated (`--flag=value`) and field-pair forms.
- Item import now correctly sets the resource type from the remote object
  instead of leaving it null.
- Item Read no longer silently restores trashed items; a missing item now
  returns the correct not-found signal to Terraform.

### Security
- Secret field values are never included in CLI error messages or diagnostics.
- `agent_reason` is injected via per-subprocess `cmd.Env`, not `os.Setenv`,
  avoiding any global process environment mutation.

## 1.0.1 (2026-03-11)

- Provider refactor, aliases, and TOTP support.

## 1.0.0 (2026-02-27)

- Initial release with vault and item resources/data sources.
