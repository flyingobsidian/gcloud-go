package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestPromptContinueIO(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr string
		wantOut string
	}{
		{"default yes", "\n", "", "msg\n\nDo you want to continue (Y/n)?  \n"},
		{"eof yes", "", "", "msg\n\nDo you want to continue (Y/n)?  \n"},
		{"explicit yes", "Y\n", "", "msg\n\nDo you want to continue (Y/n)?  \n"},
		{"no", "no\n", "Deletion aborted by user.", "msg\n\nDo you want to continue (Y/n)?  \n"},
		{"reprompt", "maybe\nn\n", "Deletion aborted by user.",
			"msg\n\nDo you want to continue (Y/n)?  Please enter 'y' or 'n':  \n"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			err := promptContinueIO(strings.NewReader(tc.input), &out, "msg", "Deletion aborted by user.")
			if tc.wantErr == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantErr != "" && (err == nil || err.Error() != tc.wantErr) {
				t.Fatalf("error = %v, want %q", err, tc.wantErr)
			}
			if out.String() != tc.wantOut {
				t.Errorf("output = %q, want %q", out.String(), tc.wantOut)
			}
		})
	}
}
