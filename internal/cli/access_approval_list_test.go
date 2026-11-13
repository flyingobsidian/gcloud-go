package cli

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"google.golang.org/api/googleapi"
)

// A failing list call is reported as an error, not a nil-pointer panic
// (#2089), for each of the three scopes.
func TestAccessApprovalRequestsListAPIError(t *testing.T) {
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprint(w, `{"error":{"code":403,"message":"Access Approval API has not been used","status":"PERMISSION_DENIED"}}`)
	})
	for _, scope := range [][]string{{"--project", "p"}, {"--folder", "1"}, {"--organization", "2"}} {
		t.Run(scope[0], func(t *testing.T) {
			inv, stdout, _ := newFakeAPIInvocation(t, api)
			err := runInvocation(inv, append([]string{"access-approval", "requests", "list"}, scope...)...)
			var gerr *googleapi.Error
			if !errors.As(err, &gerr) || gerr.Code != http.StatusForbidden {
				t.Fatalf("err = %v, want the 403 from the API", err)
			}
			if !strings.HasPrefix(err.Error(), "listing requests: ") {
				t.Errorf("err = %q, want it to start with %q", err, "listing requests: ")
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want nothing", stdout)
			}
		})
	}
}

// The list follows nextPageToken and prints every page's requests.
func TestAccessApprovalRequestsListPages(t *testing.T) {
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("pageToken") == "" {
			fmt.Fprint(w, `{"approvalRequests":[{"name":"projects/p/approvalRequests/a","requestTime":"t1"}],"nextPageToken":"n"}`)
			return
		}
		fmt.Fprint(w, `{"approvalRequests":[{"name":"projects/p/approvalRequests/b","requestTime":"t2"}]}`)
	})
	inv, stdout, _ := newFakeAPIInvocation(t, api)
	if err := runInvocation(inv, "access-approval", "requests", "list", "--project", "p", "--format", "value(name)"); err != nil {
		t.Fatal(err)
	}
	if want := "projects/p/approvalRequests/a\nprojects/p/approvalRequests/b\n"; stdout.String() != want {
		t.Errorf("stdout = %q, want %q", stdout, want)
	}
}
