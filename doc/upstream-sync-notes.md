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
