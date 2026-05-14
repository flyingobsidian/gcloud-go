#!/usr/bin/env bash

get_source() {
    local version="${1:?Version is required}"

    local artifact_name
    local gcs_artifact_path
    artifact_name="google-cloud-sdk-$version-linux-x86_64.tar.gz"
    gcs_artifact_path="gs://cloud-sdk-release/$artifact_name"
    gsutil -mq cp "$gcs_artifact_path" .
}

get_source "568.0.0"
