package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccTagResource_basic(t *testing.T) {
	name := acctestResourceName("tf:acc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTagConfig(name, "#112233", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_tag.test", "name", name),
					resource.TestCheckResourceAttr("misp_tag.test", "colour", "#112233"),
					resource.TestCheckResourceAttr("misp_tag.test", "exportable", "true"),
					resource.TestCheckResourceAttrSet("misp_tag.test", "id"),
				),
			},
			// Change colour and exportable; in-place update.
			{
				Config: testAccTagConfig(name, "#445566", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_tag.test", "colour", "#445566"),
					resource.TestCheckResourceAttr("misp_tag.test", "exportable", "false"),
				),
			},
			// Import by id.
			{
				ResourceName:      "misp_tag.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccTagConfig(name, colour string, exportable bool) string {
	return fmt.Sprintf(`
provider "misp" {
  insecure = true
}

resource "misp_tag" "test" {
  name       = %q
  colour     = %q
  exportable = %t
}
`, name, colour, exportable)
}
