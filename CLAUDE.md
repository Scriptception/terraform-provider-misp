# CLAUDE.md

Guidance for future contributors (human or AI) working on this repo.

## What this is

A Terraform provider for [MISP](https://www.misp-project.org/) — the open-source threat intelligence and sharing platform. Scope is **strictly MISP configuration** — organisations, users, sharing groups, feeds, taxonomies, settings, and similar. Event-level / operational data (events, attributes, sightings, analyst notes, logs, jobs) is out of scope and will never be added. That's a permanent design decision, not a v0.x limitation. Analysts use PyMISP or the MISP UI for day-to-day workflows; this provider is for declaratively bootstrapping and maintaining the MISP instance itself.

## Build & test

Go 1.24+ and Terraform 1.6+. All common tasks go through `make`:

    make build      # compile
    make install    # go install to $GOBIN (required for dev_overrides)
    make test       # unit tests only; fast, no network
    make testacc    # acceptance tests; requires a live MISP instance
    make generate   # regenerate docs/ from schema + examples/
    make lint       # golangci-lint
    make fmt        # gofmt

### Running acceptance tests against a live MISP

    export MISP_URL=https://your-misp.example.com
    export MISP_API_KEY=<your api key>
    export MISP_INSECURE=true   # if self-signed
    export TF_ACC=1
    go test -v -timeout 15m ./internal/provider/

Tests create and destroy their own resources prefixed with `tf-acc-*` / `tf:acc:*` and clean up after themselves. A few tests adopt bundled MISP resources (`admiralty-scale` taxonomy, `"List of known Akamai IP ranges"` warninglist, `gdpr` noticelist) — these are chosen because they ship disabled by default, so the test's destroy step returns the instance to its pre-test state.

## Architecture

    main.go                                — provider entry point
    internal/
      client/                              — thin HTTP wrapper around MISP's REST API
        client.go                          — do(), APIError, IsNotFound
        flexbool.go                        — FlexBool & FlexString tolerance types
        domain_list.go                     — DomainList type for restricted_to_domain
        <resource>.go                      — one file per resource's API surface
      provider/
        provider.go                        — provider definition + resource/data-source registration
        <resource>_resource.go
        <resource>_data_source.go
        <resource>_resource_test.go
        testing.go                         — acctestResourceName() helper
    tools/tools.go                         — anchors tfplugindocs for `go generate`
    docs/                                  — generated; never hand-edit; run `make generate`
    examples/                              — examples consumed by tfplugindocs
    local-test/                            — gitignored playground; never committed

**Convention:** one file per resource in each package. Resource-model types, schema, CRUD, and `fooFromModel` / `fooToModel` helpers all live in `<resource>_resource.go`. Data sources reuse the resource model where possible.

Built on `terraform-plugin-framework` (protocol v6), not the legacy SDKv2. Framework semantics differ from SDKv2 — notably, `ListNestedAttribute` uses attribute syntax (`elements = [ {...}, {...} ]`) in HCL, not block syntax (`elements { ... }`).

## MISP API quirks to know about

MISP's REST API is inconsistent across endpoints. Expect friction.

**Envelope vs flat responses.** Most endpoints wrap the object in a capitalized key (`{"Organisation":{...}}`, `{"Tag":{...}}`), but some don't:
- `GET /tags/view/{id}` returns flat — unlike `/tags/add` and `/tags/edit` which envelope.
- `GET /servers/getSetting/{name}` returns flat (it's an RPC, not a REST object).
- `GET /servers/view/{id}` doesn't exist at all — read via `/servers/index` and filter.

**JSON type instability.** The same field returns different types from different endpoints:
- Booleans may come back as `true/false`, `1/0`, or `"1"/"0"`. `client.FlexBool` absorbs all three.
- Numeric-looking strings (`nids_sid`, `rate_limit_count`) come back as quoted strings or bare numbers. `client.FlexString` absorbs both.
- `organisation.restricted_to_domain` returns four shapes across three endpoints: `null` from `/add`, `[]` from `/view`, `"[]"` from `/edit`, and `"[\"a\",\"b\"]"` (stringified JSON array) for populated edit responses. `client.DomainList` normalises all four.

**500-on-success.** MISP 2.5.x `/servers/add` and `/servers/edit` return HTTP 500 *even when the server record is successfully saved*. `client.CreateServer` and `UpdateServer` tolerate 5xx and fall back to looking the record up by name/id via `/servers/index`. If a future MISP release fixes this, the happy path returns first and the workaround never fires.

**HTTP method inconsistency.** OpenAPI sometimes says one method and the endpoint accepts/requires another:
- `/sharing_groups/delete/{id}` — DELETE per spec; POST also accepted. We use DELETE.
- `/admin/users/edit/{id}` — PUT per spec; POST also works. We use PUT.
- **`/sharing_groups/removeServer/{sgId}/{serverId}`** — OpenAPI says `(sharingGroupServerId, serverId)` but the endpoint actually wants `(sharingGroupId, serverId)`, symmetric with addServer. Verified empirically. The client carries an inline warning.

**Async actions (not modelled).** `POST /galaxy_clusters/publish/{id}` returns `{"message":"Publish job queued. Job ID: ..."}` — it queues a worker, not a synchronous result. Not Terraform-shaped. `misp_galaxy_cluster.published` is Computed-only for this reason.

**Bundled-vs-custom.** Taxonomies, warninglists, noticelists ship with MISP. The corresponding resources use an "adopt-existing" pattern: `Create` looks up by stable identifier (namespace / name) and toggles enabled; `Delete` disables (never removes). `misp_galaxy_cluster` is the exception — it creates new custom clusters, and rejects imports of bundled ones (MISP marks bundled with `default: true`).

**`json:"...,omitempty"` is dangerous on mutable collections.** For fields where "empty" is a meaningful state (e.g. clearing a list), omitempty drops the field from the wire, and MISP treats absent fields as "no change." `organisation.restricted_to_domain` does NOT use omitempty for this reason — users must be able to clear a previously populated list.

## Adding a new resource

Copy an exemplar. Pick the closest fit:
- Standard CRUD → `misp_sharing_group`
- Adopt-existing with enable/disable → `misp_taxonomy` (separate enable/disable endpoints) or `misp_warninglist` (single toggle with body) or `misp_noticelist` (pure toggle, no body — read-first-then-flip)
- Junction (many-to-many) → `misp_sharing_group_member`
- JSON quirks / flex types → `misp_role` (uses FlexBool + FlexString)

Checklist:
- [ ] Client struct + CRUD methods in `internal/client/<name>.go`
- [ ] Resource + schema + CRUD + ImportState in `internal/provider/<name>_resource.go`
- [ ] Data source in `internal/provider/<name>_data_source.go`
- [ ] Acceptance test (create → update → import → destroy) in `internal/provider/<name>_resource_test.go`
- [ ] Examples: `examples/resources/misp_<name>/{resource.tf,import.sh}` and `examples/data-sources/misp_<name>/data-source.tf`
- [ ] Register `NewXResource` / `NewXDataSource` in `internal/provider/provider.go`
- [ ] `make generate` to regenerate docs
- [ ] Run against live MISP: `TF_ACC=1 ... go test ./internal/provider/`

**Probe the API first.** MISP's OpenAPI spec is imperfect and occasionally wrong. Always curl the endpoints yourself before coding — envelope shape, request body format, HTTP method, and field types are all worth confirming. Past bugs have come from trusting the spec over live behaviour.

## Things NOT to add

Operational data is explicitly out of scope:
- `misp_event`, `misp_attribute`, `misp_object`, `misp_object_reference`
- `misp_sighting`, `misp_correlation`, `misp_event_report`
- `misp_analyst_note`, `misp_analyst_opinion`, `misp_analyst_relationship`
- Anything under `/logs`, `/audit_logs`, `/jobs`
- Trigger endpoints (`/servers/pull`, `/feeds/fetchFromFeed`, `/taxonomies/update`, `/galaxy_clusters/publish`) — these are imperative actions, not state.

If a use case comes up that feels operational, push back or direct the user toward PyMISP.

## Testing philosophy

- **Unit tests** for the client package cover JSON parsing quirks (FlexBool, FlexString, DomainList) and HTTP helpers. Fast, no external dependencies.
- **Acceptance tests** drive real Terraform lifecycles against real MISP. Each resource has at minimum: create → update-in-place → import → destroy.
- **No mocks in acceptance tests.** The whole point is to catch MISP API reality that unit tests can't.
- **Non-destructive where possible.** Tests that adopt bundled MISP resources (taxonomies, warninglists, noticelists) pick entries that ship disabled by default, so the test's destroy step returns the instance to its pre-test state. Do NOT test against `tlp` or other canonical taxonomies — you'll silently disable them on real instances.
- **Clean up or die trying.** If a test leaves `tf-acc-*` / `tf:acc:*` resources behind on MISP, that's a test bug — investigate.

## Scope self-check

Every time someone suggests a new resource, ask: is this something an admin configures once (or rarely) and checks into git, or something an analyst creates during day-to-day investigation? Only the former belongs here.
