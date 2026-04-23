# TODO

Followups that aren't blockers. Nothing here prevents the current release from
being used.

## Testing

- **Sweepers.** Register per-resource sweeper functions with
  `resource.AddTestSweepers` so orphaned `tf-acc-*` / `tf:acc:*` resources are
  cleaned up between test runs. Currently a panicking or interrupted test
  leaves garbage on the target MISP and subsequent tests fail with name
  collisions. Run with `go test -sweep=all ./internal/provider/`. One sweeper
  per resource, ~20 lines each.

- **Provider-config negative tests.** Assert helpful error messages when
  provider config is wrong (missing url/api_key, malformed url, unreachable
  host, wrong api key). Use `resource.TestCase` with `ExpectError` regex
  matchers. No live MISP required. Roughly 10 tests, ~150 lines total.

- **Unit tests for all client CRUD methods.** Currently only tag / org /
  setting / taxonomy client methods have unit tests (via `httptest.Server`
  stubs); feed / server / role / noticelist / warninglist / galaxy_cluster /
  sharing_group* are at 0% *unit* coverage. They're implicitly covered by
  acceptance tests but live-MISP coverage doesn't count in `go test -cover`.
  Adding per-method stubs would bring client coverage from 27% → ~80% and let
  PR CI catch MISP-shape regressions without needing a live instance.

- **Drift-detection tests.** Delete a resource out-of-band (via `curl`), then
  run `terraform plan` and confirm it wants to recreate. Confirms that `Read`
  correctly 404-handles every resource, not just the ones we happened to test
  manually.

## Deferred resources

- **`misp_auth_key`**. Config-shaped (admin provisions bot-account API keys
  declaratively) but carries secret material, so needs careful design for
  write-only argument handling, interaction with `Security.advanced_authkeys`,
  and rotation semantics. High value only at scale (dozens of bot accounts);
  for small deployments the MISP UI is fine. Deliberately held back from
  v0.1.

## Deferred attributes on existing resources

- **`misp_sharing_group_member.extend`.** MISP's `/sharing_groups/addOrg/{sg}/{org}`
  endpoint accepts no request body — only URL path params — so there's no way
  to set `extend` at create time via the exposed API. Supporting it would
  require editing the parent sharing group with the full member array.

- **`misp_sharing_group_server.all_orgs`.** Same problem as above.

- **`misp_galaxy_cluster.published` as mutable.** `POST /galaxy_clusters/publish/{id}`
  is asynchronous (queues a worker job), so exposing `published` for write
  would cause plan flap as the async state converges. Currently Computed-only;
  users publish via the MISP UI or PyMISP.

## Quality-of-life

- **Retry with backoff** for transient 5xx in `internal/client/client.go`.
  MISP's server endpoints already use a targeted 500-tolerance workaround;
  a general retry layer could fold other flaky paths into the happy path too.

- **Better error messages.** Some MISP error paths return HTML login pages
  instead of JSON (e.g. when the auth key is wrong and MISP redirects to the
  login screen); the client currently surfaces the raw HTML which is ugly.
  Parse the response, detect redirects, return a "looks like your API key
  is wrong" hint.

- **Bootstrap-module example.** A standalone `examples/modules/bootstrap/`
  that takes a handful of variables (org name, admin email, taxonomies to
  enable) and produces a clean-MISP baseline. Much more useful for onboarding
  than the piecemeal `examples/resources/` content. The `local-test/`
  directory is a close cousin but it's gitignored.

- **Richer `_list` data sources.** `data.misp_organisations` (plural) for
  iteration, `data.misp_feeds`, etc. — lets consumers loop over existing
  resources instead of hardcoding ids.

- **Import-by-name / uuid** where applicable. Currently most resources
  require numeric id for import; `misp_organisation` by name or UUID,
  `misp_tag` by name, `misp_user` by email would be friendlier.

## Test-infra annoyances

- `TestAccSettingResource_basic` can leave `MISP.welcome_text_*` set to
  non-empty values because MISP rejects updates with empty strings (HTTP 403).
  Either pick a different setting to exercise, or make the test tolerate
  whatever default MISP reports.

- Taxonomy / warninglist / noticelist tests rely on specific bundled entries
  (`admiralty-scale`, `"List of known Akamai IP ranges"`, `gdpr`) being
  present and disabled by default. If MISP changes its bundled catalog,
  these tests break silently. Consider dynamically probing for any disabled
  entry and using the first match.
