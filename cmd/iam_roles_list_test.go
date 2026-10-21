package cmd

import (
	"context"
	"fmt"
	"testing"

	iam "google.golang.org/api/iam/v1"
)

func TestRolesListParent(t *testing.T) {
	cases := []struct {
		name, org, project string
		projectSet         bool
		want               string
		wantErr            bool
	}{
		{name: "predefined", project: "from-config", want: ""},
		{name: "explicit project", project: "p", projectSet: true, want: "projects/p"},
		{name: "organization", org: "123", want: "organizations/123"},
		{name: "both", org: "123", project: "p", projectSet: true, wantErr: true},
	}
	for _, tc := range cases {
		got, err := rolesListParent(tc.org, tc.project, tc.projectSet)
		if (err != nil) != tc.wantErr || got != tc.want {
			t.Errorf("%s: got %q, %v; want %q (error %v)", tc.name, got, err, tc.want, tc.wantErr)
		}
	}
}

func TestListRoles(t *testing.T) {
	calls := 0
	page := func(_ context.Context, token string) (*iam.ListRolesResponse, error) {
		calls++
		if token == "" {
			return &iam.ListRolesResponse{Roles: []*iam.Role{{Name: "roles/a"}, {Name: "roles/b"}}, NextPageToken: "2"}, nil
		}
		return &iam.ListRolesResponse{Roles: []*iam.Role{{Name: "roles/c"}}}, nil
	}
	got, err := listRoles(context.Background(), page, 2)
	if err != nil || len(got) != 2 || calls != 1 {
		t.Errorf("limit 2: %d roles in %d calls, %v", len(got), calls, err)
	}
	calls = 0
	got, err = listRoles(context.Background(), page, 0)
	if err != nil || fmt.Sprint(len(got), calls) != "3 2" {
		t.Errorf("all: %d roles in %d calls, %v", len(got), calls, err)
	}
	empty := func(context.Context, string) (*iam.ListRolesResponse, error) { return &iam.ListRolesResponse{}, nil }
	if got, err := listRoles(context.Background(), empty, 0); err != nil || got == nil || len(got) != 0 {
		t.Errorf("empty: %v, %v", got, err)
	}
}

func TestIamRolesListRegistered(t *testing.T) {
	list, _, err := rolesCmd.Find([]string{"list"})
	if err != nil || list != roleListCmd {
		t.Fatalf("iam roles list not registered: %v", err)
	}
}
