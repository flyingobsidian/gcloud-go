package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestDeveloperKnowledgeCommandTree(t *testing.T) {
	s := newDeveloperKnowledgeCmds(legacyTestInvocation())
	assertSubcommands(t, s.developerKnowledgeCmd, []string{"answer-query", "documents"})
	assertSubcommands(t, s.dkDocumentsCmd, []string{"describe", "search-chunks"})
}

// dkFakeAPI records requests and serves developerknowledge v1 responses. The
// search-chunks results span three pages of two chunks each.
type dkFakeAPI struct {
	requests []string
	bodies   []string
}

func (f *dkFakeAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	f.requests = append(f.requests, r.Method+" "+r.URL.Path+"?"+r.URL.RawQuery)
	f.bodies = append(f.bodies, string(body))
	w.Header().Set("Content-Type", "application/json")
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/v1:answerQuery":
		fmt.Fprint(w, `{"answer":{"answerText":"Use gcloud storage buckets create."}}`)
	case r.Method == http.MethodGet && r.URL.Path == "/v1/documents:searchDocumentChunks":
		page := map[string]int{"": 0, "p1": 1, "p2": 2}[r.URL.Query().Get("pageToken")]
		next := ""
		if page < 2 {
			next = fmt.Sprintf("p%d", page+1)
		}
		fmt.Fprintf(w, `{"results":[{"id":"c%d","parent":"documents/d%d"},{"id":"c%d","parent":"documents/d%d"}],"nextPageToken":%q}`,
			2*page, 2*page, 2*page+1, 2*page+1, next)
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/documents/"):
		fmt.Fprintf(w, `{"name":%q,"title":"Create buckets"}`, strings.TrimPrefix(r.URL.Path, "/v1/"))
	default:
		http.Error(w, r.Method+" "+r.URL.Path, http.StatusNotFound)
	}
}

func TestDeveloperKnowledgeAnswerQuery(t *testing.T) {
	api := &dkFakeAPI{}
	inv, stdout, _ := newFakeAPIInvocation(t, api)
	err := runInvocation(inv, "developer-knowledge", "answer-query", "--query", "How?",
		"--query-filter", `data_source = "docs.cloud.google.com"`)
	if err != nil {
		t.Fatal(err)
	}
	var sent map[string]string
	if err := json.Unmarshal([]byte(api.bodies[0]), &sent); err != nil ||
		sent["query"] != "How?" || sent["filter"] != `data_source = "docs.cloud.google.com"` {
		t.Errorf("body = %s", api.bodies[0])
	}
	if !strings.Contains(stdout.String(), "answerText: Use gcloud storage buckets create.") {
		t.Errorf("stdout = %q, want the answer", stdout)
	}
}

func TestDeveloperKnowledgeDocumentsDescribe(t *testing.T) {
	api := &dkFakeAPI{}
	inv, stdout, _ := newFakeAPIInvocation(t, api)
	err := runInvocation(inv, "developer-knowledge", "documents", "describe",
		"documents/docs.cloud.google.com/storage/docs/creating-buckets", "--view", "basic", "--format", "value(title)")
	if err != nil {
		t.Fatal(err)
	}
	want := "GET /v1/documents/docs.cloud.google.com/storage/docs/creating-buckets?alt=json&prettyPrint=false&view=DOCUMENT_VIEW_BASIC"
	if len(api.requests) != 1 || api.requests[0] != want {
		t.Errorf("requests = %v, want [%s]", api.requests, want)
	}
	if stdout.String() != "Create buckets\n" {
		t.Errorf("stdout = %q", stdout)
	}
}

func TestDeveloperKnowledgeDocumentsDescribeDefaultView(t *testing.T) {
	api := &dkFakeAPI{}
	inv, _, _ := newFakeAPIInvocation(t, api)
	if err := runInvocation(inv, "developer-knowledge", "documents", "describe", "documents/x"); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(api.requests[0], "view=DOCUMENT_VIEW_CONTENT") {
		t.Errorf("request = %s, want the content view", api.requests[0])
	}
}

func TestDeveloperKnowledgeDocumentsDescribeBadView(t *testing.T) {
	api := &dkFakeAPI{}
	inv, _, _ := newFakeAPIInvocation(t, api)
	err := runInvocation(inv, "developer-knowledge", "documents", "describe", "documents/x", "--view", "foo")
	want := "argument --view: Invalid choice: 'foo'.\n\nValid choices are [basic, content, full]."
	if err == nil || err.Error() != want {
		t.Errorf("err = %v, want %q", err, want)
	}
	if len(api.requests) != 0 {
		t.Errorf("requests = %v, want none", api.requests)
	}
}

func TestDeveloperKnowledgeSearchChunks(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		want     string
		requests int
	}{
		// The default --limit of 5 stops paging once five chunks are in.
		{"default limit", nil, "c0\nc1\nc2\nc3\nc4\n", 3},
		{"limit", []string{"--limit", "2"}, "c0\nc1\n", 1},
		{"sort-by reads every page", []string{"--sort-by", "~id", "--limit", "2"}, "c5\nc4\n", 3},
		{"filter", []string{"--filter", "id:c3"}, "c3\n", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &dkFakeAPI{}
			inv, stdout, _ := newFakeAPIInvocation(t, api)
			args := append([]string{"developer-knowledge", "documents", "search-chunks", "--query", "buckets",
				"--query-filter", "uri:storage", "--page-size", "2", "--format", "value(id)"}, tt.args...)
			if err := runInvocation(inv, args...); err != nil {
				t.Fatal(err)
			}
			if stdout.String() != tt.want {
				t.Errorf("stdout = %q, want %q", stdout, tt.want)
			}
			if len(api.requests) != tt.requests {
				t.Errorf("requests = %v, want %d", api.requests, tt.requests)
			}
			if q := api.requests[0]; !strings.Contains(q, "query=buckets") || !strings.Contains(q, "filter=uri%3Astorage") ||
				!strings.Contains(q, "pageSize=2") {
				t.Errorf("first request = %s", q)
			}
		})
	}
}
