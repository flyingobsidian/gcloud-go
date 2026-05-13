package cmd

import "testing"

func TestParseSCPTarget(t *testing.T) {
	tests := []struct {
		input        string
		wantUser     string
		wantInstance string
		wantPath     string
		wantRemote   bool
	}{
		{"/local/file.txt", "", "", "/local/file.txt", false},
		{"my-vm:/remote/path", "", "my-vm", "/remote/path", true},
		{"user@my-vm:/remote/path", "user", "my-vm", "/remote/path", true},
		{"my-vm:file.txt", "", "my-vm", "file.txt", true},
		{"./relative/path", "", "", "./relative/path", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseSCPTarget(tt.input)
			if got.User != tt.wantUser {
				t.Errorf("user = %q, want %q", got.User, tt.wantUser)
			}
			if got.Instance != tt.wantInstance {
				t.Errorf("instance = %q, want %q", got.Instance, tt.wantInstance)
			}
			if got.Path != tt.wantPath {
				t.Errorf("path = %q, want %q", got.Path, tt.wantPath)
			}
			if got.IsRemote != tt.wantRemote {
				t.Errorf("isRemote = %v, want %v", got.IsRemote, tt.wantRemote)
			}
		})
	}
}

func TestFormatSCPArg(t *testing.T) {
	tests := []struct {
		name   string
		target scpTarget
		host   string
		want   string
	}{
		{
			"local path",
			scpTarget{Path: "/local/file.txt"},
			"10.0.0.1",
			"/local/file.txt",
		},
		{
			"remote no user",
			scpTarget{Instance: "my-vm", Path: "/remote/file.txt", IsRemote: true},
			"localhost",
			"localhost:/remote/file.txt",
		},
		{
			"remote with user",
			scpTarget{User: "admin", Instance: "my-vm", Path: "/remote/file.txt", IsRemote: true},
			"localhost",
			"admin@localhost:/remote/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatSCPArg(tt.target, tt.host)
			if got != tt.want {
				t.Errorf("formatSCPArg() = %q, want %q", got, tt.want)
			}
		})
	}
}
