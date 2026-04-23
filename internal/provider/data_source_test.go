package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganisationDataSource_basic(t *testing.T) {
	name := acctestResourceName("tf-acc-ds-org")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "misp" { insecure = true }

resource "misp_organisation" "seed" {
  name  = %q
  local = true
}

data "misp_organisation" "by_id" {
  id = misp_organisation.seed.id
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.misp_organisation.by_id", "name", name),
					resource.TestCheckResourceAttrSet("data.misp_organisation.by_id", "uuid"),
				),
			},
		},
	})
}

func TestAccTagDataSource_basic(t *testing.T) {
	name := acctestResourceName("tf:ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "misp" { insecure = true }

resource "misp_tag" "seed" {
  name   = %q
  colour = "#ABCDEF"
}

data "misp_tag" "by_id" {
  id = misp_tag.seed.id
}
`, name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.misp_tag.by_id", "name", name),
					resource.TestCheckResourceAttr("data.misp_tag.by_id", "colour", "#ABCDEF"),
				),
			},
		},
	})
}
