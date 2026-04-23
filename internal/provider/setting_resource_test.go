package provider

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/Scriptception/terraform-provider-misp/internal/client"
)

// settingTestName is the MISP setting used for acceptance tests.
//
// MISP.welcome_text_top is a display-only string shown on the login page
// before the MISP logo. It defaults to an empty string and changing it has no
// operational impact on the MISP instance — making it safe to toggle during
// tests.
const settingTestName = "MISP.welcome_text_top"

// settingTestOriginalValue holds the value read from MISP before the test
// runs so that the final step can restore it.
var settingTestOriginalValue string

func TestAccSettingResource_basic(t *testing.T) {
	testAccPreCheck(t)

	// Capture the current value so we can restore it in the final step.
	c, err := client.New(client.Config{
		URL:      os.Getenv("MISP_URL"),
		APIKey:   os.Getenv("MISP_API_KEY"),
		Insecure: os.Getenv("MISP_INSECURE") == "true",
	})
	if err != nil {
		t.Fatalf("creating client for PreCheck: %v", err)
	}
	original, err := c.GetSetting(context.Background(), settingTestName)
	if err != nil {
		t.Fatalf("reading original setting value: %v", err)
	}
	settingTestOriginalValue = original.Value.String()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create (set a test value).
			{
				Config: testAccSettingConfig("terraform-provider-misp-acceptance-test"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_setting.under_test", "name", settingTestName),
					resource.TestCheckResourceAttr("misp_setting.under_test", "value", "terraform-provider-misp-acceptance-test"),
					resource.TestCheckResourceAttr("misp_setting.under_test", "id", settingTestName),
					resource.TestCheckResourceAttrSet("misp_setting.under_test", "type"),
					resource.TestCheckResourceAttrSet("misp_setting.under_test", "description"),
				),
			},
			// Step 2: Update (change to a different value).
			{
				Config: testAccSettingConfig("terraform-provider-misp-acceptance-test-v2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_setting.under_test", "value", "terraform-provider-misp-acceptance-test-v2"),
				),
			},
			// Step 3: Import by name.
			{
				ResourceName:      "misp_setting.under_test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     settingTestName,
			},
			// Step 4: Restore original value so the MISP instance is left clean.
			{
				Config: testAccSettingConfig(settingTestOriginalValue),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_setting.under_test", "value", settingTestOriginalValue),
				),
			},
		},
	})
}

func testAccSettingConfig(value string) string {
	return fmt.Sprintf(`
provider "misp" { insecure = true }

resource "misp_setting" "under_test" {
  name  = %q
  value = %q
}
`, settingTestName, value)
}
