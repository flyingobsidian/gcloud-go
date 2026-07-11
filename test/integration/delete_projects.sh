#!/usr/bin/env bash

# Delete the two per-run projects for the live integration suite:
#   gcloud-go-<id>
#   gcloud-py-<id>

set -uo pipefail

cd "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/../.." || exit 1

gcloud_py="gcloud"
gcloud_go="$(pwd)/bin/gcloud-go"

py_project="${PY_PROJECT:?PY_PROJECT is required}"
go_project="${GO_PROJECT:?GO_PROJECT is required}"

echo ">>> Deleting ${py_project} with gcloud (Python)"
"$gcloud_py" projects delete "$py_project"

echo ">>> Deleting ${go_project} with gcloud-go"
"$gcloud_go" projects delete "$go_project"
