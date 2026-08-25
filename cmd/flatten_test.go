package cmd

import (
	"reflect"
	"sort"
	"testing"
)

func TestFlattenBindingsMembers(t *testing.T) {
	// The exact shape from #1728.
	policy := map[string]any{
		"bindings": []any{
			map[string]any{
				"role":    "roles/A",
				"members": []any{"serviceAccount:sa1", "user:u1"},
			},
			map[string]any{
				"role":    "roles/B",
				"members": []any{"serviceAccount:sa1"},
			},
			map[string]any{
				"role":    "roles/C",
				"members": []any{"user:u2"},
			},
		},
	}

	rows, err := flattenRecords(policy, []string{"bindings[].members"})
	if err != nil {
		t.Fatalf("flattenRecords: %v", err)
	}

	if len(rows) != 4 {
		t.Fatalf("expected 4 flattened rows, got %d: %v", len(rows), rows)
	}

	// Each row should have `bindings.role` and `bindings.members` as scalars.
	type flat struct {
		role   string
		member string
	}
	got := make([]flat, 0, len(rows))
	for _, r := range rows {
		b, ok := r["bindings"].(map[string]any)
		if !ok {
			t.Fatalf("row missing bindings map: %v", r)
		}
		role, _ := b["role"].(string)
		member, _ := b["members"].(string)
		got = append(got, flat{role: role, member: member})
	}

	want := []flat{
		{"roles/A", "serviceAccount:sa1"},
		{"roles/A", "user:u1"},
		{"roles/B", "serviceAccount:sa1"},
		{"roles/C", "user:u2"},
	}
	sort.Slice(got, func(i, j int) bool {
		if got[i].role != got[j].role {
			return got[i].role < got[j].role
		}
		return got[i].member < got[j].member
	})
	sort.Slice(want, func(i, j int) bool {
		if want[i].role != want[j].role {
			return want[i].role < want[j].role
		}
		return want[i].member < want[j].member
	})
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestFlattenSingleLevel(t *testing.T) {
	obj := map[string]any{
		"bindings": []any{
			map[string]any{"role": "roles/A"},
			map[string]any{"role": "roles/B"},
		},
	}
	rows, err := flattenRecords(obj, []string{"bindings[]"})
	if err != nil {
		t.Fatalf("flattenRecords: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("want 2, got %d", len(rows))
	}
	for _, r := range rows {
		if _, ok := r["bindings"].(map[string]any); !ok {
			t.Errorf("bindings should be a map after flattening: %v", r)
		}
	}
}

func TestFlattenEmptySpec(t *testing.T) {
	obj := map[string]any{"a": 1}
	rows, err := flattenRecords(obj, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
}

func TestFlattenMissingArray(t *testing.T) {
	obj := map[string]any{"other": "x"}
	rows, err := flattenRecords(obj, []string{"bindings[]"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// Missing/non-array path: no expansion, record is emitted unchanged so
	// callers can still --filter it out.
	if len(rows) != 1 {
		t.Fatalf("want 1 row for missing array, got %d", len(rows))
	}
}
