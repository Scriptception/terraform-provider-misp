package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccUserResource_basic(t *testing.T) {
	prefix := acctestResourceName("tfacc")
	email := fmt.Sprintf("%s@example.com", prefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccUserConfig(email, "1", "3", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_user.test", "email", email),
					resource.TestCheckResourceAttr("misp_user.test", "org_id", "1"),
					resource.TestCheckResourceAttr("misp_user.test", "role_id", "3"),
					resource.TestCheckResourceAttr("misp_user.test", "disabled", "true"),
					resource.TestCheckResourceAttrSet("misp_user.test", "id"),
				),
			},
			// Toggle disabled in-place.
			{
				Config: testAccUserConfig(email, "1", "3", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_user.test", "disabled", "false"),
				),
			},
			{
				ResourceName:      "misp_user.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccUserConfig(email, orgID, roleID string, disabled bool) string {
	return fmt.Sprintf(`
provider "misp" { insecure = true }

resource "misp_user" "test" {
  email    = %q
  org_id   = %q
  role_id  = %q
  disabled = %t
}
`, email, orgID, roleID, disabled)
}
