package cmd

import (
	"io"
	"os"
	"strings"
)

// httpDebugOut is where curl-format HTTP debug output is written. Package
// level so tests can capture it.
var httpDebugOut io.Writer = os.Stderr

// httpDebugWriter returns the destination for curl-format logging of outgoing
// REST calls, or nil when it is switched off. --verbosity=debug turns it on,
// as does --log-http, which exists for exactly this in gcloud-python.
//
// Output goes to stderr so it never mixes into --format output on stdout.
func httpDebugWriter() io.Writer {
	if flagLogHTTP || strings.EqualFold(flagVerbosity, "debug") {
		return httpDebugOut
	}
	return nil
}
