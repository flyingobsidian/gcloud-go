package cmd

import "testing"

func TestBuildCreateServiceAccountRequest(t *testing.T) {
	t.Run("display name and description", func(t *testing.T) {
		flagSADisplayName = "My SA"
		flagSADescription = "does things"
		t.Cleanup(func() { flagSADisplayName = ""; flagSADescription = "" })

		req := buildCreateServiceAccountRequest("my-sa")
		if req.AccountId != "my-sa" {
			t.Errorf("AccountId = %q, want my-sa", req.AccountId)
		}
		if req.ServiceAccount == nil {
			t.Fatal("ServiceAccount is nil, want populated")
		}
		if req.ServiceAccount.DisplayName != "My SA" {
			t.Errorf("DisplayName = %q, want My SA", req.ServiceAccount.DisplayName)
		}
		if req.ServiceAccount.Description != "does things" {
			t.Errorf("Description = %q, want does things", req.ServiceAccount.Description)
		}
	})

	t.Run("display name only", func(t *testing.T) {
		flagSADisplayName = "Only Name"
		flagSADescription = ""
		t.Cleanup(func() { flagSADisplayName = "" })

		req := buildCreateServiceAccountRequest("sa2")
		if req.ServiceAccount == nil {
			t.Fatal("ServiceAccount is nil, want populated")
		}
		if req.ServiceAccount.DisplayName != "Only Name" {
			t.Errorf("DisplayName = %q, want Only Name", req.ServiceAccount.DisplayName)
		}
		if req.ServiceAccount.Description != "" {
			t.Errorf("Description = %q, want empty", req.ServiceAccount.Description)
		}
	})

	t.Run("no optional fields omits ServiceAccount body", func(t *testing.T) {
		flagSADisplayName = ""
		flagSADescription = ""

		req := buildCreateServiceAccountRequest("bare-sa")
		if req.AccountId != "bare-sa" {
			t.Errorf("AccountId = %q, want bare-sa", req.AccountId)
		}
		if req.ServiceAccount != nil {
			t.Errorf("ServiceAccount = %+v, want nil", req.ServiceAccount)
		}
	})
}
