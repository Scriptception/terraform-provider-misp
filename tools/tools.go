//go:build tools

// Package tools anchors tool-only dependencies so go.mod tracks them for
// reproducible `go generate` runs. This package is never imported from the
// provider binary.
package tools

import (
	// tfplugindocs generates docs/ from schema + examples/.
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-name misp
