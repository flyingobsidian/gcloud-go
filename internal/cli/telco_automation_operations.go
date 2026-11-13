package cli

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// --- gcloud telco-automation operations (#847) ---

func (s *telcoAutomationCmds) initTelcoAutomationOperations() {
	all := []*cobra.Command{s.taOpDescribeCmd, s.taOpWaitCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagTAOpLocation, "location", "", "Telco Automation location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagTAOpFormat, "format", "", "Output format")
	}
	s.taOpWaitCmd.Flags().DurationVar(&s.flagTAOpTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the operation to finish")

	s.taOpCmd.AddCommand(all...)
	s.telcoAutomationCmd.AddCommand(s.taOpCmd)
}

func (s *telcoAutomationCmds) taOpQualifiedName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s/operations/%s", project, s.flagTAOpLocation, id), nil
}

func (s *telcoAutomationCmds) runTAOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.taOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(telcoAutomationRest()).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagTAOpFormat)
}

func (s *telcoAutomationCmds) runTAOpWait(cmd *cobra.Command, args []string) error {
	name, err := s.taOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	op, err := s.inv.rest(telcoAutomationRest()).waitOperation(ctx, name, s.flagTAOpTimeout)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Operation %s completed.\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagTAOpFormat)
}
