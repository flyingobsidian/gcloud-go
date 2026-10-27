package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- --recommend on projects delete / remove-iam-policy-binding (#1939) ---
//
// gcloud-python's beta track keeps --recommend for compatibility but, since
// 588.0.0, it has no effect: the Active Assist change-risk recommendations
// it showed were retired. gcloud-go does not split tracks, so the flag is
// accepted (and validated as gcloud's ArgBoolean) on both commands and
// otherwise ignored.

var flagProjectRecommend string

func addProjectRecommendFlag(c *cobra.Command) {
	c.Flags().StringVar(&flagProjectRecommend, "recommend", "false",
		"This flag has no effect. It previously added a change-risk warning from Active Assist, which has been retired.")
	prev := c.PreRunE
	c.PreRunE = func(cmd *cobra.Command, args []string) error {
		if err := validateArgBoolean("recommend", flagProjectRecommend); err != nil {
			return err
		}
		if prev != nil {
			return prev(cmd, args)
		}
		return nil
	}
}

// validateArgBoolean mirrors arg_parsers.ArgBoolean's default truthy/falsey
// strings.
func validateArgBoolean(flag, v string) error {
	switch strings.ToLower(v) {
	case "true", "yes", "false", "no":
		return nil
	}
	return fmt.Errorf("argument --%s: Invalid flag value [%s], expected one of [true, yes, false, no]", flag, v)
}

func init() {
	addProjectRecommendFlag(projectDeleteCmd)
	addProjectRecommendFlag(projectRemoveIamBindingCmd)
}
