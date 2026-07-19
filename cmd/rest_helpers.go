package cmd

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

	"github.com/flyingobsidian/gcloud-go/internal/auth"
	"golang.org/x/oauth2"
)

// restClient is a thin OAuth2 HTTP client keyed to a specific API endpoint
// (e.g. https://designcenter.googleapis.com/v1alpha) for services that
// aren't exposed by google.golang.org/api.
type restClient struct {
	endpoint string
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

func (c *restClient) do(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	ts, err := auth.TokenSource(ctx, flagAccount, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return fmt.Errorf("obtaining credentials: %w", err)
	}
	client := oauth2.NewClient(ctx, ts)
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

// paginate lists a REST collection into a flat slice of resources.
func (c *restClient) paginate(ctx context.Context, path string, base url.Values, sliceField string, pageSize int64) ([]map[string]any, error) {
	var all []map[string]any
	pageToken := ""
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
		if arr, ok := raw[sliceField].([]any); ok {
			for _, e := range arr {
				if m, ok := e.(map[string]any); ok {
					all = append(all, m)
				}
			}
		}
		tok, _ := raw["nextPageToken"].(string)
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
