package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// --- gcloud pam operations (#964) ---

func (s *pamCmds) initPamOperations() {
	all := []*cobra.Command{s.pamOpDeleteCmd, s.pamOpDescribeCmd, s.pamOpListCmd, s.pamOpWaitCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagPAMOpLocation, "location", "",
			"Location containing the operation (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagPAMOpFormat, "format", "", "Output format")
	}
	s.pamOpListCmd.Flags().Int64Var(&s.flagPAMOpPageSize, "page-size", 0, "Maximum results per page")
	s.pamOpWaitCmd.Flags().DurationVar(&s.flagPAMOpTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the operation to finish")

	s.pamOpCmd.AddCommand(all...)
	s.pamCmd.AddCommand(s.pamOpCmd)
}

func (s *pamCmds) pamOpParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagPAMOpLocation), nil
}

func (s *pamCmds) pamOpName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") ||
		strings.HasPrefix(id, "folders/") ||
		strings.HasPrefix(id, "organizations/") {
		return id, nil
	}
	parent, err := s.pamOpParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/operations/%s", parent, id), nil
}

func (s *pamCmds) runPAMOpDelete(cmd *cobra.Command, args []string) error {
	name, err := s.pamOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if err := s.inv.rest(pamRest).do(ctx, http.MethodDelete, "/"+name, nil, nil, nil); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func (s *pamCmds) runPAMOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.pamOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(pamRest).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagPAMOpFormat)
}

func (s *pamCmds) runPAMOpList(cmd *cobra.Command, args []string) error {
	parent, err := s.pamOpParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(pamRest).paginate(ctx, "/"+parent+"/operations", nil, "operations", s.flagPAMOpPageSize)
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagPAMOpFormat)
}

func (s *pamCmds) runPAMOpWait(cmd *cobra.Command, args []string) error {
	name, err := s.pamOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	op, err := s.inv.rest(pamRest).waitOperation(ctx, name, s.flagPAMOpTimeout)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Operation %s completed.\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagPAMOpFormat)
}
