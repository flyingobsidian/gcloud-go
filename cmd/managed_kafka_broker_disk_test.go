package cmd

import "testing"

func TestParseBrokerDiskSizeGiB(t *testing.T) {
	cases := []struct {
		in      string
		want    int64
		wantErr bool
	}{
		{in: "100", want: 100},
		{in: "0", want: 0},
		{in: "100Gi", want: 100},
		{in: "100GiB", want: 100},
		{in: "1Ti", want: 1024},
		{in: "1TiB", want: 1024},
		{in: "  200Gi ", want: 200},
		// Fractional GiB values are rejected.
		{in: "1.5Gi", wantErr: true},
		{in: "500Mi", wantErr: true},
		// Unknown units.
		{in: "10Xi", wantErr: true},
		// Empty / non-numeric.
		{in: "", wantErr: true},
		{in: "Gi", wantErr: true},
		{in: "-5", wantErr: true},
	}
	for _, tc := range cases {
		got, err := parseBrokerDiskSizeGiB(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("parseBrokerDiskSizeGiB(%q) = %d, want error", tc.in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseBrokerDiskSizeGiB(%q) error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("parseBrokerDiskSizeGiB(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestMKClusterCommandsHaveBrokerDiskFlag(t *testing.T) {
	cases := []struct {
		path []string
	}{
		{path: []string{"managed-kafka", "clusters", "create"}},
		{path: []string{"managed-kafka", "clusters", "update"}},
	}
	for _, tc := range cases {
		var cur = findRootSub(tc.path[0])
		for _, n := range tc.path[1:] {
			cur = findSub(cur, n)
		}
		if cur == nil {
			t.Fatalf("%v missing", tc.path)
		}
		if cur.Flags().Lookup("broker-disk-size-gib") == nil {
			t.Errorf("%v missing --broker-disk-size-gib flag", tc.path)
		}
	}
}
