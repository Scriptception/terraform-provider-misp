package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSharingGroupResource_basic(t *testing.T) {
	name := acctestResourceName("tf-acc-sg")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSharingGroupConfig(name, "initial desc", "For internal use only"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_sharing_group.test", "name", name),
					resource.TestCheckResourceAttr("misp_sharing_group.test", "description", "initial desc"),
					resource.TestCheckResourceAttr("misp_sharing_group.test", "releasability", "For internal use only"),
					resource.TestCheckResourceAttrSet("misp_sharing_group.test", "id"),
					resource.TestCheckResourceAttrSet("misp_sharing_group.test", "uuid"),
					resource.TestCheckResourceAttrSet("misp_sharing_group.test", "org_id"),
				),
			},
			{
				Config: testAccSharingGroupConfig(name, "updated desc", "Releasable to partners"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_sharing_group.test", "description", "updated desc"),
					resource.TestCheckResourceAttr("misp_sharing_group.test", "releasability", "Releasable to partners"),
				),
			},
			{
				ResourceName:      "misp_sharing_group.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccSharingGroupConfig(name, desc, rel string) string {
	return fmt.Sprintf(`
provider "misp" { insecure = true }

resource "misp_sharing_group" "test" {
  name          = %q
  description   = %q
  releasability = %q
  active        = true
  local         = false
  roaming       = false
}
`, name, desc, rel)
}
