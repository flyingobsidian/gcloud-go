#!/usr/bin/env bash

src_base="gs://cloud-sdk-release"

get_latest_version() {
    gsutil ls "$src_base/" \
        | awk -F- '/-linux-x86.tar.gz/ {print $6}' \
        | sort -n | tail -1
}

get_source() {
    local version="${1:?Version is required}"

    local artifact_name
    artifact_name="google-cloud-sdk-$version-linux-x86_64.tar.gz"
    if [[ -f "$artifact_name" ]]; then
        echo "File $artifact_name already exists. Skipping download." >&2
        return
    fi
    gsutil -mq cp "$src_base/$artifact_name" .
}

latest_version="$(get_latest_version)"
echo "Latest version: $latest_version"

v="582.0.0"

if [[ "$v" != "$latest_version" ]]; then
    echo "Warning: $v is not the latest version. Latest version is $latest_version." >&2
    exit 1
fi

get_source "$v"
