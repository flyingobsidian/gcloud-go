package cmd

import (
	"reflect"
	"testing"
)

func TestBuildCreateRoleRequest(t *testing.T) {
	flagRoleTitle = "My Role"
	flagRoleDescription = "does things"
	flagRolePermissions = []string{"some.permission", "some.other.permission"}
	flagRoleStage = "GA"
	t.Cleanup(func() {
		flagRoleTitle = ""
		flagRoleDescription = ""
		flagRolePermissions = nil
		flagRoleStage = ""
	})

	req := buildCreateRoleRequest("myRole")

	if req.RoleId != "myRole" {
		t.Errorf("RoleId = %q, want myRole", req.RoleId)
	}
	if req.Role == nil {
		t.Fatal("Role is nil, want populated")
	}
	if req.Role.Title != "My Role" {
		t.Errorf("Title = %q, want My Role", req.Role.Title)
	}
	if req.Role.Description != "does things" {
		t.Errorf("Description = %q, want does things", req.Role.Description)
	}
	if req.Role.Stage != "GA" {
		t.Errorf("Stage = %q, want GA", req.Role.Stage)
	}
	want := []string{"some.permission", "some.other.permission"}
	if !reflect.DeepEqual(req.Role.IncludedPermissions, want) {
		t.Errorf("IncludedPermissions = %v, want %v", req.Role.IncludedPermissions, want)
	}
}

func TestBuildCreateRoleRequestMinimal(t *testing.T) {
	flagRoleTitle = ""
	flagRoleDescription = ""
	flagRolePermissions = nil
	flagRoleStage = ""

	req := buildCreateRoleRequest("bareRole")

	if req.RoleId != "bareRole" {
		t.Errorf("RoleId = %q, want bareRole", req.RoleId)
	}
	if req.Role == nil {
		t.Fatal("Role is nil, want non-nil")
	}
	if req.Role.Title != "" || req.Role.Stage != "" || len(req.Role.IncludedPermissions) != 0 {
		t.Errorf("expected empty role fields, got %+v", req.Role)
	}
}
