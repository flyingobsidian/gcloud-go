package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestParseDiskSizeGB(t *testing.T) {
	cases := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{"200", 200, false},
		{"200GB", 200, false},
		{"200gb", 200, false},
		{" 200 GB ", 200, false},
		{"1TB", 1024, false},
		{"1tb", 1024, false},
		{"+50GB", 50, false},
		{"0", 0, true},
		{"-10GB", 0, true},
		{"", 0, true},
		{"abc", 0, true},
		{"200MB", 0, true}, // unsupported unit
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got, err := parseDiskSizeGB(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error, got nil (result=%d)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("parseDiskSizeGB(%q) = %d, want %d", tc.in, got, tc.want)
			}
		})
	}
}

// TestComputeDisksResizeRegistered guards the fix for #1727: the `resize`
// subcommand must exist under `compute disks` and accept a `--size` flag.
func TestComputeDisksResizeRegistered(t *testing.T) {
	var compute *cobra.Command
	for _, c := range rootCmd.Commands() {
		if c.Name() == "compute" {
			compute = c
			break
		}
	}
	if compute == nil {
		t.Fatal("compute command not registered")
	}
	var disks *cobra.Command
	for _, c := range compute.Commands() {
		if c.Name() == "disks" {
			disks = c
			break
		}
	}
	if disks == nil {
		t.Fatal("compute disks group not registered")
	}
	var resize *cobra.Command
	for _, c := range disks.Commands() {
		if c.Name() == "resize" {
			resize = c
			break
		}
	}
	if resize == nil {
		t.Fatal("compute disks resize command missing (#1727)")
	}
	if resize.Flags().Lookup("size") == nil {
		t.Error("compute disks resize should expose --size")
	}
	if resize.Flags().Lookup("zone") == nil {
		t.Error("compute disks resize should expose --zone")
	}
	if resize.Flags().Lookup("region") == nil {
		t.Error("compute disks resize should expose --region")
	}
	// One or more positional DISK_NAMEs required.
	if err := resize.Args(resize, []string{}); err == nil {
		t.Error("resize should reject zero positional args")
	}
	if err := resize.Args(resize, []string{"d1"}); err != nil {
		t.Errorf("resize should accept one positional arg, got: %v", err)
	}
	if err := resize.Args(resize, []string{"d1", "d2"}); err != nil {
		t.Errorf("resize should accept multiple positional args, got: %v", err)
	}
}
