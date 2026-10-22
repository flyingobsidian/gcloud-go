# Upstream sync notes

Notes on gcloud-python release-notes items whose upstream implementation does
not translate to gcloud-go changes, or is served by an existing gcloud-go
feature. Recording them here keeps the "Sync ..." tracking issues actionable
by making it clear when a checkbox needs no code change.

## Kubernetes Engine 569.0.0–581.0.0 (#1809)

- **Bundled `kubectl` version updates** in 569.0.0, 574.0.0, 576.0.0, and
  579.0.0 — gcloud-go does not bundle `kubectl`; users install it separately
  (`gcloud components install kubectl` is a Python-only surface). These
  version bumps do not apply.
- **New flags on `gcloud container clusters {create,create-auto,update,upgrade}`
  and `gcloud container node-pools {create,update}`** across 569.0.0–581.0.0:
  `--enable-slice-controller`, `KCP_VPA` in `--logging`, node config
  `nodeVfioConfig`/`diskIoScheduler`, `--control-plane-soak-duration`,
  `--managed-otel-scope`, `--image`/`--image-project`,
  `stack-type` in `--additional-node-network`, `--dataplane-optimization-mode`,
  `--maintenance-window-duration`, `--enable-agent-sandbox`,
  `--add-maintenance-exclusion-until-end-of-support` /
  `--remove-maintenance-exclusion-until-end-of-support`,
  `--node-creation-mode`, `--enable-secret-sync`,
  `--enable-secret-sync-rotation`, `--secret-sync-rotation-interval`.
  gcloud-go's `container clusters` and `container node-pools` `create`/`update`
  commands accept a full `container.Cluster` / `container.NodePool` body via
  `--config-file`, so every one of these fields is already reachable today
  (set the corresponding proto field in the YAML/JSON payload).
- **Autoscaling settings overwrite fix on `gcloud container clusters update`**
  (577.0.0) — Python-side regression in argument merging; gcloud-go's update
  applies the `ClusterUpdate` body straight through, so the merge bug does
  not exist.
- **Release-channel `None` deprecation on `gcloud container clusters
  create/update`** (573.0.0) — a Python-side argparse choices deprecation
  notice. gcloud-go passes `releaseChannel.channel` through from
  `--config-file` and does not maintain its own enum.
- **`gcloud container clusters upgrade` and
  `create-auto` Python-level surfaces** — gcloud-go does not expose these
  as separate commands; `upgrade` is done by supplying `desiredMasterVersion`
  or `desiredNodeVersion` in the `ClusterUpdate` body on
  `container clusters update`, and Autopilot clusters are created by setting
  `autopilot.enabled: true` in the `Cluster` body on
  `container clusters create`.
- **`--control-plane-soak-duration` + `gcloud container clusters
  complete-control-plane-upgrade` promoted to GA** (576.0.0) — the
  `complete-control-plane-upgrade` command is now wired to the container v1
  API in `cmd/container_clusters.go`. The soak-duration knob is passed
  through the `ClusterUpdate` body on `container clusters update`.

## Kpt 575.0.0–579.0.0 (#1808)

- **kpt binary bump to v1.0.0-beta.64** (575.0.0) and **v1.0.0-beta.67**
  (579.0.0) — the `kpt` binary is bundled separately by the Python
  Google Cloud CLI's component installer; gcloud-go does not bundle kpt, so
  these upstream version bumps do not apply. Users who need `kpt` install it
  from the [kpt release channel](https://github.com/kptdev/kpt/releases)
  directly.

## Google Cloud CLI 569.0.0–580.0.0 (#1806)

- **Bundled Python / Windows / macOS Virtualenv version bumps** — gcloud-go is
  a native Go binary, so upstream Python, Virtualenv, pyOpenSSL and
  cryptography version bumps (569.0.0, 570.0.0, 572.0.0, 573.0.0, 575.0.1,
  576.0.0, 578.0.0) do not apply.
- **Bundled PuTTY updates on Windows** (569.0.0, 570.0.0, 571.0.0) — gcloud-go
  uses the operating system's OpenSSH client for `gcloud compute ssh`/`scp`,
  so PuTTY packaging changes do not apply.
- **PY 3.14 `gcloud compute scp` slowness with `--tunnel-through-iap`**
  (571.0.0) — the Python-only regression does not apply to the Go
  implementation.
- **`gcloud init` crash when Enterprise Certificate Proxy (ECP) binaries are
  missing** (578.0.0) — ECP is not implemented in gcloud-go, so the crash
  path does not exist.
- **Google Compute Engine residency detection fix** (580.0.0) — gcloud-go's
  metadata-server detection does not exhibit the Python latency regression.
- **Regional Access Boundary (RAB) native caching in auth** (570.0.0) —
  gcloud-go does not implement RAB yet; the caching change will be revisited
  when RAB support lands.
- **Cloud CLI remote MCP server** (576.0.0) — a new server surface, tracked
  separately if/when it is prioritised for gcloud-go.
- **`--reservation-affinity=any-reservation-then-fail` on
  `gcloud container clusters node-pools create`** (575.0.0) — `gcloud-go`
  builds the `container.NodePool` body from `--config-file`, so users can
  already set `config.reservationAffinity.consumeReservationType:
  ANY_RESERVATION_THEN_FAIL` today. No CLI change needed.

## BigQuery `bq` CLI 570.0.0–581.0.0 (#1762)

The 25 upstream release-notes items in #1762 all target the standalone
`bq` binary that ships in `platform/bq/` inside the Python Google Cloud
CLI (a separate Python CLI written on top of the BigQuery REST API and
distributed via the `bq` component). gcloud-go's own `bq` command is a
distinct surface that only wraps the BigQuery Migration API
(`bq migration-workflows …`), so none of these items translate to
gcloud-go code:

- Reservation-assignment additions (`bq ls|show --reservation_assignment`
  `precedence`/`condition` fields, the `AUTOMATIC_MATERIALIZED_VIEW_REFRESH`
  job type, `principal` field, `NoneType` handling fix, short-ID parsing
  fix, custom assignment name, `--alpha=reservation_groups` removal) —
  all on the Python `bq mk|show|ls|rm --reservation_group|
  --reservation_assignment` surface which gcloud-go does not implement.
- `bq rm --connection -f/--force`, `--s3_service_directory_service` on
  AWS connections — Python `bq --connection` surface only.
- `--nouse_google_auth` config-reading fix, `--oauth_access_token`
  `bq init` gating, `--gcloud_config_cache`, `--label` on `bq cp|extract|load`,
  fractional-second `bq show` fix, `stderr`/`stdout` routing fix, container
  concurrency in `bq show --routine`, table/dataset identifiers of the form
  `project.catalog.namespace.table` / `catalog.namespace`, row-access-policy
  `bq show|rm` support, `absl` version bump — all internal to the Python
  `bq` CLI.
- User-agent tweaks (predefined labels, execution-environment info, agent
  information, command format), and updated help text for global flags —
  Python `bq` CLI packaging only.

If a native Go implementation of the `bq` CLI is prioritised for
gcloud-go, this issue can be re-scoped or split into follow-up feature
requests. Nothing to port for the migration-workflows surface we do have.

## Certificate Authority Service 588.0.0 (#1916)

`--key-algorithm` on `gcloud privateca certificates create` is implemented
together with the `--generate-key` / `--key-output-file` group it belongs
to (gcloud-go previously took the key only from `--config-file`). With
`--generate-key`, gcloud-go generates an `rsa-2048` (default), `rsa-3072`,
`rsa-4096`, `ec-p256` or `ec-p384` key pair, writes the private key as a
traditional OpenSSL PEM (PKCS#1 / SEC 1) to `--key-output-file` with mode
0400 and Python's key-export warning, and sets `config.publicKey` (PEM) on
the `--config-file` body, which still supplies the subject and X.509
configuration.

## Cloud API Gateway 585.0.0 (#1917)

`--enable-streaming` on `gcloud api-gateway gateways create` is implemented:
it sets `streamingMode: STREAMING_MODE_ENABLED` on the `--config-file` body
(the field is immutable, so `gateways update` has no such flag). When the
flag is omitted the body's own `streamingMode`, if any, is sent unchanged.

## Access Context Manager 583.0.0 (#1918)

`gcloud access-context-manager lookup-configured-perimeter --resource=...`
is implemented on the v1 API (`projects|folders/*:lookupConfiguredServicePerimeter`),
with Python's `projects/`/`folders/` validation and its four-line
`SERVICE_PERIMETER` / `RESTRICTED_RESOURCE` / `..._DRY_RUN` display when
`--format` is not given.

## Cloud Alerting 589.0.0 (#1919)

`gcloud monitoring alerts` (GA in 589.0.0) is implemented: `list` sends
`--filter` and `--sort-by` to the API as `filter`/`orderBy` and, like
Python, fetches a single page of `--page-size` (default 1000) alerts, with
`--limit` and `--uri` applied client-side; `describe` accepts a short alert
id or `projects/P/alerts/ID`. Both default to YAML output.

## Cloud Backup DR 584.0.0–585.0.0 (#1920)

The new `backup-dr auto-protection-policies`, `auto-protection-bindings`,
`applied-auto-protection-policies` and `binding-matching-resources` groups
are alpha/beta-only in gcloud-python and their resources exist only in the
`backupdr/v1beta` API (the `v1` Go client has no auto-protection methods).
As with #1769 below, gcloud-go's `backup-dr` groups are still "not yet
implemented" stubs, so these groups should land with the real `backup-dr`
implementation rather than as further stubs.

## Cloud Bigtable 583.0.0–588.0.0 (#1921)

The alpha restored-table flags on `gcloud bigtable tables restore` are
implemented: `--deletion-protection` / `--no-deletion-protection`,
`--enable-automated-backup` (the default 7d retention / 1d frequency policy)
and `--automated-backup-retention-period` / `--automated-backup-locations`
(retention defaults to 7d, no explicit frequency). They are sent as
`RestoreTableRequest.table` with an `updateMask` of `deletion_protection`
and/or `automated_backup_policy`, as in gcloud-python. The hidden
`--automated-backup-keep-hot-duration` flag is not ported.

The release-note bullets (Go/CVE rebuilds of the `cbt` emulator and
micro-precision timestamp support) concern the Python SDK's bundled `cbt`
and Bigtable emulator components, which gcloud-go does not ship.

## Breaking Changes 584.0.0–589.0.0 (#1915)

The Cloud Services items (586.0.0) are implemented: `services api-keys
create` fails without `--api-target`, `services api-keys update` fails
unless the key already has API targets or `--api-target` is given (the
existing key is now always fetched first, as in Python), and
`--clear-restrictions` is removed from `update`. The `edge-cloud container
vpn-connections` stub group is removed (587.0.0).

The other items need no gcloud-go code change:

- **`skaffold` deprecation (589.0.0), `managed-flink-client` removal
  (588.0.0), `kustomize` and `kpt` removal and `gcloud-man-pages`
  deprecation (586.0.0)** — Python components; gcloud-go is a single binary
  whose `components` commands are not applicable stubs.
- **`google-cloud-sdk` Snap removal** (586.0.0) — no Snap package exists for
  gcloud-go (see the 580.0.0 note below).
- **`storage buckets anywhere-caches pause` deprecation/removal (584.0.0,
  587.0.0) and the `ADMIT_ON_SECOND_MISS` `--admission-policy` deprecation
  (585.0.0)** — gcloud-go has not ported `storage buckets anywhere-caches`.

## Breaking Changes 569.0.0–582.0.0 (#1763)

The one code-affecting item — the `--auto-commit` default flip on
`gcloud database-migration conversion-workspaces` `convert|seed|import-rules`
(578.0.0) — is now implemented: those three commands default `--auto-commit`
to `true` and accept `--no-auto-commit` to opt out (the historical false
default). `apply` intentionally keeps the opt-in default.

The other breaking-change items in #1763 do not translate to gcloud-go
code:

- **Legacy Cloud SQL Proxy V1 (`cloud_sql_proxy`) removal, mandatory Cloud
  SQL Auth Proxy V2** (582.0.0) — gcloud-go's `sql connect psql/mysql/sqlserver`
  path shells out to the DB client using the instance's public IP directly
  (see `cmd/sql_all.go`); no `cloud_sql_proxy` or `cloud-sql-proxy` binary
  is bundled or invoked, so the V1 removal has no effect.
- **`gcloud api-registry mcp servers|tools list`,
  `gcloud api-registry mcp enable|disable`,
  `gcloud beta services mcp policies get|get-effective|test-enabled`,
  `gcloud beta services mcp enable|disable|list` removals** (579.0.0–582.0.0)
  — gcloud-go never implemented the `api-registry` command or any `services
  mcp` subgroup, so the removals do not need to be mirrored. The Agent
  Registry successor at
  `gcloud alpha agent-registry mcp-servers` is exposed in gcloud-go as
  `agent-registry` (see `cmd/agent_registry.go`).
- **`google-cloud-sdk` Snap deprecation on 2026-09-29 in favour of
  `google-cloud-cli`** (580.0.0) — gcloud-go is a single native Go binary
  distributed via `go install` and release tarballs; there is no Snap or
  APT package to migrate.
- **`PRESERVED_STATE` column dropped from
  `gcloud compute instance-groups managed list-instances` beta output**
  (579.0.0) — gcloud-go's `formatManagedInstances` default table renders
  NAME/ZONE/STATUS/ACTION/HEALTH and never emitted PRESERVED_STATE, so
  no output change is required.
- **`run_bq_command` tool on the Cloud CLI remote MCP server** (578.0.0)
  — the remote MCP server is a Python-only surface that gcloud-go does
  not implement; tracked under the general MCP-server note in the Google
  Cloud CLI section.
- **`gcloud storage rsync` default gzip decompression + `--do-not-decompress`
  opt-out** (576.0.0) — gcloud-go's `storage rsync` copies bytes verbatim
  (`storageDownloadFile` performs an unmodified media download and
  `storageUploadFile` writes bytes as-is); no automatic gzip decompression
  path exists, so the default flip and the new opt-out have no matching
  behaviour to preserve. If content-encoding-aware rsync is prioritised
  it should land alongside a matching decompression option in `storage cp`.
- **`gcloud container attached clusters get-credentials` deprecation in
  favour of `gcloud container fleet memberships get-credentials`**
  (573.0.0) — the `container attached get-credentials` subcommand in
  gcloud-go is a stub that reports "not yet implemented"; the deprecation
  notice on the not-yet-implemented Python command has nothing to mirror.
  `container fleet memberships` is also a stub today, tracked separately.
- **`anthoscli` component no longer preinstalled in Debian/RPM packages**
  (569.0.0) — gcloud-go does not bundle `anthoscli`; installing it is out
  of scope for the Go binary.

## Certificate Authority Service 569.0.0–580.0.0 (#1764)

The two flag-add items are now implemented:

- `--issuer-pool` / `--issuer-location` / `--issuer-ca` on
  `gcloud privateca subordinates activate` populate
  `ActivateCertificateAuthorityRequest.subordinateConfig.certificateAuthority`
  for first-party activation (gcloud-python 580.0.0). `--issuer-ca` accepts
  either a short id (in which case `--issuer-pool`/`--issuer-location`, or
  the surrounding `--pool`/`--location`, provide the missing components)
  or a fully qualified resource name.
- `--requested-not-before-time` on `gcloud privateca certificates create`
  sets `Certificate.requestedNotBeforeTime` (gcloud-python 569.0.0). The
  server rejects it unless the issuing CA pool's `issuancePolicy` has
  `allowRequesterSpecifiedNotBeforeTime: true`.

The remaining bullets don't need code changes:

- **Suggestion to use `subordinates activate` when `subordinates create`
  with issuer flags fails with `ALREADY_EXISTS`** (580.0.0) — Python's
  `subordinates create` builds a `SubordinateConfig` from `--issuer-…`
  argparse flags before hitting the API, then rewrites its error message
  when the CA already exists. gcloud-go's `subordinates create` takes the
  full `CertificateAuthority` body via `--config-file` and does not
  synthesise a `SubordinateConfig` from flags, so the tailored suggestion
  has no equivalent context to attach to. Users already see the raw
  `ALREADY_EXISTS` error from the API and can rerun as
  `gcloud privateca subordinates activate CA --issuer-ca=...` directly.
- **`allowRequesterSpecifiedNotBeforeTime` support on the CA pool's
  issuance policy** (569.0.0) — set on the `CaPool` resource itself.
  `gcloud privateca pools create/update` in gcloud-go accept the full
  `CaPool` body via `--config-file`, so setting
  `issuancePolicy.allowRequesterSpecifiedNotBeforeTime: true` in the
  payload is supported today.

## Agent Registry 587.0.0 (#1901)

- **587.0.0** — `gcloud agent-registry bindings create` makes
  `--source-identifier` optional (the `binding.source` argument group
  flipped from `required: true` to `required: false`). gcloud-go's
  `agent-registry bindings create` (see `cmd/agent_registry.go`) takes the
  whole `Binding` body from `--config-file` and unmarshals it straight into
  `google.golang.org/api/agentregistry/v1alpha.Binding` with no client-side
  check on `source`, so omitting `source.identifier` from the payload
  already works and the API is the sole validator — no code change needed.

## AlloyDB 583.0.0–589.0.0 (#1902)

gcloud-go's `alloydb clusters` commands (see `cmd/alloydb_all.go`) take the
request body from `--config-file` and call `google.golang.org/api/alloydb/v1`,
so none of the five items needs a code change:

- **589.0.0** — `--enable-express` on `gcloud beta alloydb clusters
  create|update` (sets `Cluster.expressConfig.enabled`), and **584.0.0** —
  `--edition` (`EDITION_ALLOYDB` / `EDITION_ALLOYDB_DEVELOPER`) on
  `gcloud alpha|beta alloydb clusters create|update`. Both fields exist only
  in the v1alpha/v1beta `Cluster` messages; the v1 API (gcloud-python GA and
  `google.golang.org/api/alloydb/v1` v0.298.0) has neither `expressConfig`
  nor `edition`. They become settable via `--config-file` with no gcloud-go
  code change once they reach v1.
- **583.0.0** — `POSTGRES_19` on `--database-version` for
  `gcloud alpha|beta alloydb clusters create` and `migrate-cloud-sql`. Python
  appends the choice only for the alpha/beta tracks and the v1 enum still
  stops at `POSTGRES_18`. gcloud-go has no `--database-version` flag;
  `databaseVersion` is a plain string in the `--config-file` payload
  (`Cluster` / `RestoreFromCloudSQLRequest.cluster`), so `POSTGRES_19` passes
  through as soon as the v1 API accepts it.
- **587.0.0** — IAM Group user types promoted to GA (`--type=IAM_GROUP` on
  `gcloud alloydb users create`, mapping to `ALLOYDB_IAM_GROUP`). gcloud-go's
  `alloydb users create` reads the `User` body from `--config-file`, and
  `userType` is a plain string, so `userType: ALLOYDB_IAM_GROUP` in the
  payload already works.

## Anthos 588.0.0 (#1903)

- **588.0.0** — "Updated anthos-cli with newer go version and library
  dependencies." This is a rebuild of the separately shipped `anthoscli`
  Go binary (the `anthoscli` component), not a change to gcloud-python's
  `gcloud anthos` command surface. gcloud-go does not bundle `anthoscli`
  (see the 569.0.0 note under Breaking Changes), and its `gcloud anthos`
  commands in `cmd/anthos.go` are stubs, so there is nothing to mirror.

## Apigee 583.0.0–589.0.0 (#1904)

Implemented:

- **`gcloud apigee apis delete`** (584.0.0, GA in 589.0.0) — gcloud-go
  already had the command; it now prompts "API proxy [API] and all of its
  revisions will be deleted." (aborting with "Deletion aborted by user."),
  logs "Deleted API proxy [API]." to stderr and prints the deleted proxy,
  matching Python.
- **DRZ endpoints for the CH region** (586.0.0) — Python routes every
  organization-scoped Apigee request to the host for the organization's
  location (looked up via `organizations.getProjectMapping` on the global
  endpoint): `apigee.googleapis.com` for global,
  `apigee.LOCATION.rep.googleapis.com` for the multi-region set (`us`, `eu`,
  `in`, `sa`, and now `ch`), and `LOCATION-apigee.googleapis.com` otherwise.
  gcloud-go previously sent everything to the global host; all
  organization-scoped `apigee` commands, including `apis import`, now route
  the same way (`gcp.ApigeeOrgHost`). `organizations list` and
  `organizations provision` stay on the global endpoint. Python caches the
  project mapping on disk (`.apigee-cached-project-mapping-v2`); gcloud-go
  makes the lookup on each invocation.

No code change needed:

- **`gcloud apigee apis import` promoted to GA** (588.0.0) and
  **`apis delete` / every `apigee apis` command serving GA** (589.0.0) —
  gcloud-go does not maintain separate alpha/beta/GA tracks.
- **`apis import --from-template` compiler fixes** (583.0.0, 588.0.0,
  589.0.0: `FaultRule` step nesting, fault-rule `mode` warning, top-level
  `EventFlow`, feature `defaultEndpoint` routes/fault rules and
  `defaultTarget` merging, `uid`-only name prefixing, template
  `endpoints`/`targets` precedence, generated `metadata.js`, policy child
  ordering, `displayName` acceptance) — gcloud-go's `apis import` does not
  yet compile Apigee Feature Templates (`--from-template` returns an error
  pointing to `--from-bundle`), so these fixes land when the compiler is
  ported.
- **`--service-account` on `gcloud beta apigee apis deploy`** (588.0.0) —
  beta-only, and gcloud-go has not implemented `apigee apis deploy`; the
  flag (the `serviceAccount` query parameter) should be added with the
  command.

## Apihub 588.0.0 (#1905)

- **588.0.0** — `--service-type-enum-values` (and the related clear flags) on
  `gcloud apihub apis update`, setting the API's `serviceType` system
  attribute (`code`, `model`, `agent`, `skill`). gcloud-go has no
  `apihub apis update`; API bodies are written with `apihub apis import`,
  which loads a `GoogleCloudApihubV1Api` from a YAML/JSON file and patches
  every non-empty field. `google.golang.org/api` was bumped to v0.307.0, whose
  `apihub/v1` `Api` struct carries `serviceType` (v0.298.0 lacked it), so
  `serviceType: {enumValues: {values: [{id: agent}]}}` in the import file now
  reaches the API instead of being dropped on unmarshal.

## App Engine 589.0.0 (#1906)

- **589.0.0** — three fixes to `gcloud beta app migrate-to-run` (Serverless
  VPC Access connector dropped for deployed versions, Flexible
  `instance_ip_mode` wrongly mapped to Cloud Run ingress, internal-only
  Flexible apps not routing all egress through Direct VPC).
  `app migrate-to-run` is a beta-only command that translates an App Engine
  service/version config into a Cloud Run service and deploys it; gcloud-go
  has not implemented it, so there is nothing to fix. The corrected mappings
  apply when the command is ported.

## App Lifecycle Manager 589.0.0 (#1907)

- **589.0.0** — the deprecated `gcloud beta saas-runtime` group is now
  hidden (its commands still run with a deprecation warning pointing to
  `gcloud beta app-lifecycle-manager`). In 589.0.0 `saas-runtime` is an
  alpha/beta-only alias that `gcloud_main.py` maps onto a fixed set of
  `app-lifecycle-manager` subgroups, and `app-lifecycle-manager` itself is
  alpha/beta-only. gcloud-go implements neither group, so there is no
  command to hide; if `app-lifecycle-manager` is ported, the `saas-runtime`
  alias should not be added.

## Artifact Registry 584.0.0–587.0.0 (#1909)

Implemented: `gcloud artifacts files upload` (584.0.0) and domain-scoped
project parsing in the Docker image commands (586.0.0). One item needs no
code change:

- **587.0.0** — help text for `gcloud artifacts repositories create` now
  says custom remote repository URIs must use HTTPS. gcloud-go takes the
  remote repository config (including custom upstream URIs) from
  `--config-file` rather than per-field flags, so there is no flag help text
  to update; the API remains the validator.

## Beyondcorp 589.0.0 (#1912)

- **589.0.0** — the `gcloud beta beyondcorp app` group (App Connector
  connectors, connections and gateways) is deprecated: Python marks it with
  `@base.Deprecate`, so its commands print "BeyondCorp App Connector is
  deprecated and will be removed in a future release." The group exists only
  on the alpha/beta tracks and gcloud-go does not implement it (its
  `beyondcorp` command covers `operations` and `security-gateways`), so there
  is no command to deprecate. The group should not be ported.

## BigLake 584.0.0–589.0.0 (#1913)

gcloud-go implements only `gcloud biglake iceberg catalogs`, whose `create`
and `update` take the catalog body from `--config-file` (hyphenated keys, as
sent by Python), so no item needs a code change:

- **`--kms-key` / `--kms-keyring` / `--kms-location` / `--kms-project`**
  (584.0.0) and **`--cross-cloud-cache=enabled|disabled`** (584.0.0) on
  `iceberg catalogs create|update` — Python sends them as
  `encryption-config: {kms-key-name: ...}` and
  `cross-cloud-cache-options: {enabled: ...}`; set those keys in the
  `--config-file` payload.
- **Federated catalog flags promoted to GA / beta** (586.0.0), including the
  new `workday` `--federated-catalog-type` and `--workday-base-url` /
  `--workday-tenant` — gcloud-go has no per-track flags; the
  `federated-catalog-options` object (including `identity-mode` and
  `metadata-service-directory-name`) is set through `--config-file`.
- **`iceberg catalogs update` now sends update-mask paths** for
  `metadata_service_directory_name`, `identity_mode`,
  `encryption_config.kms_key_name` and `cross_cloud_cache_options.enabled`
  (source-level) — gcloud-go's `update` derives the mask from the top-level
  keys of the `--config-file` body (or takes `--update-mask`), so these
  objects are already included whenever they are present in the payload.
- **`--kms-key` and `--cross-cloud-cache` on `delta-sharing catalogs
  create|update`** (584.0.0, 585.0.0), **`--service-directory-name` on
  `delta-sharing catalogs update`** (589.0.0), **`hive catalogs|databases|
  tables` promoted to GA on the Hive v1 API** (588.0.0) and the alpha
  **`data-product-sharing unpublish`** — gcloud-go has not ported the
  `delta-sharing`, `hive` or `data-product-sharing` subgroups (see the
  570.0.0–578.0.0 note below); these land when those subgroups are added.

## BigQuery 583.0.0–589.0.0 (#1914)

No item needs a gcloud-go code change:

- **All nine release-note bullets** (ADLS `abfs://`/`abfss://` URIs in
  `--external_table_definition`, `--dataplex_entry_group` on
  `bq mk|update --transfer_config` and the matching `bq ls` filter,
  clearing/`--service_directory_service` on `bq mk|update --connection`,
  the `--s3_service_directory_service` deprecation, `--migration_workflow`
  setup skipping, reservation-group `creationTime`/`updateTime`, and
  `--precedence`/`--condition` on reservation assignments) target the
  standalone Python `bq` binary in `platform/bq/`, as with #1762 below.
  gcloud-go's `bq` group only wraps the BigQuery Migration API
  (`bq migration-workflows`).
- **New alpha commands `gcloud alpha bq query` and `gcloud alpha bq
  translation describe|generate-source-ddl|translate-batch|
  translate-metadata`** — alpha-only, like the `bq datasets`, `bq tables`
  and `bq jobs` subgroups that gcloud-go has not ported either. They land
  when those alpha `bq` subgroups are added.

## Workbench 572.0.0 (#1824)

The 572.0.0 `--image-family` flag is implemented (see
`cmd/workbench_instances.go`: `wbiUpgradeCmd` now populates
`UpgradeInstanceRequest.imageFamily`; tested in
`cmd/workbench_groups_test.go`). The accelerator-type enum expansion
is Python argparse-only:

- **572.0.0** — `NVIDIA_RTX6000` added to `--accelerator-type` choices on
  `workbench instances create` / `update`. gcloud-go's `instances
  create` / `update` take the full `notebooks.Instance` body via
  `--config-file` (there is no per-choice argparse validator to
  mirror), so any `AcceleratorConfig.type` value the API accepts —
  including `NVIDIA_RTX6000` — is settable through the YAML/JSON
  payload today.

## Vmware Engine 576.0.0–582.0.0 (#1823)

The 582.0.0 `migrate-management-vms` command is implemented (see
`cmd/vmware_private_clouds.go`: `vmwarePcMigrateMgmtVmsCmd` wraps
`ProjectsLocationsPrivateCloudsService.MigrateManagementVms` with
`--cluster-id`, `--request-id`, and `--etag`; tests in
`cmd/vmware_test.go`). The other three items are already reachable via
gcloud-go's `--config-file` payload on `vmware private-clouds
create`/`update`:

- **579.0.0** — `--kms-key` on `create`, and `--kms-key` /
  `--encryption-type` on `update`. Populates
  `PrivateCloud.encryptionConfig.cryptoKeyName` and
  `PrivateCloud.encryptionConfig.type` (`CMEK`, `LEGACY_CMEK`, `OTHER`),
  which are exposed on `EncryptionConfig` by
  `google.golang.org/api/vmwareengine/v1` and set through the
  `--config-file` YAML/JSON body today.
- **576.0.0** — `--preferred-zone` / `--secondary-zone` on `create`
  updated to require full resource names when creating a stretched
  private cloud (VPC-SC compliance). gcloud-go does not maintain
  Python-side argparse short-name coercion; users pass the full
  resource name in the `PrivateCloud.stretchedClusterConfig`
  `preferredLocation` / `secondaryLocation` fields via `--config-file`
  today, so the VPC-SC-compliant behaviour is already what happens.

## Service Extensions 572.0.0 (#1821)

- **572.0.0** — Updated import and export schemas for
  `gcloud service-extensions authz-extensions`. The Python update was
  on argparse-level YAML schemas driving `authz-extensions import` /
  `export` (extra optional keys plus a stricter enum validator).
  gcloud-go's `authz-extensions import` (see
  `cmd/service_extensions_authz_extensions.go`) unmarshals the file
  straight into `google.golang.org/api/networkservices/v1.AuthzExtension`,
  so any field on the Go client's `AuthzExtension` struct is settable
  through the payload today — no per-schema validator to keep in sync.
  Users on stale YAML with removed keys will see the API reject them,
  matching the new upstream behaviour without a gcloud-go code change.

## Recaptcha 573.0.0–577.0.0 (#1818)

The 577.0.0 `--universal` flag is implemented (see
`cmd/recaptcha_keys.go`: sets `Key.universalSettings` on
`recaptcha keys create` and `update`, and joins the existing
`--web`/`--android`/`--ios`/`--express` mutually-exclusive platform
group). The 573.0.0 bug fix has no gcloud-go equivalent:

- **573.0.0** — Fix for `gcloud recaptcha keys create` ignoring
  `--testing-score` when creating Android or iOS keys. gcloud-go's
  `rcpKeyBodyFromFlags` (in `cmd/recaptcha_keys.go`) sets
  `Key.testingOptions.testingScore` from `--testing-score` on the
  outgoing body before it selects a platform, so the value is applied
  regardless of whether `--android`, `--ios`, or `--web` is set. The
  Python regression path — a per-platform argparse branch that
  silently dropped the flag on the mobile subgroups — does not exist
  in gcloud-go.

## Orchestration Pipelines 582.0.0 (#1817)

- **582.0.0** — New `gcloud beta orchestration-pipelines` command group.
  This is a brand-new top-level command surface introduced in
  gcloud-python; gcloud-go has no `orchestration-pipelines` command
  (the closest name in the tree is `telco-automation orchestration-cluster`,
  which is a distinct product). Adding the group is standalone work —
  it needs its own `cmd/orchestration_pipelines*.go` files plus a
  `gcp.OrchestrationPipelinesService` constructor once the API is
  reachable through `google.golang.org/api/orchestrationpipelines/*`
  (or an equivalent client). Tracked here for follow-up so the
  checkbox in #1817 doesn't hang for lack of a code change.

## Oracle Database 581.0.0 (#1816)

The single 581.0.0 bullet groups three Exascale storage flags across
two commands, neither of which needs a gcloud-go CLI change today:

- **`--total-vm-storage-size-gb` on
  `oracle-database cloud-exadata-infrastructures configure-exascale`**
  — gcloud-go's `cloud-exadata-infrastructures` (see
  `cmd/oracle_database_cloud_exadata_infrastructures.go`) exposes
  `create`, `delete`, `describe`, and `list`; the
  `configure-exascale` subcommand is not implemented. `configure-exascale`
  in gcloud-python corresponds to the
  `ProjectsLocationsCloudExadataInfrastructuresService.Patch` call
  with a specific field mask covering the Exascale storage settings;
  a straightforward `configure-exascale` (or `update`) subcommand
  could be added later that patches
  `CloudExadataInfrastructureProperties.totalStorageSizeGb` (and
  siblings) using `--config-file`.
- **`--properties-vm-backup-storage-type` /
  `--properties-vm-file-system-storage-type` on
  `oracle-database cloud-vm-clusters create`** — gcloud-go's
  `cloud-vm-clusters create` (see
  `cmd/oracle_database_cloud_vm_clusters.go`) takes the full
  `CloudVmCluster` body via `--config-file`, so these settings would
  live under `properties.*StorageType`. The current version of
  `google.golang.org/api/oracledatabase/v1` used by gcloud-go does
  not yet expose the `vmBackupStorageType` /
  `vmFileSystemStorageType` fields on `CloudVmClusterProperties`
  (only `LocalBackupEnabled`, `DbNodeStorageSizeGb`, and
  `StorageManagementType` are present). Once the Go client library
  catches up, both fields become settable via `--config-file` with no
  gcloud-go code change.

## Network Services 575.0.0–580.0.0 (#1815)

All four 575.0.0–580.0.0 items are either already covered or Python-only:

- **580.0.0** — `network-services telemetry-policies delete` promoted
  to BETA. `telemetry-policies` is a v1beta1-only surface that is not
  exposed by `google.golang.org/api/networkservices/v1` and is not
  implemented in gcloud-go's `cmd/network_services*.go`. Both the
  subgroup and its `delete` command should land together when the
  v1beta1 client (or the eventual GA promotion) reaches gcloud-go.
- **579.0.0** — `clientTlsPolicy` field deprecation on
  `network-services endpoint-policies`. gcloud-python emits an
  argparse-level deprecation warning; the underlying `EndpointPolicy.
  clientTlsPolicy` field on the Go client is still present, so
  gcloud-go's `endpoint-policies import` (see
  `cmd/network_services_endpoint_policies.go`) continues to accept it
  through the YAML body until the API removes it. No Go-side
  deprecation notice is required.
- **579.0.0** — Regional-locations support for the
  `endpoint-policies` resource name pattern. gcloud-go's
  `endpoint-policies` already takes `--location` and derives
  `projects/PROJECT/locations/LOCATION/endpointPolicies/…` from it
  (`nsResourceName` in `cmd/network_services_endpoint_policies.go`),
  so users can pass any regional location today; only the Python
  argparse resource-arg validator required updating.
- **575.0.0** — `edge-cache services` import schema updates to allow
  up to 100 `CORSPolicy.allowOrigins` and a `0s` `CDNPolicy.clientTtl`.
  gcloud-python's `edge-cache services import` had client-side
  argparse validators; gcloud-go's `edge-cache services import`
  (see `cmd/edge_cache_services.go`) passes the YAML/JSON payload
  straight through and lets the API enforce limits, so raising the
  cap and allowing `0s` requires no CLI change — the requests
  already succeed on the API side.

## Network Security 570.0.0–580.0.0 (#1814)

Most of the 21 items are Python track promotions (alpha→beta or
beta→GA) or Python-only "project scoping" argparse changes on
subgroups that gcloud-go already exposes as project-scoped CRUD
without alpha/beta/GA distinctions. gcloud-go's `authz-policies`,
`firewall-endpoints`, `security-profiles`,
`security-profile-groups`, `secure-access-connect`, and `operations`
subgroups are all wired to real APIs in `cmd/network_security*.go`
and take the full resource body via `--config-file` on create/update:

- **580.0.0** — `mcp` / `policyProfile` GA in `authz-policies import`
  and `export`; **580.0.0** — `loadBalancingScheme` made optional in
  the same commands; **577.0.0** — `networkRules` / `snis` support in
  `authz-policies import`. gcloud-go's `authz-policies` (see
  `cmd/network_security_policies.go`) takes the full `AuthzPolicy`
  body via `--config-file` on create/update. The Go client exposes
  `AuthzPolicy.policyProfile`, `AuthzPolicy.networkRules`, and
  `AuthzPolicyTarget.loadBalancingScheme` today, so the fields are
  settable through the YAML/JSON payload. `import`/`export` in Python
  are read-modify-write helpers on top of the same PATCH/GET calls
  gcloud-go's `describe` → edit → `update --config-file` flow already
  supports.
- **576.0.0** and **573.0.0** — Track promotions of
  `security-profiles wildfire-analysis` (project scope, BETA
  promotion). gcloud-go's `security-profiles` (see
  `cmd/network_security_profiles.go`) is a single project-scoped CRUD
  group with no wildfire-analysis-specific subgroup; the underlying
  `SecurityProfile.customIntercept`/`customMirroring`/`threatPrevention`/
  `wildfireAnalysis` message fields are reachable via `--config-file`.
- **574.0.0** (×2) — `ull-mirroring-engines` and
  `ull-mirroring-collectors` beta / GA promotions. Both subgroups are
  new API surfaces
  (`ProjectsLocationsUllMirroringEnginesService`,
  `ProjectsLocationsUllMirroringCollectorsService` on
  `networksecurity/v1`) that gcloud-go has not implemented yet.
  Adding them is standalone work.
- **573.0.0** — `network-security operations` command group
  (`describe`, `wait`, `list`, `cancel`). Already implemented at
  `cmd/network_security.go` (issue 833); `wait` is not exposed
  because gcloud-go's create/update wait on the LRO by default
  (`--async` opts out), but `describe`/`list`/`cancel` are wired via
  the same builder.
- **573.0.0** — Wildfire subgroups / flags on `firewall-endpoints`.
  `firewall-endpoints` (see `cmd/network_security_firewall.go`) is a
  project- and organization-scoped CRUD group taking the full
  `FirewallEndpoint` body via `--config-file`; wildfire-related
  fields on the body are reachable through the payload. The
  `wildfire-verdict-change-requests` sub-resource
  (`ProjectsLocationsFirewallEndpointAssociationsWildfireVerdictChangeRequestsService`
  on the Go client) is a new subgroup and belongs to the same
  follow-up as the `ull-mirroring-*` groups above.
- **572.0.0** — `firewall-endpoints` project-scoping GA. Already
  project-scoped in gcloud-go via `addNSOrgFlags`; the org flag is
  optional.
- **571.0.0** (×6) — Project-scoping GA promotions for
  `security-profiles` (delete/describe/export/import/list),
  `security-profiles custom-intercept`,
  `security-profiles custom-mirroring`,
  `security-profiles threat-prevention` (including override
  helpers), `security-profiles url-filtering`, and
  `security-profile-groups`. gcloud-go's `security-profiles` and
  `security-profile-groups` (see `cmd/network_security_profiles.go`)
  are single project-scoped CRUD groups; per-type subgroups exist in
  the Python argparse hierarchy but gcloud-go exposes the same field
  surface through `--config-file` on the parent CRUD. Threat-prevention
  override helpers (`add-override`, `delete-override`,
  `list-overrides`, `update-override`) map to fields on the
  `SecurityProfile.threatPrevention.severityOverrides` /
  `threatOverrides` arrays; those are set through the payload today.
- **570.0.0** — `secure-access-connect` GA (attachments, realms).
  Already implemented in gcloud-go via the v1beta1 REST client (see
  `cmd/network_security_sac.go`, issue #835).
- **570.0.0** — Symantec integration / localization flags restricted
  to ALPHA/BETA. gcloud-go does not maintain per-track argparse
  gates; the underlying Symantec-integration fields on the
  security-profile body are always reachable via `--config-file`
  because they are not track-limited on the API side.

## Network Management 570.0.0–578.0.0 (#1813)

Two of the six items map onto fields the current
`connectivity-tests create` already accepts via `--config-from-file`, and
the other four are new subgroups gcloud-go has not implemented yet:

- **578.0.0** — `--source-dms-private-connection` on
  `network-management connectivity-tests`. Populates
  `Endpoint.dmsPrivateConnection` (a string field on the connectivity
  test's source/destination endpoint, exposed by
  `google.golang.org/api/networkmanagement/v1`). gcloud-go's
  `connectivity-tests create` (see `cmd/network_management.go`) takes
  the full `ConnectivityTest` body via `--config-from-file`, so the
  field is settable through the YAML/JSON payload today; adding a
  dedicated convenience flag is optional.
- **575.0.0** — `--source-cloud-run-job` on
  `network-management connectivity-tests`. Populates
  `Endpoint.cloudRunJob`; reachable via `--config-from-file` for the
  same reason.
- **570.0.0** — `network-management network-monitoring-providers
  monitoring-points` subgroup (agents that send probes). gcloud-go's
  `network-monitoring-providers` (see
  `cmd/network_management_network_monitoring_providers.go`) is served
  by a raw REST client (`netmgmtMonRest`) against the v1beta1 surface;
  it wires the provider-level commands (`create`, `delete`, `describe`,
  `generate-monitoring-point-config`, `generate-provider-access-token`,
  `list`) but not the nested `monitoring-points` sub-resource. Adding
  the subgroup is a distinct piece of work.
- **570.0.0** — Same story for the `network-paths` subgroup
  (hop-by-hop route and active delivery quality between a monitoring
  point and a destination).
- **570.0.0** — Same story for the `web-paths` subgroup (monitored web
  applications or URLs).
- **570.0.0** — `network-management network-monitoring-providers`
  top-level command group. Already implemented in gcloud-go at
  `cmd/network_management_network_monitoring_providers.go` (tracked
  under #954); only the three nested subgroups above remain.

## Network Connectivity 570.0.0–580.0.0 (#1812)

gcloud-go's `network-connectivity hubs` and `spokes` commands (see
`cmd/network_connectivity_hubs.go`) take the full `Hub` / `Spoke` body
via `--config-file`, and gcloud-go does not maintain separate
alpha/beta/GA tracks. That covers four of the five items:

- **580.0.0** —
  `--export-psc-published-services-and-regional-google-apis` and
  `--export-psc-global-google-apis` on `network-connectivity hubs
  create` and `update`. Both flags populate boolean fields on the
  `Hub.exportPsc*` surface (see
  `google.golang.org/api/networkconnectivity/v1.Hub`), settable via
  `--config-file` today.
- **573.0.0** — Updated `--region` help text on
  `network-connectivity transports create` (`registerNCTransports` in
  `cmd/network_connectivity_resources.go`) to list supported regions.
  gcloud-go's `--location` flag description does not enumerate valid
  values — that list is served by the API and changes over time; the
  Python help-text expansion is a Python-only refresh and adding a
  static enumeration in gcloud-go would go stale between releases.
- **570.0.0** — `network-connectivity spokes gateways` `create` / `update`
  GA promotion. gcloud-go's `spokes create` / `update` already accept
  the full `Spoke` body via `--config-file`; `Spoke.gateway`
  (`*Gateway`) is exposed by the Go client, so gateway spokes are
  creatable today by populating `gateway: {...}` in the YAML/JSON payload.
- **570.0.0** — `HYBRID_INSPECTION` preset topology on `hubs create`
  GA. `Hub.presetTopology` accepts `"HYBRID_INSPECTION"` via
  `--config-file`; no Go-side enum to update.

The remaining item is a legitimate new subgroup gcloud-go has not
implemented:

- **572.0.0** — `network-connectivity spokes gateways advertised-routes`
  subgroup GA (`create`, `delete`, `describe`, `list`). This is a
  distinct API surface
  (`ProjectsLocationsSpokesGatewayAdvertisedRoutesService` on the Go
  client), not a flag on an existing command. Adding the subgroup is a
  standalone piece of work (its own `registerNCSpokeGatewayAdvertisedRoutes`
  builder mirroring the existing `ncCRUD` pattern in
  `cmd/network_connectivity_hubs.go`) — tracked here for follow-up.

## Metastore 580.0.0 (#1811)

Both 580.0.0 items are on `gcloud beta metastore services migrations
start`, which is a subgroup that does not exist in gcloud-go. The
`metastore services` group in `cmd/metastore.go` currently wires
`create`, `delete`, `describe`, `list`, `update`, IAM, alter/move/query,
export, import, restore, and the nested `backups` subgroup — no
`migrations` subgroup — so both items land when the migrations subgroup
is added:

- **580.0.0** — Lakehouse runtime catalog support in
  `metastore services migrations start`. The `MigrationExecution` /
  `StartMigrationRequest` proto exposes the target-catalog fields the
  new flag populates; nothing to add until the migrations subgroup lands.
- **580.0.0** — Cloud SQL migration arguments deprecated in
  `metastore services migrations start`. gcloud-go never exposed the
  Cloud SQL migration flags, so there is no deprecation to mirror.

## Looker 576.0.0 (#1810)

Both 576.0.0 items are already reachable in gcloud-go without new flags
or output changes:

- **576.0.0** — `--release-channel` and
  `--accelerated-security-patch-enabled` GA promotion on
  `gcloud looker instances create` and `update`. In gcloud-go those
  commands take the full `looker.Instance` body via `--config-file` (see
  `cmd/looker.go`), and `google.golang.org/api/looker/v1` already
  exposes `Instance.releaseChannel` and
  `Instance.acceleratedSecurityPatchEnabled`, so both fields are
  settable today by populating them in the YAML/JSON payload.
  gcloud-go does not maintain separate alpha/beta/GA tracks, so the
  promotion has no track surface to sync.
- **576.0.0** — `RELEASE_CHANNEL` and `ACCELERATED_SECURITY_PATCH_ENABLED`
  columns added to `gcloud looker instances describe` output in GA.
  gcloud-python's describe uses a per-track projection that hides some
  fields on GA; gcloud-go's `runLKInstDescribe` (`cmd/looker.go`) calls
  `emitFormatted(got, flagLKFormat)` on the full `Instance` object, so
  users already see `releaseChannel` and
  `acceleratedSecurityPatchEnabled` (and every other resource field) in
  the default YAML output today.

## Cloud Functions 579.0.0 (#1779)

Both 579.0.0 items land in `gcloud functions deploy` / `gcloud functions
upgrade`, neither of which is implemented in gcloud-go (`deploy` is
registered as a `"Not yet implemented"` stub in `cmd/functions.go`, and
`upgrade` is not present at all, since only the IAM, event-types, logs,
regions, and runtimes subgroups are wired to real APIs today under
#860–#864 as part of #342). The two items should land alongside a real
implementation of `functions deploy` / `functions upgrade`:

- **579.0.0** — `all-traffic` added as an allowed value for
  `--direct-vpc-egress` on `functions deploy`. In the v2 Cloud Functions
  API this maps to `serviceConfig.vpcConnectorEgressSettings =
  "ALL_TRAFFIC"` — a value the Go client
  (`google.golang.org/api/cloudfunctions/v2`) already accepts on the
  `ServiceConfig` proto. Nothing to add until `functions deploy` gains a
  real create/update path.
- **579.0.0** — `functions upgrade` promoted to GA. gcloud-go has no
  `functions upgrade` command yet; the promotion is a Python-track
  change with no analogue in the Go binary.

## Cloud Firestore Emulator 577.0.0–579.0.0 (#1778)

gcloud-go's `emulators firestore` group is a stub (see `cmd/emulators.go`:
`registerStubGroup(emulatorsCmd, "firestore", ..., "start", "env-init")`)
because gcloud-go does not bundle the Java-based Cloud Firestore emulator
jar. Both upstream items are therefore not applicable to the current
gcloud-go binary and should land alongside a future implementation of
`emulators firestore start`:

- **579.0.0** — `--require-indexes` and `--index-file` flags on
  `gcloud emulators firestore start`. Both flags are Java-jar-side
  toggles that only make sense once gcloud-go actually launches the
  emulator process; they belong with the eventual `emulators firestore`
  implementation.
- **577.0.0** — Cloud Firestore emulator v1.22.0 bump (DML support for
  the Pipelines API, the same `--require-indexes`/`--index-file` flags,
  and a `<JRE 25` deprecation warning). Bundled-jar version bumps do
  not translate to gcloud-go, and the JRE-version warning is emitted by
  the jar at start-up — no CLI code to sync.

## Cloud Firestore 582.0.0 (#1777)

- **582.0.0** — Search shorthands on
  `gcloud beta firestore indexes composite create`. The Python change
  adds argparse convenience syntax (e.g. `--field=name,search=true`)
  that maps onto `GoogleFirestoreAdminV1IndexField.searchConfig` (a
  `GoogleFirestoreAdminV1SearchConfig`) inside the composite index body.
  gcloud-go's `firestore indexes composite create` command (see
  `cmd/firestore_fields_indexes.go`) takes the full
  `GoogleFirestoreAdminV1Index` via `--config-file`, and the Go client
  already exposes `IndexField.searchConfig` and
  `IndexField.vectorConfig`, so search-index configuration is settable
  today by populating `fields[i].searchConfig` in the YAML/JSON payload.
  gcloud-go also does not maintain a separate `beta` track — the
  gcloud-python beta-only surface lands under the single command in
  gcloud-go.

## Cloud Datastream 571.0.0–574.0.0 (#1776)

The `--sql-where-clause` flag on `gcloud datastream objects start-backfill`
(571.0.0) is implemented — it populates
`StartBackfillJobRequest.eventFilter.sqlWhereClause` on the outgoing
request; see `cmd/datastream_all.go`. The remaining three items don't
need CLI changes:

- **574.0.0** — Regional Endpoints (REP) support for all Datastream
  commands. gcloud-python routed its HTTP client through
  `<region>-datastream.googleapis.com` transparently for regional
  operations. gcloud-go relies on Google's global
  `datastream.googleapis.com` endpoint (fronted at the load-balancer
  layer for regional routing); the Go client library
  (`google.golang.org/api/datastream/v1`) does not require or benefit
  from per-region endpoint overrides for correctness. If future latency
  work calls for regional endpoints, `DatastreamService` in
  `internal/gcp/clients.go` can grow the same
  `option.WithEndpoint("https://<region>-datastream.googleapis.com/")`
  routing already used by `dataproc` and `aiplatform`.
- **573.0.0** — Dataverse, Salesforce Marketing Cloud, and ServiceNow
  connection-profile types on `datastream connection-profiles create`
  and `update`. gcloud-go's `connection-profiles create`/`update`
  commands take the full `ConnectionProfile` body via `--config-file`
  (see `cmd/datastream_all.go`), and
  `google.golang.org/api/datastream/v1` already exposes
  `ConnectionProfile.dataverseProfile`,
  `ConnectionProfile.salesforceMarketingCloudProfile`, and
  `ConnectionProfile.serviceNowProfile`, so all three source types are
  settable through the YAML/JSON payload today.
- **573.0.0** — Same three source types on `datastream streams create`
  and `update`. `Stream.sourceConfig.sourceConnectionProfile` references
  the connection-profile resource, and the corresponding source-specific
  fields on `SourceConfig` (`dataverseSourceConfig`,
  `salesforceMarketingCloudSourceConfig`, `serviceNowSourceConfig`) are
  exposed by the Go client; both `streams create` and `update` take the
  full `Stream` body via `--config-file`.

## Cloud Dataproc 569.0.0–579.0.0 (#1775)

gcloud-go's `dataproc clusters create/update`, `dataproc workflow-templates
set-managed-cluster`, and `dataproc batches submit` all take the full
`Cluster` / `ManagedCluster` / `Batch` body via `--config-file`
(`--source` for workflow templates — see `cmd/dataproc_clusters.go`,
`cmd/dataproc_workflow_templates.go`, `cmd/dataproc_batches.go`), so any
proto field exposed by `google.golang.org/api/dataproc/v1` is already
settable through the YAML/JSON payload. The five payload-reachable
items therefore don't need new CLI flags:

- **579.0.0** — GA promotion of `--master-instance-selection`,
  `--master-instance-flexibility-policy-file`, `--worker-instance-selection`,
  `--worker-instance-flexibility-policy-file`,
  `--secondary-worker-instance-selection`, and
  `--secondary-worker-instance-flexibility-policy-file` on `dataproc
  clusters create` and `dataproc workflow-templates set-managed-cluster`.
  All six populate `InstanceGroupConfig.instanceFlexibilityPolicy`
  (`instanceSelectionList` / `provisioningModelMix`) on the master,
  worker, or secondary-worker `InstanceGroupConfig` in the Cluster body.
  gcloud-go does not maintain separate alpha/beta/GA tracks and accepts
  the field directly from the payload today.
- **575.0.0** — `--master-machine-types` on `dataproc clusters create`.
  Populates `masterConfig.instanceFlexibilityPolicy.instanceSelectionList`
  entries (`{machineType, rank}`) — same field set as the flexibility
  policy above; reachable via `--config-file`.
- **571.0.0** — `--confidential-compute-type` on `dataproc clusters
  create`. Populates
  `gceClusterConfig.confidentialInstanceConfig.confidentialInstanceType`
  (`SEV` / `SEV_SNP` / `TDX`) which is exposed by the Go client's
  `ConfidentialInstanceConfig` struct.
- **571.0.0** — `--confidential-compute` deprecation on `dataproc
  clusters create`. gcloud-go never surfaced its own
  `--confidential-compute` flag; both the deprecated
  `enableConfidentialCompute` bool and the replacement
  `confidentialInstanceType` enum are set on the same
  `ConfidentialInstanceConfig` message and reached through `--config-file`.
  The deprecation is a Python-argparse warning with no Go analogue.
- **569.0.0** — `--resource-manager-tag` on `dataproc batches submit`.
  Populates `environmentConfig.executionConfig.resourceManagerTags` on
  the Batch body, which is already exposed by the Go client's
  `ExecutionConfig` struct.

The remaining two items don't map onto existing gcloud-go surfaces:

- **576.0.0** — GA promotion of `dataproc batches submit pyspark-notebook`.
  gcloud-go's `batches submit` takes the full `Batch` body via
  `--config-file`, and the Go client already exposes
  `Batch.pysparkNotebookBatch` (`*PySparkNotebookBatch`), so callers can
  submit a PySpark-notebook batch today by setting `pysparkNotebookBatch`
  in the YAML/JSON payload. gcloud-python's dedicated
  `submit pyspark-notebook` subcommand is a Python argparse convenience
  that maps a small set of positional/flag arguments to the same field;
  the GA promotion has no track distinction to mirror.
- **569.0.0** — `--resource-manager-tag` on `dataproc sessions create`.
  gcloud-go does not implement the `dataproc sessions` subgroup;
  `resourceManagerTags` on `Session.environmentConfig.executionConfig`
  should land as part of adding the sessions group.

## Cloud Backup DR 570.0.0–580.0.0 (#1769)

All four items add flags to `gcloud backup-dr backup-plans` and
`gcloud backup-dr backups restore` commands. gcloud-go's entire
`backup-dr` surface (all twelve sub-groups, including `backup-plans`,
`backups`, `management-servers`, etc.) is registered as `"not yet
implemented"` stubs in `cmd/backup_dr.go` (tracked under #303), so there
is no create-request/patch-request builder or restore surface for these
flags to hang off of yet:

- **580.0.0** — `boot-disk-only` and `disk-exclusion-labels` selective-disk
  backup properties under `--compute-instance-properties` in
  `backup-dr backup-plans create` / `update`.
- **580.0.0** — `--source-instance-boot-disk` and
  `--source-instance-disk-device-name` on `backup-dr backups restore disk`
  for individual disk restore from a compute instance backup.
- **578.0.0** — `--log-retention-days` across `backup-dr backup-plans
  create` / `update` for PITR log retention configuration.
- **570.0.0** — `--use-project-service-account` on
  `backup-dr backups restore compute` and
  `backup-dr backups restore disk`.

These four items should land alongside the eventual real implementation
of the `backup-dr` surface (#303); adding CLI flags to stubs that return
`"not yet implemented"` would only mislead callers.

## Cloud Auth 579.0.0–581.0.0 (#1768)

All three items add or enable Enterprise Certificate Proxy (ECP) HTTP
Proxy support in the Python auth stack. gcloud-go does not implement ECP
(the `auth enterprise-certificate-config` group is a stub — see
`registerStubGroup(authCmd, "enterprise-certificate-config", ...)` in
`cmd/auth.go`), and the Go binary uses the standard `crypto/tls` stack
plus `HTTP_PROXY`/`HTTPS_PROXY` environment variables for its transport,
so none of these translate to gcloud-go code today:

- **581.0.0** — ECP HTTP Proxy enabled by default for context-aware mTLS
  requests. gcloud-go has no ECP HTTP proxy toggle to flip.
- **579.0.0** — ECP HTTP Proxy support enabled for external users
  (disabled by default). Same reason: no ECP support to gate.
- **579.0.0** — `--ecp-http-proxy` flag on
  `gcloud auth enterprise-certificate-config create` to specify a custom
  ECP HTTP proxy binary path. `enterprise-certificate-config create` is a
  stub in gcloud-go, so the flag has no create-request builder to wire
  into. When ECP is implemented in gcloud-go, `--ecp-http-proxy` should be
  added alongside the rest of the create-request surface.

## Cloud Alerting 570.0.0 (#1767)

- **Fixing List Alerts `gcloud` CLI example commands** (570.0.0) — the
  upstream fix is entirely inside the Python argparse `examples` block of
  `surface/monitoring/policies/list.yaml`, refining the `--sort-by` and
  `--filter` snippets shown in `gcloud monitoring policies list --help`.
  gcloud-go's `monitoring policies list` command (see `cmd/monitoring.go`)
  provides its own short help text and does not carry the Python
  example strings, so there is nothing to sync into the Go binary. Users
  who want the refreshed example commands can consult the upstream
  [alerting policies filter/sort reference](https://cloud.google.com/monitoring/api/v3/sorting-and-filtering#alertpolicy).

## Access Context Manager 572.0.0–575.0.0 (#1766)

Both cloud-bindings items are already reachable through gcloud-go's
`--config-file` body and do not require new flags:

- **`--service-account` and `--service-account-project-number` exposed for
  `gcloud access-context-manager cloud-bindings` in the GA track**
  (575.0.0) — gcloud-python promoted two ALPHA-only argparse flags that map
  onto `GcpUserAccessBinding.principal.serviceAccount` and
  `GcpUserAccessBinding.principal.serviceAccountProjectNumber`. gcloud-go's
  `cloud-bindings create` and `update` commands (see `cmd/acm_cloud_bindings.go`)
  do not maintain separate alpha/beta/GA tracks and take the full
  `GcpUserAccessBinding` body via `--config-file`, so both fields are already
  settable today by populating `principal.serviceAccount` /
  `principal.serviceAccountProjectNumber` in the YAML or JSON payload.
- **`--group-key` on `gcloud access-context-manager cloud-bindings create`
  made optional in GA** (572.0.0) — Python dropped its argparse `required:
  true` marker so the GA track matches the alpha behaviour. gcloud-go never
  exposed `--group-key` as a flag; `groupKey` is one of the many
  `GcpUserAccessBinding` fields loaded from `--config-file`, so the field
  has always been optional in gcloud-go and no CLI change is required.

## BigLake 570.0.0–578.0.0 (#1761)

- **Track promotions (alpha→beta, beta→GA)** across 570.0.0–576.0.0 for
  `gcloud biglake hive catalogs/databases/tables`,
  `gcloud biglake delta-sharing <catalogs|shares|schemas|tables>`,
  `gcloud biglake data-product-sharing publish`,
  `gcloud biglake iceberg catalogs`, and the
  `--unity-service-principal-application-id` / federated unity catalog
  flags of `gcloud biglake iceberg catalogs`. gcloud-go does not maintain
  separate alpha/beta/GA tracks, so track-promotion release notes do not
  translate to code changes.
- **New flags/values on `gcloud biglake iceberg catalogs create/update`**
  across 570.0.0–576.0.0 (`--primary-location` promoted to GA, the
  `lakehouse` option on `--catalog=type`, and the federated unity catalog
  flags) — gcloud-go's `biglake iceberg catalogs create` and `update`
  accept the catalog body via `--config-file`, so setting
  `catalog-type: lakehouse`, `primary-location`, and the federated unity
  fields (including `unity-service-principal-application-id`) in the YAML
  or JSON payload is supported today.
- **`X-Iceberg-Access-Delegation: vended-credentials` header fix on
  `gcloud biglake iceberg tables`** (578.0.0),
  **`gcloud biglake hive tables create/update`** (575.0.0/576.0.0),
  **`gcloud biglake delta-sharing …`**, and
  **`gcloud biglake data-product-sharing publish`** — gcloud-go has not
  ported the `biglake iceberg tables`, `biglake iceberg namespaces`,
  `biglake hive`, `biglake delta-sharing`, or
  `biglake data-product-sharing` subgroups yet, so these upstream changes
  land automatically when those subgroups are added. The only BigLake
  surface gcloud-go currently exposes is `biglake iceberg catalogs`, which
  is unaffected by the header bug.
