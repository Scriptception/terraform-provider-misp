// Command terraform-provider-misp is the Terraform provider for MISP.
//
// Run with -debug to attach a debugger (e.g. delve). Run with -version to
// print build info and exit. Normally the provider is invoked by Terraform
// itself, not directly.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/Scriptception/terraform-provider-misp/internal/provider"
)

// Set at build time via -ldflags (see .goreleaser.yml).
var (
	version = "dev"
	commit  = "none"
)

func main() {
	var (
		debug       bool
		showVersion bool
	)

	flag.BoolVar(&debug, "debug", false, "run the provider with support for debuggers like delve")
	flag.BoolVar(&showVersion, "version", false, "print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Printf("terraform-provider-misp %s (commit %s)\n", version, commit)
		return
	}

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/Scriptception/misp",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New(version), opts); err != nil {
		log.Fatal(err.Error())
	}
}
