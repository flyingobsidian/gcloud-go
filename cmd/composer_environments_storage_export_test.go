package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestComposerPathInsideDir(t *testing.T) {
	dest := t.TempDir()
	outside := t.TempDir()
	if err := os.Symlink(outside, filepath.Join(dest, "link")); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		rel  string
		want bool
	}{
		{"dag.py", true},
		{"sub/dir/dag.py", true},
		{"../escape.py", false},
		{"sub/../../escape.py", false},
		{"link/evil.py", false},
		{".", false},
	}
	for _, c := range cases {
		if got := composerPathInsideDir(filepath.Join(dest, c.rel), dest); got != c.want {
			t.Errorf("%s: got %v, want %v", c.rel, got, c.want)
		}
	}
}
