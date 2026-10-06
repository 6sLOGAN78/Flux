package app

import "github.com/6sLOGAN78/flux/internal/lifecycle"

// Cleanup preserves the app ownership API over the infrastructure-neutral primitive.
type Cleanup = lifecycle.Cleanup

// ResourceError retains safe named cleanup failures and private causes.
type ResourceError = lifecycle.ResourceError

// ErrCleanupClosed rejects registration after cleanup begins.
var ErrCleanupClosed = lifecycle.ErrCleanupClosed
