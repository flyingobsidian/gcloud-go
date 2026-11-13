package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/httplog"
	"golang.org/x/oauth2"
)

// restClient is a thin OAuth2 HTTP client keyed to a specific API endpoint
// (e.g. https://designcenter.googleapis.com/v1alpha) for services that
// aren't exposed by google.golang.org/api.
type restClient struct {
	endpoint string
	// inv supplies credentials, the billing project and HTTP logging. A
	// client from newRESTClient has none and only names an endpoint: bind
	// it to an invocation with inv.rest before making calls.
	inv *invocation
}

// restError is returned by restClient.do when the HTTP response is non-2xx.
// Callers use errors.As to inspect StatusCode and Body for structured checks
// (e.g. treating 404 as "resource does not exist").
type restError struct {
	Method     string
	URL        string
	StatusCode int
	Body       string
}

func (e *restError) Error() string {
	return fmt.Sprintf("%s %s: HTTP %d: %s", e.Method, e.URL, e.StatusCode, e.Body)
}

func newRESTClient(endpoint string) *restClient {
	return &restClient{endpoint: strings.TrimRight(endpoint, "/")}
}

// tokenSource returns the token source for c's calls.
func (c *restClient) tokenSource(ctx context.Context, scopes ...string) (oauth2.TokenSource, error) {
	if c.inv.restTokenSource != nil {
		return c.inv.restTokenSource(ctx, scopes...)
	}
	return c.inv.tokens.TokenSource(ctx, c.inv.flags.account, scopes...)
}

// userProject returns the billing / quota project for c's calls.
func (c *restClient) userProject() string {
	return c.inv.quotaProject()
}

// debugWriter returns where c logs its calls, or nil.
func (c *restClient) debugWriter() io.Writer {
	return c.inv.httpDebugWriter()
}

func (c *restClient) do(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	// Resolve the billing / quota project up front (--billing-project,
	// billing/quota_project config, or ADC quota_project_id) and refuse to
	// make the call without one. Google's APIs charge quota against this
	// project and reject ADC user credentials when it is missing (#1740);
	// failing here names the setting to change, where the server-side
	// PERMISSION_DENIED does not.
	qp := c.userProject()
	if qp == "" {
		return errNoBillingProject(endpointHost(c.endpoint))
	}
	ts, err := c.tokenSource(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return fmt.Errorf("obtaining credentials: %w", err)
	}
	client := oauth2.NewClient(ctx, ts)
	// Log below the oauth2 transport rather than above it, so the printed
	// request shows the Authorization header as actually sent.
	if out := c.debugWriter(); out != nil {
		if t, ok := client.Transport.(*oauth2.Transport); ok {
			t.Base = httplog.NewTransport(t.Base, out)
		}
	}
	u := c.endpoint + path
	if len(query) > 0 {
		u = u + "?" + query.Encode()
	}
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, reqBody)
	if err != nil {
		return err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	// Passing an x-goog-user-project header mirrors what the official client
	// libraries do when a quota project is configured. The value is known to
	// be non-empty: do() refuses the call above when it is not.
	req.Header.Set("X-Goog-User-Project", qp)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP %s %s: %w", method, u, err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &restError{
			Method:     method,
			URL:        u,
			StatusCode: resp.StatusCode,
			Body:       strings.TrimSpace(string(respBody)),
		}
	}
	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}
	return nil
}

// logPageProgress reports how far a paginated listing has got, once per page,
// to out when debug logging is on (out is non-nil).
//
// Google's JSON APIs put the result count in the response body, not in a
// response header: there is no X-Total-Count or equivalent to read. Listings
// that report a grand total do it with the AIP-158 `totalSize` field, which is
// itself optional -- SCC findings.list returns it, many other collections do
// not -- so the total is only printed when the server actually sent one.
func logPageProgress(out io.Writer, sliceField string, page, pageCount, running int, totalSize any, more bool) {
	if out == nil {
		return
	}
	total := ""
	// encoding/json decodes every JSON number into a float64.
	if n, ok := totalSize.(float64); ok {
		total = fmt.Sprintf(" of %d", int64(n))
	}
	remaining := "last page"
	if more {
		remaining = "more pages follow"
	}
	fmt.Fprintln(out, httplog.Note("page %d: %d %s, %d%s so far (%s)",
		page, pageCount, sliceField, running, total, remaining))
}

// paginate lists a REST collection into a flat slice of resources.
func (c *restClient) paginate(ctx context.Context, path string, base url.Values, sliceField string, pageSize int64) ([]map[string]any, error) {
	var all []map[string]any
	pageToken := ""
	page := 0
	for {
		q := url.Values{}
		for k, v := range base {
			q[k] = v
		}
		if pageSize > 0 {
			q.Set("pageSize", fmt.Sprintf("%d", pageSize))
		}
		if pageToken != "" {
			q.Set("pageToken", pageToken)
		}
		var raw map[string]any
		if err := c.do(ctx, http.MethodGet, path, q, nil, &raw); err != nil {
			return nil, err
		}
		page++
		pageCount := 0
		if arr, ok := raw[sliceField].([]any); ok {
			for _, e := range arr {
				if m, ok := e.(map[string]any); ok {
					all = append(all, m)
				}
			}
			pageCount = len(arr)
		}
		tok, _ := raw["nextPageToken"].(string)
		logPageProgress(c.debugWriter(), sliceField, page, pageCount, len(all), raw["totalSize"], tok != "")
		if tok == "" {
			break
		}
		pageToken = tok
	}
	return all, nil
}

// waitOperation polls a long-running operation on this endpoint until done or timeout.
func (c *restClient) waitOperation(ctx context.Context, opName string, timeout time.Duration) (map[string]any, error) {
	deadline := timeout
	if deadline <= 0 {
		deadline = 30 * time.Minute
	}
	waitCtx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()
	backoff := 2 * time.Second
	for {
		var op map[string]any
		if err := c.do(waitCtx, http.MethodGet, "/"+opName, nil, nil, &op); err != nil {
			return nil, err
		}
		if done, _ := op["done"].(bool); done {
			return op, nil
		}
		select {
		case <-waitCtx.Done():
			return nil, fmt.Errorf("timed out waiting for %s: %w", opName, waitCtx.Err())
		case <-time.After(backoff):
		}
		if backoff < 30*time.Second {
			backoff = time.Duration(float64(backoff) * 1.5)
		}
	}
}
