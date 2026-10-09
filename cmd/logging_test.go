package cmd

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/spf13/cobra"
	logging "google.golang.org/api/logging/v2"
)

func loggingSubgroup(name string) *cobra.Command {
	for _, c := range loggingCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

type loggingGroupCase struct {
	name string
	subs []string
}

var loggingCRUDSubcommands = []string{"create", "delete", "describe", "list", "update"}

func TestLoggingSubgroups(t *testing.T) {
	cases := []loggingGroupCase{
		{"buckets", append([]string{"undelete"}, loggingCRUDSubcommands...)},
		{"links", []string{"create", "delete", "describe", "list"}},
		{"locations", []string{"describe", "list"}},
		{"logs", []string{"delete", "list"}},
		{"metrics", loggingCRUDSubcommands},
		{"operations", []string{"cancel", "describe", "list"}},
		{"recent-queries", []string{"list"}},
		{"resource-descriptors", []string{"describe", "list"}},
		{"saved-queries", loggingCRUDSubcommands},
		{"scopes", loggingCRUDSubcommands},
		{"settings", []string{"describe", "update"}},
		{"sinks", loggingCRUDSubcommands},
		{"views", loggingCRUDSubcommands},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			g := loggingSubgroup(tc.name)
			if g == nil {
				t.Fatalf("logging %s missing", tc.name)
			}
			assertSubcommands(t, g, tc.subs)
		})
	}
}

func TestLoggingDataPlaneCommands(t *testing.T) {
	for _, name := range []string{"copy", "read", "write", "tail"} {
		if findSub(loggingCmd, name) == nil {
			t.Errorf("logging %s missing", name)
		}
	}
}

func TestLoggingReadFilter(t *testing.T) {
	now := time.Date(2026, 10, 9, 12, 0, 0, 0, time.UTC)
	week := 7 * 24 * time.Hour
	cases := []struct {
		name, filter, order string
		want                string
	}{
		{"desc no filter", "", "desc", `timestamp>="2026-10-02T12:00:00.000000Z"`},
		{"desc with filter", `resource.type="cloud_scheduler_job"`, "desc",
			`timestamp>="2026-10-02T12:00:00.000000Z" AND resource.type="cloud_scheduler_job"`},
		{"desc with timestamp filter", `timestamp>="2026-01-01T00:00:00Z"`, "desc", `timestamp>="2026-01-01T00:00:00Z"`},
		{"asc", `severity>=ERROR`, "asc", `severity>=ERROR`},
		{"asc no filter", "", "asc", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := loggingReadFilter(tc.filter, tc.order, week, now); got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLoggingReadFilterNonUTCNow(t *testing.T) {
	now := time.Date(2026, 10, 9, 13, 0, 0, 0, time.FixedZone("BST", 3600))
	want := `timestamp>="2026-10-09T11:00:00.000000Z"`
	if got := loggingReadFilter("", "desc", time.Hour, now); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// pagedLister serves n entries in pages of the requested size and records requests.
func pagedLister(n int, reqs *[]logging.ListLogEntriesRequest) logEntriesLister {
	return func(ctx context.Context, req *logging.ListLogEntriesRequest) (*logging.ListLogEntriesResponse, error) {
		*reqs = append(*reqs, *req)
		start := 0
		if req.PageToken != "" {
			fmt.Sscanf(req.PageToken, "%d", &start)
		}
		end := min(start+int(req.PageSize), n)
		resp := &logging.ListLogEntriesResponse{}
		for i := start; i < end; i++ {
			resp.Entries = append(resp.Entries, &logging.LogEntry{InsertId: fmt.Sprint(i)})
		}
		if end < n {
			resp.NextPageToken = fmt.Sprint(end)
		}
		return resp, nil
	}
}

func TestLoggingListEntries(t *testing.T) {
	cases := []struct {
		name                string
		available           int
		pageSize, limit     int64
		wantLen, wantCalls  int
		wantRequestPageSize int64
	}{
		{"no limit single page", 3, 0, 0, 3, 1, 1000},
		{"no limit many pages", 5, 2, 0, 5, 3, 2},
		{"limit smaller than page", 50, 0, 10, 10, 1, 10},
		{"limit across pages", 50, 4, 10, 10, 3, 4},
		{"limit above available", 3, 0, 10, 3, 1, 10},
		{"limit above backend cap", 5, 0, 5000, 5, 1, 1000},
		{"empty", 0, 0, 0, 0, 1, 1000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var reqs []logging.ListLogEntriesRequest
			req := &logging.ListLogEntriesRequest{PageSize: tc.pageSize}
			got, err := loggingListEntries(context.Background(), pagedLister(tc.available, &reqs), req, tc.limit)
			if err != nil {
				t.Fatalf("loggingListEntries: %v", err)
			}
			if got == nil {
				t.Fatal("got nil slice, want non-nil")
			}
			if len(got) != tc.wantLen {
				t.Errorf("len = %d, want %d", len(got), tc.wantLen)
			}
			if len(reqs) != tc.wantCalls {
				t.Errorf("calls = %d, want %d", len(reqs), tc.wantCalls)
			}
			if reqs[0].PageSize != tc.wantRequestPageSize {
				t.Errorf("page size = %d, want %d", reqs[0].PageSize, tc.wantRequestPageSize)
			}
			for i, e := range got {
				if e.InsertId != fmt.Sprint(i) {
					t.Errorf("entry %d InsertId = %q", i, e.InsertId)
				}
			}
		})
	}
}

func TestLoggingListEntriesError(t *testing.T) {
	wantErr := errors.New("boom")
	list := func(ctx context.Context, req *logging.ListLogEntriesRequest) (*logging.ListLogEntriesResponse, error) {
		return nil, wantErr
	}
	if _, err := loggingListEntries(context.Background(), list, &logging.ListLogEntriesRequest{}, 0); !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want %v", err, wantErr)
	}
}
