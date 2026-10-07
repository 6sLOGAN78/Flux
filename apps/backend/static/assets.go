// Package static owns documentation assets embedded into the API binary.
package static

import "embed"

// Assets contains the UI and the generated, canonical OpenAPI document.
//
//go:embed openapi.html openapi.json
var Assets embed.FS

// DocumentationCSP permits only the integrity-pinned standalone Scalar script.
// Scalar requires inline styles; scripts and connections remain constrained.
const DocumentationCSP = "default-src 'none'; " +
	"script-src 'sha384-OKyMdsDX84ypSZEhVun8YElXk5c2GQaH3EXPOc6ItmVcLDUAvKHYwvDLvAgsqVtB'; " +
	"style-src 'unsafe-inline'; img-src 'self' data:; font-src 'self'; " +
	"connect-src 'self'; object-src 'none'; base-uri 'none'; " +
	"frame-ancestors 'none'; form-action 'self'"
