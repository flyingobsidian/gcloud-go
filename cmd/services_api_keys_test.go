package cmd

import (
	"testing"

	apikeys "google.golang.org/api/apikeys/v2"
)

func TestAPIKeyCreateRequiresAPITarget(t *testing.T) {
	if apiKeyCreateCmd.Flags().Changed("api-target") {
		t.Skip("api-target already set by another test")
	}
	if err := runAPIKeyCreate(apiKeyCreateCmd, nil); err != errAPIKeyCreateNeedsTarget {
		t.Fatalf("runAPIKeyCreate without --api-target = %v, want %v", err, errAPIKeyCreateNeedsTarget)
	}
}

func TestCheckAPIKeyUpdateTargets(t *testing.T) {
	withTargets := &apikeys.V2Key{Restrictions: &apikeys.V2Restrictions{
		ApiTargets: []*apikeys.V2ApiTarget{{Service: "foo.googleapis.com"}},
	}}
	noTargets := &apikeys.V2Key{Restrictions: &apikeys.V2Restrictions{
		ServerKeyRestrictions: &apikeys.V2ServerKeyRestrictions{AllowedIps: []string{"1.2.3.4"}},
	}}
	cases := []struct {
		name      string
		current   *apikeys.V2Key
		specified bool
		wantErr   bool
	}{
		{"existing targets", withTargets, false, false},
		{"specifying targets", noTargets, true, false},
		{"no restrictions, specifying", &apikeys.V2Key{}, true, false},
		{"other restrictions only", noTargets, false, true},
		{"no restrictions", &apikeys.V2Key{}, false, true},
	}
	for _, c := range cases {
		err := checkAPIKeyUpdateTargets(c.current, c.specified)
		if (err != nil) != c.wantErr {
			t.Errorf("%s: err = %v, wantErr %v", c.name, err, c.wantErr)
		}
	}
}

func TestAPIKeyUpdateHasNoClearRestrictions(t *testing.T) {
	if apiKeyUpdateCmd.Flags().Lookup("clear-restrictions") != nil {
		t.Error("--clear-restrictions was removed in gcloud-python 586.0.0")
	}
}
