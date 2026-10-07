package cmd

import (
	"reflect"
	"testing"
)

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

func TestParseSCPArgs(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantSrcs int
		wantErr  bool
	}{
		{"single local to remote", []string{"a.txt", "vm:"}, 1, false},
		{"multiple local to remote", []string{"file1.txt", "file2.txt", "file3.txt", "user@vm:"}, 3, false},
		{"single remote to local", []string{"vm:a.txt", "."}, 1, false},
		{"multiple remote to local", []string{"user@vm:a.txt", "user@vm:b.txt", "."}, 2, false},
		{"remote source with remote destination", []string{"a.txt", "vm:b.txt", "vm:"}, 0, true},
		{"all local", []string{"a.txt", "b.txt"}, 0, true},
		{"mixed sources to local", []string{"vm:a.txt", "b.txt", "."}, 0, true},
		{"different instances to local", []string{"vm1:a.txt", "vm2:b.txt", "."}, 0, true},
		{"different users to local", []string{"alice@vm:a.txt", "bob@vm:b.txt", "."}, 0, true},
		{"too few args", []string{"a.txt"}, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcs, dst, err := parseSCPArgs(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseSCPArgs(%q) succeeded, want error", tt.args)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseSCPArgs(%q): %v", tt.args, err)
			}
			if len(srcs) != tt.wantSrcs {
				t.Errorf("got %d sources, want %d", len(srcs), tt.wantSrcs)
			}
			if want := parseSCPTarget(tt.args[len(tt.args)-1]); dst != want {
				t.Errorf("dst = %+v, want %+v", dst, want)
			}
		})
	}
}

func TestSCPTargetArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		remoteUser string
		want       []string
	}{
		{
			"multiple local to remote",
			[]string{"/path/file1.txt", "/path/file2.txt", "vm:"},
			"alice",
			[]string{"/path/file1.txt", "/path/file2.txt", "alice@localhost:"},
		},
		{
			"multiple remote to local, user resolved later",
			[]string{"vm:a.txt", "vm:b.txt", "."},
			"alice_example_com",
			[]string{"alice_example_com@localhost:a.txt", "alice_example_com@localhost:b.txt", "."},
		},
		{
			"no user",
			[]string{"a.txt", "vm:/tmp/"},
			"",
			[]string{"a.txt", "localhost:/tmp/"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srcs, dst, err := parseSCPArgs(tt.args)
			if err != nil {
				t.Fatalf("parseSCPArgs(%q): %v", tt.args, err)
			}
			got := scpTargetArgs(srcs, dst, tt.remoteUser, "localhost")
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("scpTargetArgs() = %q, want %q", got, tt.want)
			}
		})
	}
}
