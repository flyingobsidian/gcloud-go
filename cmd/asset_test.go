package cmd

import (
	"testing"
)

func TestResolveAssetScope(t *testing.T) {
	cases := []struct {
		project, folder, org string
		want                 string
		wantErr              bool
	}{
		{"my-project", "", "", "projects/my-project", false},
		{"", "123", "", "folders/123", false},
		{"", "folders/123", "", "folders/123", false},
		{"", "", "789", "organizations/789", false},
		{"", "", "organizations/789", "organizations/789", false},
		{"my-project", "123", "", "", true},
		{"my-project", "", "789", "", true},
		{"", "", "", "", true},
	}
	for _, c := range cases {
		got, err := resolveAssetScope(c.project, c.folder, c.org)
		if c.wantErr {
			if err == nil {
				t.Errorf("resolveAssetScope(%q,%q,%q) expected error", c.project, c.folder, c.org)
			}
			continue
		}
		if err != nil {
			t.Errorf("resolveAssetScope(%q,%q,%q) unexpected error: %v", c.project, c.folder, c.org, err)
			continue
		}
		if got != c.want {
			t.Errorf("resolveAssetScope(%q,%q,%q) = %q, want %q", c.project, c.folder, c.org, got, c.want)
		}
	}
}

func TestFeedName(t *testing.T) {
	cases := []struct{ parent, id, want string }{
		{"projects/my-project", "my-feed", "projects/my-project/feeds/my-feed"},
		{"projects/my-project", "feeds/my-feed", "projects/my-project/feeds/my-feed"},
		{"projects/my-project", "projects/other/feeds/foo", "projects/other/feeds/foo"},
	}
	for _, c := range cases {
		if got := feedName(c.parent, c.id); got != c.want {
			t.Errorf("feedName(%q,%q) = %q, want %q", c.parent, c.id, got, c.want)
		}
	}
}

func TestSavedQueryName(t *testing.T) {
	cases := []struct{ parent, id, want string }{
		{"projects/my-project", "my-query", "projects/my-project/savedQueries/my-query"},
		{"projects/my-project", "savedQueries/my-query", "projects/my-project/savedQueries/my-query"},
		{"projects/my-project", "folders/1/savedQueries/foo", "folders/1/savedQueries/foo"},
	}
	for _, c := range cases {
		if got := savedQueryName(c.parent, c.id); got != c.want {
			t.Errorf("savedQueryName(%q,%q) = %q, want %q", c.parent, c.id, got, c.want)
		}
	}
}

func TestNormalizeContentType(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"resource", "RESOURCE"},
		{"iam-policy", "IAM_POLICY"},
		{"org-policy", "ORG_POLICY"},
		{"os-inventory", "OS_INVENTORY"},
	}
	for _, c := range cases {
		if got := normalizeContentType(c.in); got != c.want {
			t.Errorf("normalizeContentType(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestBuildAssetExportOutputConfigGCS(t *testing.T) {
	flagAssetExportGCS = "gs://bucket/object"
	flagAssetExportGCSPrefix = ""
	flagAssetExportBQDataset = ""
	flagAssetExportBQTable = ""
	t.Cleanup(func() { flagAssetExportGCS, flagAssetExportGCSPrefix = "", "" })

	cfg, err := buildAssetExportOutputConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.GcsDestination == nil || cfg.GcsDestination.Uri != "gs://bucket/object" {
		t.Errorf("unexpected GCS destination: %+v", cfg.GcsDestination)
	}
	if cfg.BigqueryDestination != nil {
		t.Errorf("unexpected BQ destination: %+v", cfg.BigqueryDestination)
	}
}

func TestBuildAssetExportOutputConfigBQ(t *testing.T) {
	flagAssetExportGCS = ""
	flagAssetExportGCSPrefix = ""
	flagAssetExportBQDataset = "projects/p/datasets/d"
	flagAssetExportBQTable = "assets_snapshot"
	flagAssetExportBQForce = true
	t.Cleanup(func() {
		flagAssetExportBQDataset, flagAssetExportBQTable, flagAssetExportBQForce = "", "", false
	})

	cfg, err := buildAssetExportOutputConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.BigqueryDestination == nil || cfg.BigqueryDestination.Dataset != "projects/p/datasets/d" ||
		cfg.BigqueryDestination.Table != "assets_snapshot" || !cfg.BigqueryDestination.Force {
		t.Errorf("unexpected BQ destination: %+v", cfg.BigqueryDestination)
	}
}

func TestBuildAssetExportOutputConfigErrors(t *testing.T) {
	flagAssetExportGCS, flagAssetExportGCSPrefix = "", ""
	flagAssetExportBQDataset, flagAssetExportBQTable = "", ""
	if _, err := buildAssetExportOutputConfig(); err == nil {
		t.Error("expected error for no destination")
	}

	flagAssetExportGCS = "gs://bucket/o"
	flagAssetExportBQDataset = "projects/p/datasets/d"
	if _, err := buildAssetExportOutputConfig(); err == nil {
		t.Error("expected error when both GCS and BQ set")
	}
	t.Cleanup(func() { flagAssetExportGCS, flagAssetExportBQDataset = "", "" })

	flagAssetExportGCS = ""
	flagAssetExportBQTable = ""
	if _, err := buildAssetExportOutputConfig(); err == nil {
		t.Error("expected error when only BQ dataset set")
	}
}

func TestBuildIamPolicyAnalysisQuery(t *testing.T) {
	flagAssetAnalyzeFullResourceName = "//cloudresourcemanager.googleapis.com/projects/123"
	flagAssetAnalyzeIdentity = "user:alice@example.com"
	flagAssetAnalyzePermissions = []string{"compute.instances.get"}
	flagAssetAnalyzeExpandGroups = true
	t.Cleanup(func() {
		flagAssetAnalyzeFullResourceName = ""
		flagAssetAnalyzeIdentity = ""
		flagAssetAnalyzePermissions = nil
		flagAssetAnalyzeExpandGroups = false
	})

	q := buildIamPolicyAnalysisQuery("projects/my-project")
	if q.Scope != "projects/my-project" {
		t.Errorf("Scope = %q", q.Scope)
	}
	if q.ResourceSelector == nil || q.ResourceSelector.FullResourceName != flagAssetAnalyzeFullResourceName {
		t.Errorf("ResourceSelector = %+v", q.ResourceSelector)
	}
	if q.IdentitySelector == nil || q.IdentitySelector.Identity != flagAssetAnalyzeIdentity {
		t.Errorf("IdentitySelector = %+v", q.IdentitySelector)
	}
	if q.AccessSelector == nil || len(q.AccessSelector.Permissions) != 1 {
		t.Errorf("AccessSelector = %+v", q.AccessSelector)
	}
	if q.Options == nil || !q.Options.ExpandGroups {
		t.Errorf("Options = %+v", q.Options)
	}
}

func TestBuildIamPolicyAnalysisQueryMinimal(t *testing.T) {
	flagAssetAnalyzeFullResourceName = ""
	flagAssetAnalyzeIdentity = ""
	flagAssetAnalyzePermissions = nil
	flagAssetAnalyzeRoles = nil
	flagAssetAnalyzeExpandGroups = false
	flagAssetAnalyzeExpandRoles = false
	flagAssetAnalyzeExpandResources = false
	flagAssetAnalyzeGroupEdges = false
	flagAssetAnalyzeResourceEdges = false
	flagAssetAnalyzeSAImp = false

	q := buildIamPolicyAnalysisQuery("organizations/1")
	if q.Scope != "organizations/1" || q.Options != nil || q.ResourceSelector != nil ||
		q.IdentitySelector != nil || q.AccessSelector != nil {
		t.Errorf("expected only Scope populated, got %+v", q)
	}
}

func TestParentForPolicyNameAssetFriendly(t *testing.T) {
	// Sanity: reuse of resource-manager helpers is not accidental — asset.go
	// declares its own helpers and does not reuse resource-manager parenting.
	// This test exists to lock the resolveAssetScope contract rather than
	// tie it to unrelated helpers.
	got, err := resolveAssetScope("my-project", "", "")
	if err != nil || got != "projects/my-project" {
		t.Errorf("resolveAssetScope changed contract: %q err=%v", got, err)
	}
}
