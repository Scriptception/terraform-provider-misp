package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// noticelistTestName is the short name of the noticelist used in acceptance
// tests. "gdpr" is the only noticelist that ships on a stock MISP 2.5.36
// install (id=1). It is disabled by default, so enabling it during the test
// and disabling it on destroy leaves the instance in its original state.
// Do NOT switch this to a noticelist that a real operator is likely to have
// enabled already; the test would leave that noticelist disabled when it
// finishes.
const noticelistTestName = "gdpr"

func TestAccNoticelistResource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Enable the noticelist.
			{
				Config: fmt.Sprintf(`
provider "misp" { insecure = true }

resource "misp_noticelist" "under_test" {
  name    = %q
  enabled = true
}
`, noticelistTestName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_noticelist.under_test", "name", noticelistTestName),
					resource.TestCheckResourceAttr("misp_noticelist.under_test", "enabled", "true"),
					resource.TestCheckResourceAttrSet("misp_noticelist.under_test", "id"),
					resource.TestCheckResourceAttrSet("misp_noticelist.under_test", "version"),
					resource.TestCheckResourceAttrSet("misp_noticelist.under_test", "expanded_name"),
				),
			},
			// Toggle off.
			{
				Config: fmt.Sprintf(`
provider "misp" { insecure = true }

resource "misp_noticelist" "under_test" {
  name    = %q
  enabled = false
}
`, noticelistTestName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_noticelist.under_test", "enabled", "false"),
				),
			},
			// Import by numeric id; name isn't preserved through import-by-id so
			// skip its verification (same workaround used for taxonomy's namespace
			// and warninglist's name).
			{
				ResourceName:            "misp_noticelist.under_test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name"},
			},
		},
	})
}
