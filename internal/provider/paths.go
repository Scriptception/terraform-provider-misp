package provider

import "github.com/hashicorp/terraform-plugin-framework/path"

// pathRoot is a small wrapper to keep provider.go readable.
func pathRoot(name string) path.Path {
	return path.Root(name)
}
