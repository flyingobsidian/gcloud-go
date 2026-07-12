#!/usr/bin/env bash

# normalize_json: read JSON on stdin, strip volatile / server-assigned fields,
# sort object keys and array elements so semantically-equal payloads compare
# equal regardless of ordering or timestamps.
normalize_json() {
    jq -S '
        def scrub:
            walk(
                if type == "object" then
                    del(
                        .etag, .createTime, .create_time, .createdAt,
                        .updateTime, .update_time, .selfLink, .expireTime,
                        .timeCreated, .updated, .generation, .metageneration,
                        .uid, .oauth2ClientId, .id, .softDeleteTime,
                        .hardDeleteTime, .projectNumber, .updateMask
                    )
                else . end
            );
        scrub
        | walk(if type == "array" then sort_by(tojson) else . end)
    ' 2>/dev/null
}

# run_capture OUTFILE -- cmd...: run a command, tee stdout to OUTFILE and the
# log, capture stderr alongside. Returns the command's exit status.
run_capture() {
    local out="$1"; shift
    [ "$1" = "--" ] && shift
    printf '  $ %s\n' "$*"
    "$@" >"$out" 2>"${out}.err"
    local rc=$?
    sed 's/^/    stdout: /' "${out}"
    sed 's/^/    stderr: /' "${out}.err"
    return $rc
}

# _neutralize: read stdin, replace per-project identifiers with placeholders so
# output produced in two different projects can be compared. Handles project
# ids (GO_PROJECT/PY_PROJECT), numeric project numbers inside resource names
# (projects/<number>/...), and any extra tokens listed in $NEUTRALIZE (space
# separated, e.g. per-run bucket names).
_neutralize() {
    local -a sed_args=(-E -e 's#projects/[0-9]+/#projects/PROJECT/#g')
    [ -n "$GO_PROJECT" ] && sed_args+=(-e "s/${GO_PROJECT}/PROJECT/g")
    [ -n "$PY_PROJECT" ] && sed_args+=(-e "s/${PY_PROJECT}/PROJECT/g")
    local t
    for t in ${NEUTRALIZE:-}; do sed_args+=(-e "s/${t}/TOKEN/g"); done
    sed "${sed_args[@]}"
}

cmp_json() {
    local label="${1:?label is required}"
    local py_json="${2:?py_json is required}"
    local go_json="${3:?go_json is required}"
    _neutralize <"$py_json" | normalize_json >"${py_json}.norm"
    _neutralize <"$go_json" | normalize_json >"${go_json}.norm"
    if ! diff -u "${py_json}.norm" "${go_json}.norm" >"${go_json}.diff"; then
        echo "FAIL $label: normalised output differs (see diff below)"
        sed 's/^/    /' "${go_json}.diff"
        return 1
    fi
    return 0
}

cmp_text() {
    local label="${1:?label is required}"
    local py_txt="${2:?py_txt is required}"
    local go_txt="${3:?go_txt is required}"
    _neutralize <"$py_txt" >"${py_txt}.norm"
    _neutralize <"$go_txt" >"${go_txt}.norm"
    if ! diff -u "${py_txt}.norm" "${go_txt}.norm" >"${go_txt}.diff"; then
        echo "FAIL $label: listing differs (see diff below)"
        sed 's/^/    /' "${go_txt}.diff"
        return 1
    fi
    return 0
}
