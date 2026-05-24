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
