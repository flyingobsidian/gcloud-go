package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
)

// flattenRecords applies a series of gcloud `--flatten` specs to v and returns
// the resulting list of records. Each spec is a dotted path with `[]` markers
// indicating levels where an inner array should be expanded into separate
// records; for example `bindings[].members` expands both `bindings` and
// `members` (the trailing array-typed field is expanded implicitly).
//
// When specs is empty, the function returns a single-element slice containing
// the JSON round-tripped form of v.
//
// The returned records are JSON-shaped (map[string]any), safe for
// emitFormatted to consume for --format="table(...)" and similar.
func flattenRecords(v any, specs []string) ([]map[string]any, error) {
	base, err := toJSONMap(v)
	if err != nil {
		return nil, err
	}
	records := []map[string]any{base}
	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}
		var next []map[string]any
		for _, rec := range records {
			expanded, err := expandSpec(rec, spec)
			if err != nil {
				return nil, err
			}
			next = append(next, expanded...)
		}
		records = next
	}
	return records, nil
}

// toJSONMap round-trips v through JSON into a map[string]any.
func toJSONMap(v any) (map[string]any, error) {
	if m, ok := v.(map[string]any); ok {
		return cloneMap(m), nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("flatten: value must be a JSON object, got %s", strings.TrimSpace(string(b)))
	}
	return m, nil
}

// expandSpec applies one flatten spec to a single record. Each `[]` marker in
// the spec, plus the trailing segment when it resolves to an array, becomes
// an expansion point. Segment paths are accumulated from the root, so
// `bindings[].members` expands `bindings` then `bindings.members` (matching
// gcloud-python's behaviour).
func expandSpec(rec map[string]any, spec string) ([]map[string]any, error) {
	points := flattenExpansionPoints(spec)
	if len(points) == 0 {
		return []map[string]any{rec}, nil
	}
	return expandAt(rec, points, 0)
}

// flattenExpansionPoints returns the cumulative root-relative paths at which
// arrays should be expanded. For `bindings[].members` the points are
// [`bindings`, `bindings.members`] — the trailing accessor is included so
// that a leaf array is also expanded, mirroring gcloud-python's behaviour.
func flattenExpansionPoints(spec string) [][]string {
	parts := strings.Split(spec, "[]")
	// Clean each segment: remove wrapping dots produced by adjacent `[]`.
	var cleaned []string
	for _, p := range parts {
		p = strings.Trim(p, ".")
		cleaned = append(cleaned, p)
	}
	// Build cumulative dotted paths. Skip empty accumulations.
	var points [][]string
	cum := ""
	for i, p := range cleaned {
		if p == "" {
			continue
		}
		if cum == "" {
			cum = p
		} else {
			cum = cum + "." + p
		}
		// A cleaned segment before the last position marks a `[]` split, so
		// its path is definitely an expansion point. The last cleaned segment
		// is also treated as an expansion point (implicit trailing `[]`); at
		// eval time we simply skip it if the value there isn't an array.
		_ = i
		points = append(points, strings.Split(cum, "."))
	}
	return points
}

// expandAt expands at points[i]. If the value at that path isn't an array we
// simply move on to the next expansion point without producing extra records.
func expandAt(rec map[string]any, points [][]string, i int) ([]map[string]any, error) {
	if i >= len(points) {
		return []map[string]any{rec}, nil
	}
	path := points[i]
	arr, ok := getPathArray(rec, path)
	if !ok {
		// Not an array (or path missing): don't split this record here.
		return expandAt(rec, points, i+1)
	}
	var out []map[string]any
	for _, elem := range arr {
		clone := cloneMap(rec)
		if err := setPath(clone, path, elem); err != nil {
			return nil, err
		}
		expanded, err := expandAt(clone, points, i+1)
		if err != nil {
			return nil, err
		}
		out = append(out, expanded...)
	}
	return out, nil
}

// getPathArray descends rec via path and returns the array found at the end.
// It returns (nil, false) if any intermediate node is missing or the final
// node isn't an array.
func getPathArray(rec map[string]any, path []string) ([]any, bool) {
	var cur any = rec
	for _, p := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil, false
		}
		v, ok := m[p]
		if !ok {
			return nil, false
		}
		cur = v
	}
	arr, ok := cur.([]any)
	return arr, ok
}

// setPath sets the value at path in rec to val, deep-cloning maps along the
// way so callers can safely modify siblings.
func setPath(rec map[string]any, path []string, val any) error {
	if len(path) == 0 {
		return fmt.Errorf("flatten: cannot set empty path")
	}
	cur := rec
	for _, p := range path[:len(path)-1] {
		inner, ok := cur[p].(map[string]any)
		if !ok {
			return fmt.Errorf("flatten: intermediate segment %q is not an object", p)
		}
		clone := cloneMap(inner)
		cur[p] = clone
		cur = clone
	}
	cur[path[len(path)-1]] = val
	return nil
}

// cloneMap returns a shallow copy of m. The tree is walked lazily as setPath
// descends so nested maps are cloned on demand.
func cloneMap(m map[string]any) map[string]any {
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
