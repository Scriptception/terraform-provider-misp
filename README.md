# Terraform Provider for MISP

A [Terraform](https://www.terraform.io) provider for [MISP](https://www.misp-project.org/) — the open source threat intelligence and sharing platform.

Manage MISP organisations, tags, sharing groups, users, and other resources as code.

> **Status:** feature-complete for v0.4. Fourteen resources cover the MISP configuration surface:
> `misp_organisation`, `misp_tag`, `misp_sharing_group`, `misp_sharing_group_member`, `misp_sharing_group_server`, `misp_user`, `misp_taxonomy`, `misp_feed`, `misp_server`, `misp_role`, `misp_warninglist`, `misp_noticelist`, `misp_setting`, `misp_galaxy_cluster`.
>
> Scope is strictly MISP **configuration** — events, attributes, sightings, and other operational data are deliberately out of scope and will never be added. Use PyMISP or the MISP UI for analyst workflows.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.6
- [Go](https://go.dev/doc/install) >= 1.24 (to build from source)
- A running MISP instance and an API key

## Usage

```hcl
terraform {
  required_providers {
    misp = {
      source  = "Scriptception/misp"
      version = "~> 0.1"
    }
  }
}

provider "misp" {
  url     = "https://misp.example.com"
  api_key = var.misp_api_key
}

resource "misp_organisation" "acme" {
  name        = "ACME"
  description = "ACME Corp threat intel team"
  local       = true
}

resource "misp_tag" "tlp_amber" {
  name   = "tlp:amber"
  colour = "#FFC000"
}
```

Provider configuration can also be supplied via environment variables: `MISP_URL`, `MISP_API_KEY`, `MISP_INSECURE`.

## Development

```sh
make build      # compile
make test       # unit tests
make testacc    # acceptance tests (requires a live MISP instance)
make generate   # regenerate docs from schema + examples/
make lint       # golangci-lint
```

## License

Mozilla Public License 2.0. See [LICENSE](./LICENSE).
