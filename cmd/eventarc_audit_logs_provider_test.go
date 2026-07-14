package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func TestEventarcAuditLogsProviderRegistered(t *testing.T) {
	var alp *cobra.Command
	for _, c := range eventarcCmd.Commands() {
		if c.Name() == "audit-logs-provider" {
			alp = c
			break
		}
	}
	if alp == nil {
		t.Fatal("eventarc audit-logs-provider not registered")
	}
	sub := map[string]*cobra.Command{}
	for _, c := range alp.Commands() {
		sub[c.Name()] = c
	}
	for _, name := range []string{"method-names", "service-names"} {
		if sub[name] == nil {
			t.Fatalf("audit-logs-provider %s subgroup missing", name)
		}
	}
	// each subgroup must expose a `list` command
	for _, name := range []string{"method-names", "service-names"} {
		found := false
		for _, c := range sub[name].Commands() {
			if c.Name() == "list" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("audit-logs-provider %s missing list subcommand", name)
		}
	}
}

func TestEventarcAuditLogsMethodNamesRequiresServiceName(t *testing.T) {
	f := eventarcAuditLogsMethodNamesListCmd.Flag("service-name")
	if f == nil {
		t.Fatal("--service-name flag not registered")
	}
	req := f.Annotations[cobra.BashCompOneRequiredFlag]
	if len(req) == 0 || req[0] != "true" {
		t.Errorf("--service-name not marked required (annotations=%v)", f.Annotations)
	}
}

func TestParseAuditLogCatalog(t *testing.T) {
	body := []byte(`{
	  "services": [
	    {"serviceName": "storage.googleapis.com", "displayName": "Cloud Storage",
	     "methods": [{"methodName": "storage.buckets.get"}, {"methodName": "storage.objects.get"}]},
	    {"serviceName": "compute.googleapis.com", "displayName": "Compute Engine",
	     "methods": [{"methodName": "compute.instances.insert"}]}
	  ]
	}`)
	svcs, err := parseAuditLogCatalog(body)
	if err != nil {
		t.Fatal(err)
	}
	if len(svcs) != 2 || svcs[0].ServiceName != "storage.googleapis.com" {
		t.Fatalf("unexpected services: %+v", svcs)
	}
	methods, err := findAuditLogMethods(svcs, "compute.googleapis.com")
	if err != nil {
		t.Fatal(err)
	}
	if len(methods) != 1 || methods[0].MethodName != "compute.instances.insert" {
		t.Errorf("unexpected methods: %+v", methods)
	}
	if _, err := findAuditLogMethods(svcs, "nope.googleapis.com"); err == nil {
		t.Error("expected error for unknown service")
	}
}
