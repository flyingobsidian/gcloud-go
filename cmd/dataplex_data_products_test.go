package cmd

import (
	"reflect"
	"testing"

	dataplex "google.golang.org/api/dataplex/v1"
)

func TestDataplexDataProductsRegistered(t *testing.T) {
	c, _, err := dataplexCmd.Find([]string{"data-products"})
	if err != nil || c != dataplexDataProductsCmd {
		t.Fatalf("dataplex data-products not registered: %v", err)
	}
	assertSubcommands(t, c, []string{
		"add-iam-policy-binding", "create", "data-assets", "delete", "describe",
		"get-iam-policy", "list", "remove-iam-policy-binding", "set-iam-policy", "update",
	})
	assertSubcommands(t, dataplexDataAssetsCmd, []string{"create", "delete", "describe", "list", "update"})
	for _, f := range []string{"display-name", "owner-emails", "access-groups", "labels", "async", "location"} {
		if dpCreateCmd.Flags().Lookup(f) == nil {
			t.Errorf("data-products create missing --%s", f)
		}
	}
	for _, f := range []string{"data-product", "resource", "access-groups"} {
		if daCreateCmd.Flags().Lookup(f) == nil {
			t.Errorf("data-assets create missing --%s", f)
		}
	}
}

func TestSnakeToCamel(t *testing.T) {
	for in, want := range map[string]string{"display_name": "displayName", "google_group": "googleGroup", "id": "id", "iamRoles": "iamRoles"} {
		if got := snakeToCamel(in); got != want {
			t.Errorf("%s: got %q, want %q", in, got, want)
		}
	}
}

func TestParseDPAccessGroups(t *testing.T) {
	got, err := parseDPAccessGroups([]string{
		`{"id": "g1", "display_name": "G1", "description": "d", "principal": {"google_group": "g1@example.com"}}`,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]dataplex.GoogleCloudDataplexV1DataProductAccessGroup{
		"g1": {Id: "g1", DisplayName: "G1", Description: "d",
			Principal: &dataplex.GoogleCloudDataplexV1DataProductPrincipal{GoogleGroup: "g1@example.com"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	for _, bad := range []string{`{"display_name": "x"}`, `{"id": "g", "bogus": 1}`, `not json`} {
		if _, err := parseDPAccessGroups([]string{bad}); err == nil {
			t.Errorf("%s: expected error", bad)
		}
	}
}

func TestParseDAAccessGroupConfigs(t *testing.T) {
	got, err := parseDAAccessGroupConfigs([]string{`{"id": "g1", "iam_roles": ["roles/a", "roles/b"]}`})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]dataplex.GoogleCloudDataplexV1DataAssetAccessGroupConfig{
		"g1": {IamRoles: []string{"roles/a", "roles/b"}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
	if _, err := parseDAAccessGroupConfigs([]string{`{"iam_roles": []}`}); err == nil {
		t.Error("expected error without id")
	}
}

func TestDPXBindings(t *testing.T) {
	pol := &dataplex.GoogleIamV1Policy{}
	dpxAddBinding(pol, "roles/viewer", "user:a@x.com")
	dpxAddBinding(pol, "roles/viewer", "user:b@x.com")
	dpxAddBinding(pol, "roles/viewer", "user:a@x.com")
	if len(pol.Bindings) != 1 || !reflect.DeepEqual(pol.Bindings[0].Members, []string{"user:a@x.com", "user:b@x.com"}) {
		t.Fatalf("after add: %+v", pol.Bindings)
	}
	if err := dpxRemoveBinding(pol, "roles/viewer", "user:a@x.com"); err != nil {
		t.Fatal(err)
	}
	if err := dpxRemoveBinding(pol, "roles/viewer", "user:b@x.com"); err != nil {
		t.Fatal(err)
	}
	if len(pol.Bindings) != 0 {
		t.Errorf("empty binding not dropped: %+v", pol.Bindings)
	}
	if err := dpxRemoveBinding(pol, "roles/viewer", "user:c@x.com"); err == nil {
		t.Error("expected error removing missing binding")
	}
}

func TestDPName(t *testing.T) {
	if got := dpName("projects/p/locations/l", "dataProducts", "dp"); got != "projects/p/locations/l/dataProducts/dp" {
		t.Errorf("short: %q", got)
	}
	full := "projects/q/locations/m/dataProducts/dp/dataAssets/a"
	if got := dpName("projects/p/locations/l/dataProducts/dp", "dataAssets", full); got != full {
		t.Errorf("full: %q", got)
	}
}
