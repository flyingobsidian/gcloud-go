Here is where the Google Cloud SDK (Python) files are stored.

Current target version: **582.0.0**

# 582.0.0 - Aug 2026

See `google-cloud-sdk-{VERSION}/RELEASE_NOTES` for details.

# 568.0.0 - May 2026

diff between 565.0.0 and 568.0.0 showed that every command implemented is identical between versions. The only changes in 568 that touch our command scope either work automatically (freeform --purpose string on addresses create) or aren't available in the v1 Go API (Beta-only TERMINATION_TIMESTAMP column). No Go code changes were needed.

# 565.0.0 - May 2026

Initial version, converted from Python to Golang by Claude.
