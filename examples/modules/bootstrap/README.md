# MISP bootstrap module

A reusable Terraform module that produces a sensible baseline configuration for a freshly-installed MISP instance: a local organisation, an initial admin user, a set of enabled taxonomies/warninglists/noticelists, custom tags, and any global settings you want enforced.

Apply this once against a clean MISP install and you have a working baseline in source control, ready to be extended with sharing groups, feeds, and additional users.

## Usage

```hcl
provider "misp" {
  url     = "https://misp.example.com"
  api_key = var.misp_api_key
}

module "misp_baseline" {
  source = "../../examples/modules/bootstrap"   # or your preferred reference

  local_org = {
    name        = "ACME"
    description = "ACME Corp threat intel team"
    type        = "CSIRT"
    nationality = "AU"
    sector      = "Technology"
  }

  admin_email = "soc-admin@acme.example"

  enable_taxonomies = [
    "tlp",
    "admiralty-scale",
    "workflow",
    "kill-chain",
  ]

  enable_warninglists = [
    "List of known google domains",
    "List of known Cloudflare IP ranges",
  ]

  enable_noticelists = [
    "gdpr",
  ]

  custom_tags = {
    internal = {
      name       = "source:internal"
      colour     = "#1f77b4"
      exportable = false
    }
    automated = {
      name       = "workflow:automated"
      colour     = "#888888"
      exportable = true
    }
  }

  settings = {
    "MISP.baseurl" = "https://misp.example.com"
  }
}

# Reference module outputs to extend the baseline:
resource "misp_sharing_group" "partners" {
  name   = "Partner network"
  org_id = module.misp_baseline.local_organisation_id
}
```

## Inputs

| Name                  | Type                  | Default                                     | Description                                                |
|-----------------------|-----------------------|---------------------------------------------|------------------------------------------------------------|
| `local_org`           | object                | —                                           | The local organisation definition (name, type, nationality, sector). |
| `admin_email`         | string                | —                                           | Email of the initial admin user.                            |
| `admin_role_id`       | string                | `"1"`                                       | MISP role id for the admin user (1 = site-admin).           |
| `enable_taxonomies`   | list(string)          | `["tlp", "admiralty-scale", "workflow"]`    | Bundled taxonomy namespaces to enable.                      |
| `enable_warninglists` | list(string)          | `[]`                                        | Bundled warninglist names to enable (exact match).          |
| `enable_noticelists`  | list(string)          | `[]`                                        | Bundled noticelist names to enable (exact match).           |
| `custom_tags`         | map(object)           | `{}`                                        | Custom tags to create in the tag catalog.                   |
| `settings`            | map(string)           | `{}`                                        | Global MISP settings to enforce.                            |

## Outputs

| Name                      | Description                                                           |
|---------------------------|-----------------------------------------------------------------------|
| `local_organisation_id`   | Numeric MISP id of the local organisation.                             |
| `local_organisation_uuid` | UUID of the local organisation.                                        |
| `admin_user_id`           | Numeric MISP id of the admin user.                                     |
| `custom_tag_ids`          | Map from tag key (in `custom_tags`) to the tag's numeric MISP id.      |

## Notes

- The taxonomy / warninglist / noticelist resources use the **adopt-existing** pattern: MISP ships these disabled, and the resources toggle `enabled=true`. `terraform destroy` re-disables them, returning the instance to its original state — it never removes bundled entries.
- Warninglist and noticelist names must match MISP's exact display name (including quotes, apostrophes, and capitalisation). Browse MISP → Event Actions → List Warninglists to get the canonical names.
- `admin_role_id = "1"` is the standard site-admin role shipped with every MISP. If you replaced the default roles with custom ones via `misp_role`, pass that role's id instead.
