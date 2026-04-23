package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccSharingGroupMemberResource_basic(t *testing.T) {
	orgName := acctestResourceName("tf-acc-sgm-org")
	sgName := acctestResourceName("tf-acc-sgm-sg")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create the membership and verify it exists.
			{
				Config: testAccSharingGroupMemberConfig(orgName, sgName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("misp_sharing_group_member.link", "id"),
					resource.TestCheckResourceAttrSet("misp_sharing_group_member.link", "sharing_group_id"),
					resource.TestCheckResourceAttrSet("misp_sharing_group_member.link", "organisation_id"),
					// Verify the composite id contains a colon separator.
					resource.TestMatchResourceAttr("misp_sharing_group_member.link", "id",
						regexp.MustCompile(`.+:.+`)),
				),
			},
			// Import the resource using the composite id <sg_id>:<org_id>.
			{
				ResourceName:      "misp_sharing_group_member.link",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					rs, ok := s.RootModule().Resources["misp_sharing_group_member.link"]
					if !ok {
						return "", fmt.Errorf("resource misp_sharing_group_member.link not found in state")
					}
					return rs.Primary.ID, nil
				},
			},
		},
	})
}

func testAccSharingGroupMemberConfig(orgName, sgName string) string {
	return fmt.Sprintf(`
provider "misp" { insecure = true }

resource "misp_organisation" "member_org" {
  name  = %q
  local = true
}

resource "misp_sharing_group" "sg" {
  name          = %q
  releasability = "test"
  active        = true
}

resource "misp_sharing_group_member" "link" {
  sharing_group_id = misp_sharing_group.sg.id
  organisation_id  = misp_organisation.member_org.id
}
`, orgName, sgName)
}
