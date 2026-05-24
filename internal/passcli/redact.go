// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package passcli

import "strings"

// RedactArgs returns a copy of args with sensitive values replaced by [REDACTED].
// It handles three forms:
//   - space-separated: --flag value
//   - concatenated:    --flag=value
//   - field pair:      --field key=value (when key is sensitive)
func RedactArgs(args []string) []string {
	sensitiveFlags := map[string]bool{
		"--password": true,
		"--totp-uri": true,
		"--note":     true,
		"--secret":   true,
		"--api-key":  true,
		"--number":   true,
		"--cvv":      true,
		"--pin":      true,
	}

	sensitiveFieldKeys := map[string]bool{
		"password":               true,
		"note":                   true,
		"number":                 true,
		"verification_number":    true,
		"pin":                    true,
		"social_security_number": true,
		"passport_number":        true,
		"license_number":         true,
		"private_key":            true,
		"totp_uri":               true,
	}

	result := make([]string, len(args))
	redactNext := false
	redactNextField := false

	for i, arg := range args {
		if redactNext {
			result[i] = "[REDACTED]"
			redactNext = false
			continue
		}

		if redactNextField {
			redactNextField = false
			if eqIdx := strings.IndexByte(arg, '='); eqIdx >= 0 && sensitiveFieldKeys[arg[:eqIdx]] {
				result[i] = arg[:eqIdx] + "=[REDACTED]"
				continue
			}
			result[i] = arg
			continue
		}

		if arg == "--field" {
			result[i] = arg
			redactNextField = true
			continue
		}

		if sensitiveFlags[arg] {
			result[i] = arg
			redactNext = true
			continue
		}

		// Concatenated form: --flag=value
		redacted := false
		for flag := range sensitiveFlags {
			if strings.HasPrefix(arg, flag+"=") {
				result[i] = flag + "=[REDACTED]"
				redacted = true
				break
			}
		}
		if !redacted {
			result[i] = arg
		}
	}

	return result
}
