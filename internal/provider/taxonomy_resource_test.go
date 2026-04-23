package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccTaxonomyResource_basic uses the bundled "admiralty-scale" taxonomy.
//
// Chosen because it ships with MISP but is disabled by default on a fresh
// install — so enabling it during the test and disabling it on destroy leaves
// the instance in its original state. Do NOT switch this to a taxonomy a real
// operator is likely to have enabled (tlp, misp, etc) — the test would leave
// canonical taxonomies disabled.
const taxonomyTestNamespace = "admiralty-scale"

func TestAccTaxonomyResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "misp" { insecure = true }

resource "misp_taxonomy" "under_test" {
  namespace = %q
  enabled   = true
}
`, taxonomyTestNamespace),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_taxonomy.under_test", "namespace", taxonomyTestNamespace),
					resource.TestCheckResourceAttr("misp_taxonomy.under_test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("misp_taxonomy.under_test", "id"),
					resource.TestCheckResourceAttrSet("misp_taxonomy.under_test", "version"),
				),
			},
			// Toggle off.
			{
				Config: fmt.Sprintf(`
provider "misp" { insecure = true }

resource "misp_taxonomy" "under_test" {
  namespace = %q
  enabled   = false
}
`, taxonomyTestNamespace),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_taxonomy.under_test", "enabled", "false"),
				),
			},
			{
				ResourceName:      "misp_taxonomy.under_test",
				ImportState:       true,
				ImportStateVerify: true,
				// namespace isn't carried through import-by-id; skip its verification.
				ImportStateVerifyIgnore: []string{"namespace"},
			},
		},
	})
}
