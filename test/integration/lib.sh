#!/usr/bin/env bash

report() { # kind label detail
    local kind="$1" label="$2" detail="${3:-}"
    case "$kind" in
        PASS)
            PASS_COUNT=$((PASS_COUNT + 1)); RESULTS+=("PASS  $label")
            printf '%s[PASS]%s %s\n' "$_c_green" "$_c_reset" "$label"
            ;;
        FAIL)
            FAIL_COUNT=$((FAIL_COUNT + 1)); RESULTS+=("FAIL  $label")
            printf '%s[FAIL]%s %s -- %s\n' "$_c_red" "$_c_reset" "$label" "$detail"
            ;;
        GAP)
            GAP_COUNT=$((GAP_COUNT + 1));  RESULTS+=("GAP   $label")
            printf '%s[GAP ]%s %s -- %s\n' "$_c_yellow" "$_c_reset" "$label" "$detail"
            ;;
    esac
}

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

# cmp_json LABEL GO_JSON PY_JSON: neutralize, normalize and diff two JSON files.
cmp_json() {
    local label="$1" go_json="$2" py_json="$3"
    _neutralize <"$go_json" | normalize_json >"${go_json}.norm"
    _neutralize <"$py_json" | normalize_json >"${py_json}.norm"
    if diff -u "${py_json}.norm" "${go_json}.norm" >"${go_json}.diff"; then
        report PASS "$label"
    else
        report FAIL "$label" "normalized output differs (see diff below)"
        sed 's/^/    /' "${go_json}.diff" | head -n 40
    fi
}

# cmp_text LABEL GO_TXT PY_TXT: neutralize, sort and diff (order-insensitive).
cmp_text() {
    local label="$1" go_txt="$2" py_txt="$3"
    _neutralize <"$go_txt" | sort >"${go_txt}.norm"
    _neutralize <"$py_txt" | sort >"${py_txt}.norm"
    if diff -u "${py_txt}.norm" "${go_txt}.norm" >"${go_txt}.diff"; then
        report PASS "$label"
    else
        report FAIL "$label" "listing differs"
        sed 's/^/    /' "${go_txt}.diff" | head -n 40
    fi
}

print_summary() {
    echo
    echo "==================== SUMMARY ===================="
    printf '%s\n' "${RESULTS[@]}"
    echo "------------------------------------------------"
    printf 'PASS=%d  FAIL=%d  GAP=%d\n' "$PASS_COUNT" "$FAIL_COUNT" "$GAP_COUNT"
    echo "================================================"
    # A GAP (missing gcloud-go subcommand) is informational, not a failure.
    [ "$FAIL_COUNT" -eq 0 ]
}
