package cmd

import (
	"testing"
	"time"
)

func TestParseGcloudDuration(t *testing.T) {
	cases := []struct {
		in   string
		want time.Duration
	}{
		{"90", 90 * time.Second},
		{"1.5", 1500 * time.Millisecond},
		{"7d", 7 * 24 * time.Hour},
		{"1d", 24 * time.Hour},
		{"1h30m", 90 * time.Minute},
		{"2.5h", 150 * time.Minute},
		{"45s", 45 * time.Second},
		{"1D2H", 26 * time.Hour},
		{"P1D", 24 * time.Hour},
		{"PT1H30M", 90 * time.Minute},
		{"P1DT2H", 26 * time.Hour},
		{"pt10s", 10 * time.Second},
	}
	for _, tc := range cases {
		got, err := parseGcloudDuration(tc.in)
		if err != nil {
			t.Errorf("parseGcloudDuration(%q): %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("parseGcloudDuration(%q) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

func TestParseGcloudDurationInvalid(t *testing.T) {
	for _, in := range []string{"", "d", "7x", "-1", "-1d", "1d2", "P", "PT", "P1H", "abc"} {
		if got, err := parseGcloudDuration(in); err == nil {
			t.Errorf("parseGcloudDuration(%q) = %v, want error", in, got)
		}
	}
}
