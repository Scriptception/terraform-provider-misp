package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccRoleResource_basic(t *testing.T) {
	name := acctestResourceName("tf-acc-role")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with permission=1, perm_tagger=true
			{
				Config: testAccRoleConfig(name, "1", true, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_role.test", "name", name),
					resource.TestCheckResourceAttr("misp_role.test", "permission", "1"),
					resource.TestCheckResourceAttr("misp_role.test", "perm_tagger", "true"),
					resource.TestCheckResourceAttr("misp_role.test", "perm_sharing_group", "false"),
					// Empirically: permission=1 derives perm_add + perm_modify.
					resource.TestCheckResourceAttr("misp_role.test", "perm_add", "true"),
					resource.TestCheckResourceAttr("misp_role.test", "perm_modify", "true"),
					resource.TestCheckResourceAttr("misp_role.test", "perm_modify_org", "false"),
					resource.TestCheckResourceAttr("misp_role.test", "perm_publish", "false"),
					resource.TestCheckResourceAttr("misp_role.test", "perm_delegate", "false"),
					resource.TestCheckResourceAttrSet("misp_role.test", "id"),
				),
			},
			// Step 2: Update to permission=2, flip perm_tagger
			{
				Config: testAccRoleConfig(name, "2", false, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_role.test", "permission", "2"),
					resource.TestCheckResourceAttr("misp_role.test", "perm_tagger", "false"),
					resource.TestCheckResourceAttr("misp_role.test", "perm_sharing_group", "true"),
					// permission=2 derives perm_add, perm_modify, perm_modify_org=true
					resource.TestCheckResourceAttr("misp_role.test", "perm_add", "true"),
					resource.TestCheckResourceAttr("misp_role.test", "perm_modify", "true"),
					resource.TestCheckResourceAttr("misp_role.test", "perm_modify_org", "true"),
					resource.TestCheckResourceAttr("misp_role.test", "perm_publish", "false"),
					resource.TestCheckResourceAttr("misp_role.test", "perm_delegate", "false"),
				),
			},
			// Step 3: Import
			{
				ResourceName:      "misp_role.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccRoleConfig(name, permission string, permTagger, permSharingGroup bool) string {
	return fmt.Sprintf(`
provider "misp" { insecure = true }

resource "misp_role" "test" {
  name               = %q
  permission         = %q
  perm_tagger        = %t
  perm_sharing_group = %t
}
`, name, permission, permTagger, permSharingGroup)
}
