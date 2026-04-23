package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccOrganisationResource_basic(t *testing.T) {
	name := acctestResourceName("tf-acc-org")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccOrganisationConfig(name, "initial description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_organisation.test", "name", name),
					resource.TestCheckResourceAttr("misp_organisation.test", "description", "initial description"),
					resource.TestCheckResourceAttr("misp_organisation.test", "local", "true"),
					resource.TestCheckResourceAttrSet("misp_organisation.test", "id"),
					resource.TestCheckResourceAttrSet("misp_organisation.test", "uuid"),
				),
			},
			// Update-in-place: description changes, id/uuid must not.
			{
				Config: testAccOrganisationConfig(name, "updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_organisation.test", "description", "updated description"),
				),
			},
			// Import by id.
			{
				ResourceName:      "misp_organisation.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccOrganisationConfig(name, desc string) string {
	return fmt.Sprintf(`
provider "misp" {
  insecure = true
}

resource "misp_organisation" "test" {
  name        = %q
  description = %q
  local       = true
}
`, name, desc)
}

func TestAccOrganisationResource_restrictedToDomain(t *testing.T) {
	name := acctestResourceName("tf-acc-org-rtd")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: protoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Create with empty restricted_to_domain.
			{
				Config: testAccOrganisationRestrictedConfig(name, `[]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_organisation.test_rtd", "name", name),
					resource.TestCheckResourceAttr("misp_organisation.test_rtd", "restricted_to_domain.#", "0"),
				),
			},
			// Step 2: Update to a single domain.
			{
				Config: testAccOrganisationRestrictedConfig(name, `["example.com"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_organisation.test_rtd", "restricted_to_domain.#", "1"),
					resource.TestCheckResourceAttr("misp_organisation.test_rtd", "restricted_to_domain.0", "example.com"),
				),
			},
			// Step 3: Update to two domains.
			{
				Config: testAccOrganisationRestrictedConfig(name, `["example.com", "corp.example.com"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_organisation.test_rtd", "restricted_to_domain.#", "2"),
					resource.TestCheckResourceAttr("misp_organisation.test_rtd", "restricted_to_domain.0", "example.com"),
					resource.TestCheckResourceAttr("misp_organisation.test_rtd", "restricted_to_domain.1", "corp.example.com"),
				),
			},
			// Step 4: Update back to empty.
			{
				Config: testAccOrganisationRestrictedConfig(name, `[]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("misp_organisation.test_rtd", "restricted_to_domain.#", "0"),
				),
			},
		},
	})
}

func testAccOrganisationRestrictedConfig(name, domains string) string {
	return fmt.Sprintf(`
provider "misp" {
  insecure = true
}

resource "misp_organisation" "test_rtd" {
  name                 = %q
  local                = true
  restricted_to_domain = %s
}
`, name, domains)
}
