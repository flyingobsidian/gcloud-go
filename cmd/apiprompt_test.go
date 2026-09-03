package cmd

import (
	"fmt"
	"testing"

	"google.golang.org/api/googleapi"
)

func TestApiNotEnabledService(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "non-googleapi error",
			err:  fmt.Errorf("some other error"),
			want: "",
		},
		{
			name: "404 googleapi is ignored",
			err:  &googleapi.Error{Code: 404, Message: "not found"},
			want: "",
		},
		{
			name: "structured SERVICE_DISABLED detail wins",
			err: &googleapi.Error{
				Code: 403,
				Details: []any{
					map[string]any{
						"@type":  "type.googleapis.com/google.rpc.ErrorInfo",
						"reason": "SERVICE_DISABLED",
						"metadata": map[string]any{
							"service": "spanner.googleapis.com",
						},
					},
				},
				Message: "irrelevant",
			},
			want: "spanner.googleapis.com",
		},
		{
			name: "prose fallback via overview URL",
			err: &googleapi.Error{
				Code: 403,
				Message: "Cloud Spanner API has not been used in project foo before or it is disabled. " +
					"Enable it by visiting https://console.developers.google.com/apis/api/spanner.googleapis.com/overview?project=foo then retry.",
			},
			want: "spanner.googleapis.com",
		},
		{
			name: "prose without has-not-been-used sentence is ignored",
			err: &googleapi.Error{
				Code:    403,
				Message: "Permission denied on /apis/api/spanner.googleapis.com/overview",
			},
			want: "",
		},
		{
			name: "structured ErrorInfo with different reason is skipped",
			err: &googleapi.Error{
				Code: 403,
				Details: []any{
					map[string]any{
						"reason": "IAM_PERMISSION_DENIED",
						"metadata": map[string]any{
							"service": "spanner.googleapis.com",
						},
					},
				},
				Message: "unrelated",
			},
			want: "",
		},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := apiNotEnabledService(tc.err)
			if got != tc.want {
				t.Errorf("apiNotEnabledService = %q, want %q", got, tc.want)
			}
		})
	}
}
