package cli

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	crm "google.golang.org/api/cloudresourcemanager/v3"
)

// Verifies the heading is omitted for an empty list, matching gcloud-python's
// TablePrinter (#1986).
func TestPrintTableHeadingEmpty(t *testing.T) {
	var b bytes.Buffer
	printTableHeadingTo(&b, 0, "%-40s %s\n", "NAME", "STATUS")
	if out := b.String(); out != "" {
		t.Errorf("printTableHeadingTo(0, ...) printed %q, want nothing", out)
	}
}

func TestPrintTableHeadingNonEmpty(t *testing.T) {
	var b bytes.Buffer
	printTableHeadingTo(&b, 2, "%-10s %s\n", "NAME", "STATUS")
	if out, want := b.String(), "NAME       STATUS\n"; out != want {
		t.Errorf("printTableHeadingTo(2, ...) printed %q, want %q", out, want)
	}
}

func TestPrintListResultsDefaultTableEmpty(t *testing.T) {
	var all []*crm.TagKey
	inv, stdout, _ := newTestInvocation(t, nil, "")
	if err := inv.printListResults(all, "", func() { printTagKeyTable(inv.stdout, all) }); err != nil {
		t.Fatalf("printListResults: %v", err)
	}
	out := stdout.String()
	if out != "" {
		t.Errorf("empty default table printed %q, want nothing", out)
	}
}

func TestPrintListResultsDefaultTableRows(t *testing.T) {
	all := []*crm.TagKey{{Name: "tagKeys/123", ShortName: "env", NamespacedName: "123/env"}}
	inv, stdout, _ := newTestInvocation(t, nil, "")
	if err := inv.printListResults(all, "", func() { printTagKeyTable(inv.stdout, all) }); err != nil {
		t.Fatalf("printListResults: %v", err)
	}
	out := stdout.String()
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %d lines, want heading plus one row:\n%s", len(lines), out)
	}
	if !strings.HasPrefix(lines[0], "NAME") || !strings.Contains(lines[0], "SHORT_NAME") {
		t.Errorf("heading = %q, want NAME/SHORT_NAME/NAMESPACED_NAME", lines[0])
	}
	if !strings.HasPrefix(lines[1], "tagKeys/123") {
		t.Errorf("row = %q, want it to start with tagKeys/123", lines[1])
	}
}

func TestPrintListResultsEmptyJSON(t *testing.T) {
	var all []*crm.TagKey
	inv, stdout, _ := newTestInvocation(t, nil, "")
	if err := inv.printListResults(all, "json", func() { printTagKeyTable(inv.stdout, all) }); err != nil {
		t.Fatalf("printListResults: %v", err)
	}
	out := stdout.String()
	if strings.TrimSpace(out) != "[]" {
		t.Errorf("empty json printed %q, want []", out)
	}
}

// Guards against new hand-rolled default tables that print their heading
// unconditionally: a fmt.Printf/Println of an upper-case heading immediately
// followed by a range loop over the rows must use printTableHeadingTo instead.
func TestNoUnguardedTableHeadings(t *testing.T) {
	heading := regexp.MustCompile(`^\s*fmt\.(Printf|Println)\(.*"[A-Z][A-Z_ ]+"`)
	loop := regexp.MustCompile(`^\s*for\s.*\brange\b`)

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		fh, err := os.Open(f)
		if err != nil {
			t.Fatal(err)
		}
		var prev string
		n := 0
		sc := bufio.NewScanner(fh)
		sc.Buffer(make([]byte, 1024*1024), 1024*1024)
		for sc.Scan() {
			n++
			line := sc.Text()
			if loop.MatchString(line) && heading.MatchString(prev) {
				t.Errorf("%s:%d: table heading printed without printTableHeadingTo: %s", f, n-1, strings.TrimSpace(prev))
			}
			prev = line
		}
		fh.Close()
		if err := sc.Err(); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
	}
}
