package cmd

import (
	"os"
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

func TestBuildSSHOptsWithExistingKey(t *testing.T) {
	// Create a temp key file so os.Stat succeeds.
	tmp := t.TempDir()
	keyFile := filepath.Join(tmp, "test_key")
	os.WriteFile(keyFile, []byte("fake"), 0600)

	opts := buildSSHOpts(keyFile)
	found := false
	for i, o := range opts {
		if o == "-i" && i+1 < len(opts) && opts[i+1] == keyFile {
			found = true
			break
		}
	}
	if !found {
		t.Error("buildSSHOpts with existing key should include -i <keyFile>")
	}

	// Should also include IdentitiesOnly when key exists.
	foundID := false
	for _, o := range opts {
		if o == "IdentitiesOnly=yes" {
			foundID = true
		}
	}
	if !foundID {
		t.Error("buildSSHOpts with existing key should include IdentitiesOnly=yes")
	}
}

func TestBuildSSHOptsMissingKey(t *testing.T) {
	opts := buildSSHOpts("/nonexistent/path/to/key")
	for i, o := range opts {
		if o == "-i" && i+1 < len(opts) {
			t.Error("buildSSHOpts with missing key should not include -i")
		}
	}
	for _, o := range opts {
		if o == "IdentitiesOnly=yes" {
			t.Error("buildSSHOpts with missing key should not include IdentitiesOnly=yes")
		}
	}
}
