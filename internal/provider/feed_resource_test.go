package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccFeedResource_basic(t *testing.T) {
	name := acctestResourceName("tf-acc-feed")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccFeedConfig(name, "InitialProvider", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_feed.test", "name", name),
					resource.TestCheckResourceAttr("misp_feed.test", "provider_name", "InitialProvider"),
					resource.TestCheckResourceAttr("misp_feed.test", "url", "https://example.com/feed.json"),
					resource.TestCheckResourceAttr("misp_feed.test", "enabled", "false"),
					resource.TestCheckResourceAttrSet("misp_feed.test", "id"),
				),
			},
			{
				Config: testAccFeedConfig(name, "UpdatedProvider", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_feed.test", "provider_name", "UpdatedProvider"),
					resource.TestCheckResourceAttr("misp_feed.test", "enabled", "true"),
				),
			},
			{
				ResourceName:      "misp_feed.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccFeedConfig(name, providerName string, enabled bool) string {
	return fmt.Sprintf(`
provider "misp" { insecure = true }

resource "misp_feed" "test" {
  name          = %q
  provider_name = %q
  url           = "https://example.com/feed.json"
  source_format = "misp"
  enabled       = %t
}
`, name, providerName, enabled)
}
