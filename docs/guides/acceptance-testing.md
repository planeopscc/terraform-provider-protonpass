---
page_title: "Running Acceptance Tests"
description: |-
  How to run the protonpass provider acceptance tests against a real
  pass-cli session.
---

# Running Acceptance Tests

Acceptance tests for the `protonpass` provider run against a real `pass-cli`
session. They create and delete actual Proton Pass resources, so they require
a working internet connection and an authenticated account.

These tests are **never required** in the public CI. They are intended for
local development and for maintainers who have access to a Proton Pass
account.

## Prerequisites

1. **Terraform >= 1.11** — required for write-only attribute support.

2. **pass-cli installed and authenticated**:

   ```shell
   pip install pass-cli
   pass-cli login
   pass-cli test   # must print "Connection successful"
   ```

3. **`TF_ACC=1`** — the standard Terraform guard that gates all acceptance
   tests. Without this variable, every acceptance test is silently skipped.

## Running the Tests

```shell
TF_ACC=1 go test -v -timeout 120s ./internal/passcli/... -run TestAcc_
```

To run all acceptance tests across the whole repo:

```shell
TF_ACC=1 go test -v -timeout 300s ./... -run TestAcc_
```

Using the Makefile target:

```shell
make testacc
```

> The `testacc` target sets `TF_ACC=1` and runs `go test ./...` with a
> 120-minute timeout. Use `-run TestAcc_` to target only acceptance tests
> when running alongside unit tests.

## Environment Variables

| Variable | Required | Default | Description |
|---|---|---|---|
| `TF_ACC` | **Yes** | — | Set to any non-empty value to enable acceptance tests. |
| `PROTONPASS_ACC_CLI_PATH` | No | `pass-cli` | Path to the `pass-cli` binary, if not in `PATH`. |

No secrets or credentials need to be injected as environment variables. The
`pass-cli` session established by `pass-cli login` is used automatically.

## Test Naming Convention

All acceptance test functions are prefixed with `TestAcc_` to distinguish them
from unit tests. This allows you to run them selectively with `-run TestAcc_`
and to exclude them from coverage reports that should only reflect unit tests.

## Test Data and Cleanup

Every acceptance test:

- Generates a unique resource name with a `tf-acc-` prefix and a millisecond
  timestamp (e.g., `tf-acc-vault-1716550312345`). This makes test resources
  easy to identify and clean up if a test is interrupted.
- Registers a `t.Cleanup` function that deletes the resource after the test
  completes, whether it passes or fails.

If a test is interrupted before cleanup runs (e.g., killed with Ctrl+C), you
can find and remove leftover resources in Proton Pass by looking for items
whose name starts with `tf-acc-`.

## Adding New Acceptance Tests

Use the helpers in `internal/testutil/acctest.go`:

```go
func TestAcc_MyTest(t *testing.T) {
    testutil.SkipIfNotAcc(t)                // skip if TF_ACC is not set
    client := testutil.NewAccClient(t)      // skip if pass-cli not found

    if err := client.HealthCheck(t.Context()); err != nil {
        t.Skipf("pass-cli session not active: %v", err)
    }

    // ... test body ...
    // Always defer cleanup:
    t.Cleanup(func() { /* delete the resource */ })
}
```
