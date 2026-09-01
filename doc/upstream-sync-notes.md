# Upstream sync notes

Notes on gcloud-python release-notes items whose upstream implementation does
not translate to gcloud-go changes, or is served by an existing gcloud-go
feature. Recording them here keeps the "Sync ..." tracking issues actionable
by making it clear when a checkbox needs no code change.

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
