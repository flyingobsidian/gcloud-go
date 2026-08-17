package gfilter

import "testing"

func TestCompileEmpty(t *testing.T) {
	f, err := Compile("")
	if err != nil {
		t.Fatalf("Compile(\"\"): %v", err)
	}
	if !f.Match(map[string]any{"x": 1}) {
		t.Error("empty filter should match everything")
	}
}

func TestMatchColonSubstring(t *testing.T) {
	f, err := Compile("name:vm")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if !f.Match(map[string]any{"name": "myvm-1"}) {
		t.Error("expected match")
	}
	if f.Match(map[string]any{"name": "other"}) {
		t.Error("expected no match")
	}
}

func TestMatchArrayTraversalColon(t *testing.T) {
	// The bug from #1725: networkInterfaces.subnetwork:s should match if any
	// of the network interfaces' subnetwork strings contain "s".
	instance := map[string]any{
		"name": "vm1",
		"networkInterfaces": []any{
			map[string]any{"subnetwork": "projects/p/regions/r/subnetworks/mysub"},
		},
	}
	f, err := Compile("networkInterfaces.subnetwork:s")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if !f.Match(instance) {
		t.Error("expected match for networkInterfaces.subnetwork:s")
	}
	if !f.UsesContainsOperator() {
		t.Error("expected UsesContainsOperator() to be true")
	}
}

func TestMatchEquality(t *testing.T) {
	f, _ := Compile("status=RUNNING")
	if !f.Match(map[string]any{"status": "RUNNING"}) {
		t.Error("expected match on =")
	}
	if f.Match(map[string]any{"status": "STOPPED"}) {
		t.Error("expected no match on =")
	}
	g, _ := Compile("status!=RUNNING")
	if g.Match(map[string]any{"status": "RUNNING"}) {
		t.Error("expected no match on !=")
	}
	if !g.Match(map[string]any{"status": "STOPPED"}) {
		t.Error("expected match on !=")
	}
}

func TestMatchBooleanCombinators(t *testing.T) {
	// implicit AND
	f, err := Compile("status=RUNNING name:vm")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if !f.Match(map[string]any{"status": "RUNNING", "name": "vm1"}) {
		t.Error("expected implicit AND match")
	}
	if f.Match(map[string]any{"status": "STOPPED", "name": "vm1"}) {
		t.Error("expected implicit AND miss on status")
	}

	// explicit OR
	g, err := Compile("status=RUNNING OR status=STOPPED")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if !g.Match(map[string]any{"status": "STOPPED"}) {
		t.Error("expected OR match")
	}
	if g.Match(map[string]any{"status": "PENDING"}) {
		t.Error("expected OR miss")
	}

	// NOT
	h, err := Compile("NOT status=RUNNING")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if h.Match(map[string]any{"status": "RUNNING"}) {
		t.Error("expected NOT to invert")
	}
	if !h.Match(map[string]any{"status": "OTHER"}) {
		t.Error("expected NOT to pass")
	}
}

func TestPrecedenceAndParens(t *testing.T) {
	// (a OR b) AND c
	f, err := Compile("(status=A OR status=B) AND name:vm")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if !f.Match(map[string]any{"status": "A", "name": "vm1"}) {
		t.Error("expected (A|B)&name to match A+vm1")
	}
	if f.Match(map[string]any{"status": "C", "name": "vm1"}) {
		t.Error("expected (A|B)&name to miss C+vm1")
	}
	if f.Match(map[string]any{"status": "A", "name": "other"}) {
		t.Error("expected (A|B)&name to miss A+other")
	}
}

func TestNumericCompare(t *testing.T) {
	f, err := Compile("cpuPlatform>=8")
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if !f.Match(map[string]any{"cpuPlatform": 16}) {
		t.Error("expected 16>=8")
	}
	if f.Match(map[string]any{"cpuPlatform": 4}) {
		t.Error("expected 4>=8 to be false")
	}
}

func TestRegex(t *testing.T) {
	f, err := Compile(`name~^vm[0-9]+$`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if !f.Match(map[string]any{"name": "vm123"}) {
		t.Error("expected regex match")
	}
	if f.Match(map[string]any{"name": "server"}) {
		t.Error("expected regex miss")
	}
	g, _ := Compile(`name!~^vm`)
	if g.Match(map[string]any{"name": "vm1"}) {
		t.Error("expected !~ to reject vm1")
	}
	if !g.Match(map[string]any{"name": "other"}) {
		t.Error("expected !~ to allow other")
	}
}

func TestQuotedValues(t *testing.T) {
	f, err := Compile(`name:"my vm"`)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if !f.Match(map[string]any{"name": "this-is-my vm-here"}) {
		t.Error("expected quoted value with space to match")
	}
}
