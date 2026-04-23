package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccGalaxyClusterResource_basic exercises create, update, and import for a
// custom galaxy cluster.
//
// Prerequisites: a stock MISP install with the default galaxies loaded.
// Galaxy id 1 is the "360net-threat-actor" galaxy present on every default MISP
// install. If that galaxy is absent the test will fail with an API error.
func TestAccGalaxyClusterResource_basic(t *testing.T) {
	name := acctestResourceName("tf-acc-gc")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with elements (two entries share key "refs").
			{
				Config: testAccGalaxyClusterConfig(name, "initial description", "https://example.com",
					[]string{"terraform-provider-misp"},
					[][2]string{
						{"refs", "https://example.com/ref1"},
						{"refs", "https://example.com/ref2"},
						{"synonyms", "InitialAlias"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_galaxy_cluster.test", "galaxy_id", "1"),
					resource.TestCheckResourceAttr("misp_galaxy_cluster.test", "value", name),
					resource.TestCheckResourceAttr("misp_galaxy_cluster.test", "description", "initial description"),
					resource.TestCheckResourceAttr("misp_galaxy_cluster.test", "source", "https://example.com"),
					resource.TestCheckResourceAttr("misp_galaxy_cluster.test", "authors.#", "1"),
					resource.TestCheckResourceAttr("misp_galaxy_cluster.test", "authors.0", "terraform-provider-misp"),
					resource.TestCheckResourceAttr("misp_galaxy_cluster.test", "elements.#", "3"),
					resource.TestCheckResourceAttrSet("misp_galaxy_cluster.test", "id"),
					resource.TestCheckResourceAttrSet("misp_galaxy_cluster.test", "uuid"),
					resource.TestCheckResourceAttrSet("misp_galaxy_cluster.test", "tag_name"),
					resource.TestCheckResourceAttrSet("misp_galaxy_cluster.test", "version"),
					resource.TestCheckResourceAttrSet("misp_galaxy_cluster.test", "type"),
				),
			},
			// Step 2: Update description and swap out elements.
			{
				Config: testAccGalaxyClusterConfig(name, "updated description", "https://example.com",
					[]string{"terraform-provider-misp"},
					[][2]string{
						{"refs", "https://example.com/ref-updated"},
						{"synonyms", "UpdatedAlias"},
					},
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_galaxy_cluster.test", "description", "updated description"),
					resource.TestCheckResourceAttr("misp_galaxy_cluster.test", "elements.#", "2"),
				),
			},
			// Step 3: Import by id.
			// TODO: element ordering after import may not match plan order, causing
			// ImportStateVerify failures. If this becomes flaky, add
			//   ImportStateVerifyIgnore: []string{"elements"}
			{
				ResourceName:      "misp_galaxy_cluster.test",
				ImportState:       true,
				ImportStateVerify: true,
				// elements ordering after import may differ from plan order.
				ImportStateVerifyIgnore: []string{"elements"},
			},
		},
	})
}

// testAccGalaxyClusterConfig renders a misp_galaxy_cluster resource config.
// elements is a slice of [2]string{key, value} pairs.
//
// elements is a ListNestedAttribute on the framework side, so it uses HCL
// attribute-list syntax (`elements = [ {...}, {...} ]`), NOT block syntax.
func testAccGalaxyClusterConfig(name, description, source string, authors []string, elements [][2]string) string {
	authorsList := ""
	for _, a := range authors {
		authorsList += fmt.Sprintf("    %q,\n", a)
	}

	elementsList := ""
	for _, e := range elements {
		elementsList += fmt.Sprintf("    { key = %q, value = %q },\n", e[0], e[1])
	}

	return fmt.Sprintf(`
provider "misp" {
  insecure = true
}

resource "misp_galaxy_cluster" "test" {
  galaxy_id    = "1"
  value        = %q
  description  = %q
  source       = %q
  distribution = "1"
  authors = [
%s  ]
  elements = [
%s  ]
}
`, name, description, source, authorsList, elementsList)
}
