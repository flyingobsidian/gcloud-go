// Package httplog renders outgoing HTTP requests as runnable curl commands
// so `--verbosity debug` / `--log-http` can show exactly which REST call a
// command made, including the headers that carry authentication and the
// billing / quota project.
package httplog

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

// tokenPlaceholder replaces the bearer token in the Authorization header.
// Debug output is routinely pasted into bug reports, so the real token never
// appears; the substitution keeps the printed command runnable because a
// shell expands it when the line is re-run.
const tokenPlaceholder = "Bearer $(gcloud auth print-access-token)"

// Transport writes every request passing through it to Out as a curl command
// and then delegates to Base. Wrap the transport that sits closest to the
// network so the log shows the request as it goes on the wire -- headers
// added by outer transports (Authorization, X-Goog-User-Project, User-Agent)
// are already in place by then.
//
// A nil Out disables logging; a nil Base uses http.DefaultTransport.
type Transport struct {
	Base http.RoundTripper
	Out  io.Writer
}

// NewTransport returns a Transport logging to out. Passing a nil out yields a
// transport that only delegates, so callers can wrap unconditionally.
func NewTransport(base http.RoundTripper, out io.Writer) *Transport {
	return &Transport{Base: base, Out: out}
}

// RoundTrip implements http.RoundTripper.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	base := t.Base
	if base == nil {
		base = http.DefaultTransport
	}
	if t.Out != nil {
		fmt.Fprintln(t.Out, Curl(req))
	}
	return base.RoundTrip(req)
}

// Curl renders req as a single-line curl command. The request is not
// modified: the body is re-read through req.GetBody, which net/http populates
// for the in-memory body types the API clients use. When no GetBody is
// available the body is reported as unavailable rather than consumed, since
// draining it would break the request being logged.
func Curl(req *http.Request) string {
	var b strings.Builder
	fmt.Fprintf(&b, "curl -X %s %s", req.Method, shellQuote(req.URL.String()))
	for _, name := range sortedHeaderNames(req.Header) {
		for _, value := range req.Header[name] {
			if strings.EqualFold(name, "Authorization") {
				// Double quotes so the shell expands the placeholder.
				fmt.Fprintf(&b, ` -H "%s: %s"`, name, tokenPlaceholder)
				continue
			}
			fmt.Fprintf(&b, " -H %s", shellQuote(name+": "+value))
		}
	}
	if body := requestBody(req); body != "" {
		fmt.Fprintf(&b, " --data-binary %s", shellQuote(body))
	}
	return b.String()
}

// requestBody returns a copy of the request body, or "" when there is none or
// it cannot be replayed.
func requestBody(req *http.Request) string {
	if req.Body == nil || req.Body == http.NoBody {
		return ""
	}
	if req.GetBody == nil {
		return "<body not shown>"
	}
	rc, err := req.GetBody()
	if err != nil {
		return "<body unavailable>"
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	if err != nil {
		return "<body unavailable>"
	}
	return string(data)
}

func sortedHeaderNames(h http.Header) []string {
	names := make([]string, 0, len(h))
	for name := range h {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// shellQuote wraps s in single quotes, ending and reopening the quoting
// around any single quote it contains, which is the usual POSIX idiom.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
