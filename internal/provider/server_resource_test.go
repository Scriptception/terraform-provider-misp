package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccServerResource_basic(t *testing.T) {
	name := acctestResourceName("tf-acc-srv")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and verify.
			{
				Config: testAccServerConfig(name, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_server.test", "name", name),
					resource.TestCheckResourceAttr("misp_server.test", "url", "https://misp-peer.invalid"),
					resource.TestCheckResourceAttr("misp_server.test", "push", "false"),
					resource.TestCheckResourceAttr("misp_server.test", "pull", "false"),
					resource.TestCheckResourceAttr("misp_server.test", "self_signed", "true"),
					resource.TestCheckResourceAttrSet("misp_server.test", "id"),
				),
			},
			// Update: flip push to true.
			{
				Config: testAccServerConfig(name, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_server.test", "push", "true"),
				),
			},
			// Import: authkey is write-only and not returned by MISP on read,
			// so we exclude it from the verify check.
			{
				ResourceName:            "misp_server.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"authkey"},
			},
		},
	})
}

func testAccServerConfig(name string, push bool) string {
	return fmt.Sprintf(`
provider "misp" { insecure = true }

resource "misp_server" "test" {
  name          = %q
  url           = "https://misp-peer.invalid"
  authkey       = "0123456789abcdef0123456789abcdef01234567"
  remote_org_id = "1"
  push          = %t
  pull          = false
  self_signed   = true
}
`, name, push)
}
