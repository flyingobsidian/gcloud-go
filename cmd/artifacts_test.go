package cmd

import "testing"

func TestParseDpkgOutput(t *testing.T) {
	output := "base-files\t12.4+deb12u5\tamd64\nlibc6\t2.36-9+deb12u4\tamd64\n"
	pkgs := parseDpkgOutput(output)
	if len(pkgs) != 2 {
		t.Fatalf("got %d packages, want 2", len(pkgs))
	}
	if pkgs[0].Package != "base-files" || pkgs[0].Version != "12.4+deb12u5" {
		t.Errorf("pkg[0] = %s %s, want base-files 12.4+deb12u5", pkgs[0].Package, pkgs[0].Version)
	}
	if pkgs[0].PackageType != "DEBIAN" {
		t.Errorf("pkg[0].PackageType = %q, want DEBIAN", pkgs[0].PackageType)
	}
	if pkgs[0].Architecture != "amd64" {
		t.Errorf("pkg[0].Architecture = %q, want amd64", pkgs[0].Architecture)
	}
}

func TestParseRpmOutput(t *testing.T) {
	output := "bash\t5.2.15-3.fc38\tx86_64\ncurl\t8.0.1-2.fc38\tx86_64\n"
	pkgs := parseRpmOutput(output)
	if len(pkgs) != 2 {
		t.Fatalf("got %d packages, want 2", len(pkgs))
	}
	if pkgs[0].Package != "bash" || pkgs[0].Version != "5.2.15-3.fc38" {
		t.Errorf("pkg[0] = %s %s, want bash 5.2.15-3.fc38", pkgs[0].Package, pkgs[0].Version)
	}
	if pkgs[0].PackageType != "RPM" {
		t.Errorf("pkg[0].PackageType = %q, want RPM", pkgs[0].PackageType)
	}
}

func TestParseApkOutput(t *testing.T) {
	output := "alpine-baselayout-3.4.3-r1\nbusybox-1.36.1-r2\nmusl-1.2.4-r2\n"
	pkgs := parseApkOutput(output)
	if len(pkgs) != 3 {
		t.Fatalf("got %d packages, want 3", len(pkgs))
	}
	if pkgs[0].Package != "alpine-baselayout" || pkgs[0].Version != "3.4.3-r1" {
		t.Errorf("pkg[0] = %q %q, want alpine-baselayout 3.4.3-r1", pkgs[0].Package, pkgs[0].Version)
	}
	if pkgs[0].PackageType != "APK" {
		t.Errorf("pkg[0].PackageType = %q, want APK", pkgs[0].PackageType)
	}
}

func TestSplitApkPackage(t *testing.T) {
	tests := []struct {
		input    string
		wantName string
		wantVer  string
	}{
		{"busybox-1.36.1-r2", "busybox", "1.36.1-r2"},
		{"alpine-baselayout-3.4.3-r1", "alpine-baselayout", "3.4.3-r1"},
		{"musl-1.2.4-r2", "musl", "1.2.4-r2"},
		{"libcrypto3-3.1.4-r5", "libcrypto3", "3.1.4-r5"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			name, ver := splitApkPackage(tt.input)
			if name != tt.wantName || ver != tt.wantVer {
				t.Errorf("splitApkPackage(%q) = (%q, %q), want (%q, %q)",
					tt.input, name, ver, tt.wantName, tt.wantVer)
			}
		})
	}
}

func TestParseDpkgOutputEmpty(t *testing.T) {
	pkgs := parseDpkgOutput("")
	if len(pkgs) != 0 {
		t.Errorf("got %d packages for empty input, want 0", len(pkgs))
	}
}
