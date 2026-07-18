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
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

// --- gcloud design-center (#329) ---
//
// The Design Center API (`designcenter.googleapis.com`, v1alpha) is not
// exposed by google.golang.org/api. This file uses a small raw-HTTP client
// keyed off the standard cloud-platform OAuth token; all subgroups
// (locations, operations, spaces) live in the sibling design_center_*.go
// files and dispatch through the helpers below.

var designCenterCmd = &cobra.Command{Use: "design-center", Short: "Manage Google Cloud Design Center"}

func init() {
	rootCmd.AddCommand(designCenterCmd)
}

const designCenterEndpoint = "https://designcenter.googleapis.com/v1alpha"

func dcLocationName(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func dcJoin(parts ...string) string {
	out := parts[0]
	for _, p := range parts[1:] {
		out = out + "/" + strings.TrimPrefix(p, "/")
	}
	return out
}

func designCenterHTTPClient(ctx context.Context) (*http.Client, error) {
	ts, err := auth.TokenSource(ctx, flagAccount, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("obtaining credentials: %w", err)
	}
	return oauth2.NewClient(ctx, ts), nil
}

func dcDo(ctx context.Context, method, path string, query url.Values, body any, out any) error {
	client, err := designCenterHTTPClient(ctx)
	if err != nil {
		return err
	}
	u := designCenterEndpoint + path
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
		return fmt.Errorf("designcenter %s %s: HTTP %d: %s", method, u, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("unmarshaling response: %w", err)
		}
	}
	return nil
}

type dcResource map[string]any

func dcPaginate(ctx context.Context, path string, base url.Values, sliceField string, pageSize int64) ([]dcResource, error) {
	var all []dcResource
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
		if err := dcDo(ctx, http.MethodGet, path, q, nil, &raw); err != nil {
			return nil, err
		}
		if arr, ok := raw[sliceField].([]any); ok {
			for _, e := range arr {
				if m, ok := e.(map[string]any); ok {
					all = append(all, dcResource(m))
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

func dcWaitOperation(ctx context.Context, opName string, timeout time.Duration) (map[string]any, error) {
	deadline := timeout
	if deadline <= 0 {
		deadline = 30 * time.Minute
	}
	waitCtx, cancel := context.WithTimeout(ctx, deadline)
	defer cancel()
	backoff := 2 * time.Second
	for {
		var op map[string]any
		if err := dcDo(waitCtx, http.MethodGet, "/"+opName, nil, nil, &op); err != nil {
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
