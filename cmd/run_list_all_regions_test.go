package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	runv1 "google.golang.org/api/run/v1"
)

func runTestService(name string) *runv1.Service {
	return &runv1.Service{Metadata: &runv1.ObjectMeta{Name: name}}
}

func TestRunListAllV1(t *testing.T) {
	pages := map[string]runV1Page[*runv1.Service]{
		"":   {items: []*runv1.Service{runTestService("web"), runTestService("api")}, next: "p2", unreachable: []string{"us-east1"}},
		"p2": {items: []*runv1.Service{runTestService("batch")}, unreachable: []string{"asia-east1", "us-east1"}},
	}
	calls := 0
	page := func(_ context.Context, token string) (runV1Page[*runv1.Service], error) {
		calls++
		return pages[token], nil
	}
	var warn bytes.Buffer
	got, err := runListAllV1(context.Background(), &warn, page, func(s *runv1.Service) string { return runV1ObjectName(s.Metadata) })
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, s := range got {
		names = append(names, s.Metadata.Name)
	}
	if strings.Join(names, ",") != "api,batch,web" || calls != 2 {
		t.Errorf("got %v in %d calls, want sorted api,batch,web in 2", names, calls)
	}
	want := "WARNING: The following Cloud Run regions did not respond: asia-east1, us-east1. List results may be incomplete.\n"
	if warn.String() != want {
		t.Errorf("warning = %q, want %q", warn.String(), want)
	}
}

func TestRunListAllV1Empty(t *testing.T) {
	page := func(context.Context, string) (runV1Page[*runv1.Job], error) { return runV1Page[*runv1.Job]{}, nil }
	var warn bytes.Buffer
	got, err := runListAllV1(context.Background(), &warn, page, func(j *runv1.Job) string { return runV1ObjectName(j.Metadata) })
	if err != nil || got == nil || len(got) != 0 || warn.Len() != 0 {
		t.Errorf("got %v, %v, warning %q; want empty non-nil list and no warning", got, err, warn.String())
	}
}

func TestRunListAllV1Error(t *testing.T) {
	page := func(context.Context, string) (runV1Page[*runv1.Job], error) {
		return runV1Page[*runv1.Job]{}, errors.New("boom")
	}
	if _, err := runListAllV1(context.Background(), &bytes.Buffer{}, page, func(j *runv1.Job) string { return "" }); err == nil {
		t.Error("want error")
	}
}

func TestRunListRegionOptional(t *testing.T) {
	for name, cmd := range map[string]*cobra.Command{"services": runServicesListCmd, "jobs": runJobsListCmd} {
		f := cmd.Flags().Lookup("region")
		if f == nil {
			t.Fatalf("run %s list missing --region", name)
		}
		if _, required := f.Annotations["cobra_annotation_bash_completion_one_required_flag"]; required {
			t.Errorf("run %s list --region should be optional", name)
		}
	}
	if f := runServicesDescribeCmd.Flags().Lookup("region"); f == nil || f.Annotations["cobra_annotation_bash_completion_one_required_flag"] == nil {
		t.Error("run services describe --region should stay required")
	}
}
