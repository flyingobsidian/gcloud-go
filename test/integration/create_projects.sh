#!/usr/bin/env bash

# Create the two per-run projects for the live integration suite:
#   gcloud-go-<id>  created with gcloud-go
#   gcloud-py-<id>  created with the reference Python gcloud

wait_active() {
    local gcloud="${1:?Gcloud binary is required}"
    local proj="${2:?Project ID is required}"

    local i state
    echo -n ">>> ${proj}: Waiting for state==ACTIVE... "
    for ((i = 1; i <= 10; i++)); do
        state="$("$gcloud" projects describe "$proj" --format='value(lifecycleState)' 2>/dev/null)"
        [ "$state" = "ACTIVE" ] && { echo "$state"; return 0; }
        echo -n "$state "
        sleep 5
    done
    echo
    echo "ERROR: ${proj} did not become ACTIVE" >&2
    return 1
}

set -uo pipefail

cd "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/../.." || exit 1

gcloud_py="gcloud"
gcloud_go="$(pwd)/bin/gcloud-go"

ci_project_id="$(uuidgen | cut -d- -f1)"
py_project="gcloud-py-${ci_project_id}"
go_project="gcloud-go-${ci_project_id}"

echo ">>> Creating ${py_project} with gcloud (Python)"
"$gcloud_py" projects create "$py_project"
wait_active "$gcloud_py" "$py_project" || exit 1

echo ">>> Creating ${go_project} with gcloud-go"
"$gcloud_go" projects create "$go_project"
wait_active "$gcloud_go" "$go_project" || exit 1

if [ -n "${GCP_BILLING_ACCOUNT_ID:-}" ]; then
    "$gcloud_py" billing projects link "$py_project" --billing-account="$GCP_BILLING_ACCOUNT_ID" --quiet
    "$gcloud_go" billing projects link "$go_project" --billing-account="$GCP_BILLING_ACCOUNT_ID" --quiet
fi

if [ -n "${GITHUB_OUTPUT:-}" ]; then
  {
    echo "ci_project_id=${ci_project_id}"
    echo "py_project=${py_project}"
    echo "go_project=${go_project}"
  } >> "$GITHUB_OUTPUT"
fi
