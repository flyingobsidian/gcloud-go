package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	servicemanagement "google.golang.org/api/servicemanagement/v1"
)

func endpointsSubgroup(name string) *cobra.Command {
	for _, c := range endpointsCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestEndpointsConfigsSubcommands(t *testing.T) {
	g := endpointsSubgroup("configs")
	if g == nil {
		t.Fatal("endpoints configs missing")
	}
	assertSubcommands(t, g, []string{"describe", "list"})
}

func TestEndpointsOperationsSubcommands(t *testing.T) {
	g := endpointsSubgroup("operations")
	if g == nil {
		t.Fatal("endpoints operations missing")
	}
	assertSubcommands(t, g, []string{"describe", "list", "wait"})
}

func TestEndpointsServicesSubcommands(t *testing.T) {
	g := endpointsSubgroup("services")
	if g == nil {
		t.Fatal("endpoints services missing")
	}
	assertSubcommands(t, g, []string{
		"delete", "undeploy", "describe", "list", "deploy", "get-config-name",
		"enable", "disable",
		"add-iam-policy-binding", "remove-iam-policy-binding",
		"get-iam-policy", "set-iam-policy", "check-iam-policy",
	})
}

func TestEndpointsMergeBinding(t *testing.T) {
	p := &servicemanagement.Policy{}
	epMergeBinding(p, "user:alice@example.com", "roles/viewer")
	epMergeBinding(p, "user:bob@example.com", "roles/viewer")
	epMergeBinding(p, "user:alice@example.com", "roles/viewer") // no-op
	if len(p.Bindings) != 1 {
		t.Fatalf("want 1 binding, got %d", len(p.Bindings))
	}
	if len(p.Bindings[0].Members) != 2 {
		t.Fatalf("want 2 members, got %d", len(p.Bindings[0].Members))
	}
	epRemoveBinding(p, "user:alice@example.com", "roles/viewer")
	if len(p.Bindings[0].Members) != 1 {
		t.Fatalf("want 1 member after remove, got %d", len(p.Bindings[0].Members))
	}
	if p.Bindings[0].Members[0] != "user:bob@example.com" {
		t.Fatalf("wrong remaining member: %s", p.Bindings[0].Members[0])
	}
}

func TestEndpointsDetectFileType(t *testing.T) {
	cases := []struct {
		name string
		data string
		want string
	}{
		{"api.yaml", "swagger: '2.0'\n", "OPEN_API_YAML"},
		{"service.yaml", "type: google.api.Service\nname: my.service\n", "SERVICE_CONFIG_YAML"},
		{"desc.pb", "\x00\x01\x02", "FILE_DESCRIPTOR_SET_PROTO"},
	}
	for _, tc := range cases {
		got, err := epDetectFileType(tc.name, []byte(tc.data))
		if err != nil {
			t.Errorf("%s: unexpected err %v", tc.name, err)
			continue
		}
		if got != tc.want {
			t.Errorf("%s: got %q want %q", tc.name, got, tc.want)
		}
	}
}
