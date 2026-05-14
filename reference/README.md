Here is where the Google Cloud SDK (Python) files are stored.

Current target version: **568.0.0**

# 568.0.0

diff between 565.0.0 and 568.0.0 showed that every command implemented is identical between versions. The only changes in 568 that touch our command scope either work automatically (freeform --purpose string on addresses create) or aren't available in the v1 Go API (Beta-only TERMINATION_TIMESTAMP column). No Go code changes were needed.

# 565.0.0

Initial version, converted from Python to Golang by Claude.
