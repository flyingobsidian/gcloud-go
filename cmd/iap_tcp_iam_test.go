package cmd

import "testing"

func TestIapTcpIamSubcommands(t *testing.T) {
	for _, want := range []string{"get-iam-policy", "set-iam-policy", "add-iam-policy-binding", "remove-iam-policy-binding", "dest-groups"} {
		if findSub(iapTcpCmd, want) == nil {
			t.Errorf("iap tcp %s missing", want)
		}
	}
}

func TestIapTcpIamResourceName(t *testing.T) {
	for _, tc := range []struct {
		service, instance, region, want string
	}{
		{"", "", "", "projects/123/iap_tunnel/cloudRun"},
		{"", "", "us-central1", "projects/123/iap_tunnel/cloudRun/locations/us-central1"},
		{"svc", "", "us-central1", "projects/123/iap_tunnel/cloudRun/locations/us-central1/services/svc"},
		{"", "inst", "us-central1", "projects/123/iap_tunnel/cloudRun/locations/us-central1/instances/inst"},
	} {
		if got := iapTcpIamResourceName("123", tc.service, tc.instance, tc.region); got != tc.want {
			t.Errorf("iapTcpIamResourceName(%q, %q, %q) = %q, want %q", tc.service, tc.instance, tc.region, got, tc.want)
		}
	}
}

func TestIapTcpIamValidate(t *testing.T) {
	if err := iapTcpIamValidate("cloud-run", "", "", ""); err != nil {
		t.Errorf("project-wide: %v", err)
	}
	if err := iapTcpIamValidate("cloud-run", "svc", "", "r"); err != nil {
		t.Errorf("service: %v", err)
	}
	for _, tc := range [][4]string{
		{"", "", "", ""},
		{"cloud-run", "svc", "", ""},
		{"cloud-run", "", "inst", ""},
	} {
		if err := iapTcpIamValidate(tc[0], tc[1], tc[2], tc[3]); err == nil {
			t.Errorf("iapTcpIamValidate(%q): expected error", tc)
		}
	}
}
