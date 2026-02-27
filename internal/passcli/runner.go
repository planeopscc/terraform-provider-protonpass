// Copyright (c) PlaneOpsCc
// SPDX-License-Identifier: MPL-2.0

package passcli

import "context"

// Runner abstracts CLI execution for testing.
type Runner interface {
	Run(ctx context.Context, args ...string) (stdout []byte, stderr []byte, err error)
}
