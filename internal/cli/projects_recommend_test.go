package cli

import "testing"

func TestProjectRecommendFlag(t *testing.T) {
	for _, c := range []string{"delete", "remove-iam-policy-binding"} {
		inv, _, _ := newTestInvocation(t, nil, "")
		cmd := findSub(newProjectsCmd(inv), c)
		if cmd == nil || cmd.Flags().Lookup("recommend") == nil {
			t.Errorf("--recommend missing on projects %s", c)
		}
	}
	for _, v := range []string{"true", "False", "yes", "NO"} {
		if err := validateArgBoolean("recommend", v); err != nil {
			t.Errorf("%q: %v", v, err)
		}
	}
	if err := validateArgBoolean("recommend", "maybe"); err == nil {
		t.Error("expected error for maybe")
	}
}
