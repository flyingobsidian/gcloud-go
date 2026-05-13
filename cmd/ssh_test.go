package cmd

import (
	"path/filepath"
	"testing"

	"google.golang.org/api/compute/v1"
)

func TestParseUserInstance(t *testing.T) {
	tests := []struct {
		input        string
		wantUser     string
		wantInstance string
	}{
		{"my-vm", "", "my-vm"},
		{"user@my-vm", "user", "my-vm"},
		{"root@instance-1", "root", "instance-1"},
		{"@instance", "", "instance"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			user, instance := parseUserInstance(tt.input)
			if user != tt.wantUser {
				t.Errorf("user = %q, want %q", user, tt.wantUser)
			}
			if instance != tt.wantInstance {
				t.Errorf("instance = %q, want %q", instance, tt.wantInstance)
			}
		})
	}
}

func TestGetExternalIP(t *testing.T) {
	inst := &compute.Instance{
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				NetworkIP: "10.0.0.1",
				AccessConfigs: []*compute.AccessConfig{
					{NatIP: "35.200.1.1"},
				},
			},
		},
	}

	got := getExternalIP(inst)
	if got != "35.200.1.1" {
		t.Errorf("getExternalIP() = %q, want %q", got, "35.200.1.1")
	}
}

func TestGetExternalIPNoPublic(t *testing.T) {
	inst := &compute.Instance{
		NetworkInterfaces: []*compute.NetworkInterface{
			{NetworkIP: "10.0.0.1"},
		},
	}

	got := getExternalIP(inst)
	if got != "" {
		t.Errorf("getExternalIP() = %q, want empty", got)
	}
}

func TestGetInternalIP(t *testing.T) {
	inst := &compute.Instance{
		NetworkInterfaces: []*compute.NetworkInterface{
			{NetworkIP: "10.128.0.5"},
		},
	}

	got := getInternalIP(inst)
	if got != "10.128.0.5" {
		t.Errorf("getInternalIP() = %q, want %q", got, "10.128.0.5")
	}
}

func TestGetInternalIPEmpty(t *testing.T) {
	inst := &compute.Instance{}

	got := getInternalIP(inst)
	if got != "" {
		t.Errorf("getInternalIP() = %q, want empty", got)
	}
}

func TestBuildSSHOpts(t *testing.T) {
	// With explicit key file.
	opts := buildSSHOpts("/path/to/key")
	found := false
	for i, o := range opts {
		if o == "-i" && i+1 < len(opts) && opts[i+1] == "/path/to/key" {
			found = true
			break
		}
	}
	if !found {
		t.Error("buildSSHOpts('/path/to/key') should include -i /path/to/key")
	}

	// Without explicit key file, should default to google_compute_engine.
	opts2 := buildSSHOpts("")
	foundDefault := false
	for i, o := range opts2 {
		if o == "-i" && i+1 < len(opts2) {
			if filepath.Base(opts2[i+1]) == "google_compute_engine" {
				foundDefault = true
			}
			break
		}
	}
	if !foundDefault {
		t.Error("buildSSHOpts('') should default to google_compute_engine key")
	}

	// Should include UserKnownHostsFile.
	foundKH := false
	for _, o := range opts2 {
		if len(o) > len("UserKnownHostsFile=") && o[:len("UserKnownHostsFile=")] == "UserKnownHostsFile=" {
			if filepath.Base(o[len("UserKnownHostsFile="):]) == "google_compute_known_hosts" {
				foundKH = true
			}
		}
	}
	if !foundKH {
		t.Error("buildSSHOpts() should include UserKnownHostsFile=...google_compute_known_hosts")
	}
}
