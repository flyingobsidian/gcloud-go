package cmd

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	runv1 "google.golang.org/api/run/v1"
)

// --- run services list / run jobs list without --region (#1996) ---
//
// Like gcloud, when --region is not given these list every region through
// the global Cloud Run v1 endpoint and print the v1 (Knative) resources.

// runV1Page is one page of a global v1 list: its items, the continue token
// and any regions that did not respond.
type runV1Page[T any] struct {
	items       []T
	next        string
	unreachable []string
}

// runListAllV1 pages through a global v1 list, warns on w about unreachable
// regions like gcloud, and returns the items sorted by name. It never
// returns a nil slice, so an empty result renders as [] in JSON.
func runListAllV1[T any](ctx context.Context, w io.Writer, page func(ctx context.Context, continueToken string) (runV1Page[T], error),
	name func(T) string) ([]T, error) {
	all := []T{}
	unreachable := map[string]bool{}
	continueToken := ""
	for {
		p, err := page(ctx, continueToken)
		if err != nil {
			return nil, err
		}
		all = append(all, p.items...)
		for _, r := range p.unreachable {
			unreachable[r] = true
		}
		if p.next == "" {
			break
		}
		continueToken = p.next
	}
	if len(unreachable) > 0 {
		regions := make([]string, 0, len(unreachable))
		for r := range unreachable {
			regions = append(regions, r)
		}
		sort.Strings(regions)
		fmt.Fprintf(w, "WARNING: The following Cloud Run regions did not respond: %s. List results may be incomplete.\n",
			strings.Join(regions, ", "))
	}
	sort.SliceStable(all, func(i, j int) bool { return name(all[i]) < name(all[j]) })
	return all, nil
}

// runV1ObjectName returns metadata.name, or "" when metadata is missing.
func runV1ObjectName(m *runv1.ObjectMeta) string {
	if m == nil {
		return ""
	}
	return m.Name
}

// runListAllServices lists services in every region of project.
func runListAllServices(ctx context.Context, w io.Writer, svc *runv1.APIService, project string) ([]*runv1.Service, error) {
	parent := fmt.Sprintf("projects/%s/locations/-", project)
	page := func(ctx context.Context, continueToken string) (runV1Page[*runv1.Service], error) {
		call := svc.Projects.Locations.Services.List(parent).Context(ctx)
		if continueToken != "" {
			call = call.Continue(continueToken)
		}
		resp, err := call.Do()
		if err != nil {
			return runV1Page[*runv1.Service]{}, fmt.Errorf("listing services: %w", err)
		}
		p := runV1Page[*runv1.Service]{items: resp.Items, unreachable: resp.Unreachable}
		if resp.Metadata != nil {
			p.next = resp.Metadata.Continue
		}
		return p, nil
	}
	return runListAllV1(ctx, w, page, func(s *runv1.Service) string { return runV1ObjectName(s.Metadata) })
}

// runListAllJobs lists jobs in every region of project.
func runListAllJobs(ctx context.Context, w io.Writer, svc *runv1.APIService, project string) ([]*runv1.Job, error) {
	parent := "namespaces/" + project
	page := func(ctx context.Context, continueToken string) (runV1Page[*runv1.Job], error) {
		call := svc.Namespaces.Jobs.List(parent).Context(ctx)
		if continueToken != "" {
			call = call.Continue(continueToken)
		}
		resp, err := call.Do()
		if err != nil {
			return runV1Page[*runv1.Job]{}, fmt.Errorf("listing jobs: %w", err)
		}
		p := runV1Page[*runv1.Job]{items: resp.Items, unreachable: resp.Unreachable}
		if resp.Metadata != nil {
			p.next = resp.Metadata.Continue
		}
		return p, nil
	}
	return runListAllV1(ctx, w, page, func(j *runv1.Job) string { return runV1ObjectName(j.Metadata) })
}
