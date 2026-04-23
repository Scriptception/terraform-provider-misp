package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSharingGroupServerResource_basic(t *testing.T) {
	serverName := acctestResourceName("tf-acc-sgs-srv")
	sgName := acctestResourceName("tf-acc-sgs-sg")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create the server-to-sharing-group link and verify it exists.
			{
				Config: testAccSharingGroupServerConfig(serverName, sgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("misp_sharing_group_server.link", "id"),
					resource.TestCheckResourceAttrSet("misp_sharing_group_server.link", "sharing_group_id"),
					resource.TestCheckResourceAttrSet("misp_sharing_group_server.link", "server_id"),
					// Verify the composite id contains a colon separator.
					resource.TestMatchResourceAttr("misp_sharing_group_server.link", "id",
						regexp.MustCompile(`.+:.+`)),
				),
			},
			// Import the resource using the composite id <sg_id>:<server_id>.
			{
				ResourceName:      "misp_sharing_group_server.link",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["misp_sharing_group_server.link"]
					if !ok {
						return "", fmt.Errorf("resource misp_sharing_group_server.link not found in state")
					}
					return rs.Primary.ID, nil
				},
			},
		},
	})
}

func testAccSharingGroupServerConfig(serverName, sgName string) string {
	return fmt.Sprintf(`
provider "misp" { insecure = true }

resource "misp_server" "test" {
  name          = %q
  url           = "https://misp-peer.invalid"
  authkey       = "0123456789abcdef0123456789abcdef01234567"
  remote_org_id = "1"
  self_signed   = true
}

resource "misp_sharing_group" "test" {
  name          = %q
  releasability = "test"
  active        = true
}

resource "misp_sharing_group_server" "link" {
  sharing_group_id = misp_sharing_group.test.id
  server_id        = misp_server.test.id
}
`, serverName, sgName)
}
