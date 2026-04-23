package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccWarninglistResource_basic uses the bundled "List of known Akamai IP ranges" warninglist.
//
// Chosen because it ships with every MISP instance (id=1 on a stock install)
// but is disabled by default — so enabling it during the test and disabling it
// on destroy leaves the instance in its original state. Do NOT switch this to a
// warninglist that a real operator is likely to have enabled already; the test
// would leave that warninglist disabled when it finishes.
const warninglistTestName = "List of known Akamai IP ranges"

func TestAccWarninglistResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Enable the warninglist.
			{
				Config: fmt.Sprintf(`
provider "misp" { insecure = true }

resource "misp_warninglist" "under_test" {
  name    = %q
  enabled = true
}
`, warninglistTestName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_warninglist.under_test", "name", warninglistTestName),
					resource.TestCheckResourceAttr("misp_warninglist.under_test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("misp_warninglist.under_test", "id"),
					resource.TestCheckResourceAttrSet("misp_warninglist.under_test", "version"),
					resource.TestCheckResourceAttrSet("misp_warninglist.under_test", "type"),
				),
			},
			// Toggle off.
			{
				Config: fmt.Sprintf(`
provider "misp" { insecure = true }

resource "misp_warninglist" "under_test" {
  name    = %q
  enabled = false
}
`, warninglistTestName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_warninglist.under_test", "enabled", "false"),
				),
			},
			// Import by numeric id; name isn't preserved through import-by-id so
			// skip its verification (same workaround used for taxonomy's namespace).
			{
				ResourceName:            "misp_warninglist.under_test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name"},
			},
		},
	})
}
