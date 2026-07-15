package cmd

import (
	"testing"
)

func TestPubsubHasLiteOperationsSubgroup(t *testing.T) {
	if pubsubSubgroup("lite-operations") == nil {
		t.Fatal("pubsub missing lite-operations subgroup")
	}
}

func TestPubsubLiteOperationsSubcommands(t *testing.T) {
	g := pubsubSubgroup("lite-operations")
	if g == nil {
		t.Fatal("lite-operations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestPubsubHasLiteReservationsSubgroup(t *testing.T) {
	if pubsubSubgroup("lite-reservations") == nil {
		t.Fatal("pubsub missing lite-reservations subgroup")
	}
}

func TestPubsubLiteReservationsSubcommands(t *testing.T) {
	g := pubsubSubgroup("lite-reservations")
	if g == nil {
		t.Fatal("lite-reservations missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list", "list-topics", "update"})
}

func TestPubsubHasMessageTransformsSubgroup(t *testing.T) {
	if pubsubSubgroup("message-transforms") == nil {
		t.Fatal("pubsub missing message-transforms subgroup")
	}
}

func TestPubsubMessageTransformsSubcommands(t *testing.T) {
	g := pubsubSubgroup("message-transforms")
	if g == nil {
		t.Fatal("message-transforms missing")
	}
	assertSubcommands(t, g, []string{"test", "validate"})
}

func TestPubsubLiteRegion(t *testing.T) {
	cases := map[string]string{
		"us-central1":   "us-central1",
		"us-central1-a": "us-central1",
		"europe-west4":  "europe-west4",
	}
	for in, want := range cases {
		got, err := pubsubLiteRegion(in)
		if err != nil {
			t.Errorf("pubsubLiteRegion(%q) unexpected error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("pubsubLiteRegion(%q) = %q, want %q", in, got, want)
		}
	}
	for _, bad := range []string{"", "singleword"} {
		if _, err := pubsubLiteRegion(bad); err == nil {
			t.Errorf("pubsubLiteRegion(%q) expected an error", bad)
		}
	}
}

func TestPslResName(t *testing.T) {
	full := "projects/x/locations/us-central1/reservations/y"
	if got := pslResName(full, "ignored", "ignored"); got != full {
		t.Errorf("pslResName should pass fully qualified names, got %q", got)
	}
	if got := pslResName("my-res", "p", "us-central1"); got != "projects/p/locations/us-central1/reservations/my-res" {
		t.Errorf("pslResName = %q", got)
	}
}
