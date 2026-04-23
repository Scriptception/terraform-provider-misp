package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// protoV6ProviderFactories instantiates the provider once per test.
var protoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"misp": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck skips acceptance tests unless a live MISP instance is configured.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("TF_ACC not set; skipping acceptance test")
	}
	for _, v := range []string{"MISP_URL", "MISP_API_KEY"} {
		if os.Getenv(v) == "" {
			t.Fatalf("acceptance test requires %s to be set", v)
		}
	}
}
