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
