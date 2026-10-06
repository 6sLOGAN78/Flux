// Package templates owns email templates embedded into the worker binary.
package templates

import "embed"

// Assets contains the authored HTML exports used for safe email rendering.
//
//go:embed emails/*.html
var Assets embed.FS
