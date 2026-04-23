# Terraform Provider for MISP

[![Tests](https://github.com/Scriptception/terraform-provider-misp/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/Scriptception/terraform-provider-misp/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/Scriptception/terraform-provider-misp.svg)](https://pkg.go.dev/github.com/Scriptception/terraform-provider-misp)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL_2.0-brightgreen.svg)](./LICENSE)
[![Go version](https://img.shields.io/github/go-mod/go-version/Scriptception/terraform-provider-misp)](./go.mod)

Manage a [MISP](https://www.misp-project.org/) instance — the open-source threat intelligence and sharing platform — as code.

- **14 resources** covering organisations, users, roles, tags, taxonomies, sharing groups (with org and server membership), feeds, sync servers, warninglists, noticelists, global settings, and custom galaxy clusters.
- **12 data sources** for reading existing configuration.
- Built on [terraform-plugin-framework](https://developer.hashicorp.com/terraform/plugin/framework) (protocol v6).
- Tested against MISP 2.5.x.

> **Scope.** This provider manages MISP **configuration** — the long-lived administrative state that an operator maintains in git. Events, attributes, sightings, analyst notes, correlations, and other operational data are deliberately out of scope and will not be added. Use [PyMISP](https://github.com/MISP/PyMISP) or the MISP UI for analyst workflows. See [CLAUDE.md](./CLAUDE.md) for the full scope rationale.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) ≥ 1.6
- A running MISP instance (2.5.x tested) and an API key with admin scope
- [Go](https://go.dev/doc/install) ≥ 1.24 only if building from source

## Quick start

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
  api_key = var.misp_api_key   # or set MISP_API_KEY
  # insecure = true            # for self-signed certs; or set MISP_INSECURE=true
}
```

Provider settings also accept environment variables: `MISP_URL`, `MISP_API_KEY`, `MISP_INSECURE`.

## Walkthrough

A realistic slice of what you might declare for a small threat-sharing setup: a local org, a partner org, a custom tag, a sharing group joining the two, a user, and a CIRCL OSINT feed subscription.

```hcl
# 1. Your own organisation (the "local" one — owns events created on this instance).
resource "misp_organisation" "acme" {
  name        = "ACME"
  description = "ACME Corp threat intel team"
  type        = "CSIRT"
  nationality = "AU"
  local       = true
}

# 2. A partner org you share with (remote = local=false).
resource "misp_organisation" "partner" {
  name        = "Partner Inc"
  description = "External threat-sharing partner"
  local       = false
}

# 3. A custom tag for marking internally-sourced indicators.
resource "misp_tag" "internal" {
  name       = "source:internal"
  colour     = "#1f77b4"
  exportable = false
}

# 4. A sharing group that includes both orgs.
resource "misp_sharing_group" "partner_network" {
  name          = "ACME ↔ Partner"
  description   = "Bilateral sharing with Partner Inc"
  releasability = "Amber — partners only"
  org_id        = misp_organisation.acme.id
}

resource "misp_sharing_group_member" "acme_in_sg" {
  sharing_group_id = misp_sharing_group.partner_network.id
  organisation_id  = misp_organisation.acme.id
}

resource "misp_sharing_group_member" "partner_in_sg" {
  sharing_group_id = misp_sharing_group.partner_network.id
  organisation_id  = misp_organisation.partner.id
}

# 5. A human analyst account.
resource "misp_user" "alice" {
  email   = "alice@acme.example"
  org_id  = misp_organisation.acme.id
  role_id = "3"   # "User" — standard analyst role shipped by MISP
}

# 6. Subscribe to the CIRCL OSINT feed.
resource "misp_feed" "circl_osint" {
  name          = "CIRCL OSINT feed"
  provider_name = "CIRCL"
  url           = "https://www.circl.lu/doc/misp/feed-osint"
  source_format = "misp"
  enabled       = true
  caching_enabled = true
}
```

Apply that against a fresh MISP and you have a working cross-org sharing setup entirely in source control.

See [`examples/modules/bootstrap/`](./examples/modules/bootstrap) for a larger, runnable end-to-end module that you can use as a starting point for a new MISP instance.

## Resources and data sources

Full reference docs live under [`docs/`](./docs) and on the Terraform Registry once published. Quick index:

| Resource                        | What it manages                                                                          |
|---------------------------------|------------------------------------------------------------------------------------------|
| `misp_organisation`             | Orgs — the primary unit of ownership and access control.                                  |
| `misp_user`                     | User accounts (admin, org admin, sync user, analysts).                                    |
| `misp_role`                     | Custom permission roles.                                                                  |
| `misp_tag`                      | Tag catalog entries.                                                                      |
| `misp_taxonomy`                 | Enable/disable bundled taxonomies by namespace.                                           |
| `misp_warninglist`              | Enable/disable bundled warninglists by name.                                              |
| `misp_noticelist`               | Enable/disable bundled noticelists by name.                                               |
| `misp_sharing_group`            | Sharing group definitions.                                                                |
| `misp_sharing_group_member`     | Organisation-in-sharing-group junction.                                                   |
| `misp_sharing_group_server`     | Server-in-sharing-group junction.                                                         |
| `misp_feed`                     | Feed subscriptions.                                                                       |
| `misp_server`                   | Sync server peers.                                                                        |
| `misp_setting`                  | Global instance settings (`MISP.baseurl`, `Security.*`, …).                                |
| `misp_galaxy_cluster`           | Custom galaxy clusters (bundled clusters are read-only).                                  |

Every resource except `misp_sharing_group_member` and `misp_sharing_group_server` also exposes a matching data source (`data.misp_<name>`) for referencing existing records.

## Authentication

Create an API key in MISP: **Administration → List Auth Keys → Add authentication key**. Give it admin scope (required for most resources). Export it as `MISP_API_KEY` or pass it via the `api_key` attribute.

The API key is marked `Sensitive` in provider schema, so Terraform will not print it in plan/apply output.

## Development

```sh
make build      # compile
make install    # go install to $GOBIN (for dev_overrides)
make test       # unit tests — fast, no network
make testacc    # acceptance tests — requires a live MISP instance (set MISP_URL / MISP_API_KEY / TF_ACC=1)
make generate   # regenerate docs/ from schema + examples/
make lint       # golangci-lint
make fmt        # gofmt
```

See [CLAUDE.md](./CLAUDE.md) for architecture notes, MISP API quirks, and conventions for adding a new resource.

## Contributing

Issues and PRs welcome. If you're adding a new resource, please probe the MISP endpoints directly (curl or PyMISP) before coding — the OpenAPI spec is occasionally wrong, and envelope shapes vary between `/add`, `/edit`, and `/view`. CLAUDE.md documents the quirks worth knowing.

Follow the existing conventions: one file per resource, acceptance test covering create → update-in-place → import → destroy, matching data source, and example Terraform under `examples/resources/<name>/`.

## License

[Mozilla Public License 2.0](./LICENSE).
