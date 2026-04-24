# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and the project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.1] - 2026-04-24

### Documentation

- **Stop demoing hardcoded `authkey` values.** The `misp_server` and `misp_sharing_group_server` example `.tf` files showed `authkey = "0123456789abcdef..."` as a literal string. The schema correctly marks `authkey` as `Sensitive` so it never leaks into plan/apply output, but the example was teaching users to hardcode secrets in source control. Switched to the `var.peer_authkey` pattern with `sensitive = true` and an inline comment listing safe sources (env vars, `-var`, `*.tfvars` out of git, secrets backends).
- Added an **Authentication & secret handling** section to the README covering the same guidance at the top level.

### Fixed

- `make generate` was a silent no-op: `cd tools && go generate ./...` matched zero packages because `tools/tools.go` has a `//go:build tools` constraint. Now invokes `tfplugindocs` directly, matching what CI already does.

### Notes

- No schema changes, no behavioural changes, no breaking changes. v0.1.0 remains functionally identical and stays on the Registry; v0.1.1 supersedes it as the recommended version. If you pinned `~> 0.1`, your next `terraform init` will pick up v0.1.1 automatically.

## [0.1.0] - 2026-04-23

Initial release.

### Added

Fourteen resources and thirteen matching data sources covering MISP configuration:

- `misp_organisation` — organisations, the primary unit of access control
- `misp_tag` — tag catalog entries
- `misp_sharing_group` — sharing group definitions
- `misp_sharing_group_member` — org-in-sharing-group junction
- `misp_sharing_group_server` — server-in-sharing-group junction
- `misp_user` — user accounts
- `misp_role` — custom permission roles
- `misp_taxonomy` — enable/disable bundled taxonomies by namespace
- `misp_warninglist` — enable/disable bundled warninglists by name
- `misp_noticelist` — enable/disable bundled noticelists by name
- `misp_feed` — feed subscriptions
- `misp_server` — sync server peers
- `misp_setting` — global instance settings (`MISP.baseurl`, `Security.*`, etc.)
- `misp_galaxy_cluster` — custom galaxy clusters (bundled clusters are read-only)

### Scope

Strictly MISP **configuration** — events, attributes, sightings, analyst notes, and other operational data are out of scope by design. Use PyMISP or the MISP UI for analyst workflows.

### Implementation notes

Built on `terraform-plugin-framework` (protocol v6), Go 1.24+. All resources have acceptance test coverage against MISP 2.5.x. The client package includes `FlexBool`, `FlexString`, and `DomainList` tolerance types that absorb MISP's per-endpoint JSON-type inconsistencies.
