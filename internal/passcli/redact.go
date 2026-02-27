// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package passcli

// RedactArgs returns a copy of args with sensitive values replaced by [REDACTED].
// Values immediately following sensitive flags are redacted.
func RedactArgs(args []string) []string {
	sensitiveFlags := map[string]bool{
		"--password": true,
		"--totp-uri": true,
		"--note":     true,
		"--secret":   true,
		"--api-key":  true,
		"--number":   true, // Credit Card
		"--cvv":      true, // Credit Card
		"--pin":      true, // Credit Card
	}

	result := make([]string, len(args))
	redactNext := false

	for i, arg := range args {
		if redactNext {
			result[i] = "[REDACTED]"
			redactNext = false
			continue
		}
		if sensitiveFlags[arg] {
			result[i] = arg
			redactNext = true
			continue
		}
		result[i] = arg
	}

	return result
}
