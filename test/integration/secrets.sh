#!/usr/bin/env bash

gcloud_py=gcloud
gcloud_go="$PWD/bin/gcloud-go"
py_project="${PY_PROJECT:?PY_PROJECT is required}"
go_project="${GO_PROJECT:?GO_PROJECT is required}"
object_uid="${OBJECT_UID:?OBJECT_UID is required}"
tmp_dir="$(mktemp -d -t "gcloud-secrets-$object_uid.XXXXXX")"
py_out="$tmp_dir/py.out"
go_out="$tmp_dir/go.out"
data_file="$tmp_dir/data.txt"

cleanup() { rm -rf "$tmp_dir"; }
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
trap 'exit 129' HUP

set -euo pipefail
cd "$(dirname "$(realpath "${BASH_SOURCE[0]}")")/../.." || exit 1
source test/integration/lib.sh || exit 1

create_secrets() {
    local pargs
    echo -n "super-secret" >"$data_file"
    pargs=(
        "secrets"
        "create"
        "test-secret-$object_uid"
        "--data-file=$data_file"
        "--labels=aaa=bbb"
    )
    run_capture "$py_out" "$gcloud_py" --project="$py_project" "${pargs[@]}"
    run_capture "$go_out" "$gcloud_go" --project="$go_project" "${pargs[@]}"

    # TBD: Check results
}

delete_secrets() {
    local nochecks="${1:-}"
    local pargs
    pargs=(
        "secrets"
        "delete"
        "test-secret-$object_uid"
    )
    run_capture "$py_out" "$gcloud_py" --project="$py_project" "${pargs[@]}"
    run_capture "$go_out" "$gcloud_go" --project="$go_project" "${pargs[@]}"

    if [[ "$nochecks" == "nochecks" ]]; then
        return 0
    fi

    # TBD: Check results
}

process() {
    local py_out go_out

    echo ">>> secrets"

    create_secrets

    # List secrets

    # Get secrets

    # Add secret version

    # List secret versions

    # Get secret version

    # Delete secret version

    delete_secrets

}

tidy() {
    delete_secrets nochecks
}

if [[ "${1:-}" == tidy ]]; then
    set +e
    tidy
else
    process
fi
