package cmd

import (
	"testing"

	crm "google.golang.org/api/cloudresourcemanager/v3"
)

func TestFolderResourceName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"1234", "folders/1234"},
		{"folders/1234", "folders/1234"},
	}
	for _, c := range cases {
		if got := folderResourceName(c.in); got != c.want {
			t.Errorf("folderResourceName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestResolveParent(t *testing.T) {
	cases := []struct {
		folder, org, want string
		wantErr           bool
	}{
		{"1234", "", "folders/1234", false},
		{"folders/1234", "", "folders/1234", false},
		{"", "5678", "organizations/5678", false},
		{"", "organizations/5678", "organizations/5678", false},
		{"1234", "5678", "", true},
		{"", "", "", true},
	}
	for _, c := range cases {
		got, err := resolveParent(c.folder, c.org)
		if c.wantErr {
			if err == nil {
				t.Errorf("resolveParent(%q,%q) expected error", c.folder, c.org)
			}
			continue
		}
		if err != nil {
			t.Errorf("resolveParent(%q,%q) unexpected error: %v", c.folder, c.org, err)
			continue
		}
		if got != c.want {
			t.Errorf("resolveParent(%q,%q) = %q, want %q", c.folder, c.org, got, c.want)
		}
	}
}

func TestTagKeyResourceName(t *testing.T) {
	if got := tagKeyResourceName("123"); got != "tagKeys/123" {
		t.Errorf("tagKeyResourceName(123) = %q", got)
	}
	if got := tagKeyResourceName("tagKeys/123"); got != "tagKeys/123" {
		t.Errorf("tagKeyResourceName(tagKeys/123) = %q", got)
	}
}

func TestTagValueResourceName(t *testing.T) {
	if got := tagValueResourceName("456"); got != "tagValues/456" {
		t.Errorf("tagValueResourceName(456) = %q", got)
	}
	if got := tagValueResourceName("tagValues/456"); got != "tagValues/456" {
		t.Errorf("tagValueResourceName(tagValues/456) = %q", got)
	}
}

func TestIsAllDigits(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"123", true},
		{"", false},
		{"abc", false},
		{"12a3", false},
	}
	for _, c := range cases {
		if got := isAllDigits(c.in); got != c.want {
			t.Errorf("isAllDigits(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestRegionalCRMEndpoint(t *testing.T) {
	if got := regionalCRMEndpoint(""); got != "" {
		t.Errorf("regionalCRMEndpoint(\"\") = %q, want empty", got)
	}
	if got := regionalCRMEndpoint("us-central1"); got != "https://us-central1-cloudresourcemanager.googleapis.com/" {
		t.Errorf("regionalCRMEndpoint(us-central1) = %q", got)
	}
}

func TestResolveOrgPolicyResource(t *testing.T) {
	cases := []struct {
		project, folder, org string
		want                 string
		wantErr              bool
	}{
		{"my-project", "", "", "projects/my-project", false},
		{"", "123", "", "folders/123", false},
		{"", "folders/123", "", "folders/123", false},
		{"", "", "456", "organizations/456", false},
		{"", "", "organizations/456", "organizations/456", false},
		{"my-project", "123", "", "", true},
		{"", "", "", "", true},
	}
	for _, c := range cases {
		got, err := resolveOrgPolicyResource(c.project, c.folder, c.org)
		if c.wantErr {
			if err == nil {
				t.Errorf("resolveOrgPolicyResource(%q,%q,%q) expected error", c.project, c.folder, c.org)
			}
			continue
		}
		if err != nil {
			t.Errorf("resolveOrgPolicyResource(%q,%q,%q) unexpected error: %v", c.project, c.folder, c.org, err)
			continue
		}
		if got != c.want {
			t.Errorf("resolveOrgPolicyResource(%q,%q,%q) = %q, want %q", c.project, c.folder, c.org, got, c.want)
		}
	}
}

func TestPolicyResourceName(t *testing.T) {
	cases := []struct {
		resource, constraint, want string
	}{
		{"projects/123", "compute.disableSerialPortAccess", "projects/123/policies/compute.disableSerialPortAccess"},
		{"folders/456", "constraints/compute.disableSerialPortAccess", "folders/456/policies/compute.disableSerialPortAccess"},
		{"organizations/789", "projects/123/policies/compute.disableSerialPortAccess", "projects/123/policies/compute.disableSerialPortAccess"},
	}
	for _, c := range cases {
		if got := policyResourceName(c.resource, c.constraint); got != c.want {
			t.Errorf("policyResourceName(%q,%q) = %q, want %q", c.resource, c.constraint, got, c.want)
		}
	}
}

func TestParentForPolicyName(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"projects/123/policies/compute.disableSerialPortAccess", "projects/123"},
		{"folders/456/policies/compute.disableSerialPortAccess", "folders/456"},
		{"organizations/789/policies/foo.bar", "organizations/789"},
		{"projects/123", "projects/123"},
	}
	for _, c := range cases {
		if got := parentForPolicyName(c.in); got != c.want {
			t.Errorf("parentForPolicyName(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRmBuildCondition(t *testing.T) {
	if rmBuildCondition("", "", "") != nil {
		t.Error("expected nil for empty inputs")
	}
	c := rmBuildCondition("expr", "title", "desc")
	if c == nil {
		t.Fatal("expected non-nil condition")
	}
	if c.Expression != "expr" || c.Title != "title" || c.Description != "desc" {
		t.Errorf("unexpected condition: %+v", c)
	}
	// Type-check: must return the CRM v3 Expr type.
	var _ *crm.Expr = c
}
