package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// --- gcloud design-center operations (#1534) ---

func (s *designCenterCmds) initDesignCenterOperations() {
	all := []*cobra.Command{s.dcOpCancelCmd, s.dcOpDeleteCmd, s.dcOpDescribeCmd, s.dcOpListCmd, s.dcOpWaitCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDCOpLocation, "location", "", "Design Center location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagDCOpFormat, "format", "", "Output format")
	}
	s.dcOpListCmd.Flags().StringVar(&s.flagDCOpFilter, "filter", "", "Server-side filter expression")
	s.dcOpListCmd.Flags().Int64Var(&s.flagDCOpPageSize, "page-size", 0, "Maximum results per page")
	s.dcOpWaitCmd.Flags().DurationVar(&s.flagDCOpTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the operation to finish")

	s.dcOpCmd.AddCommand(all...)
	s.designCenterCmd.AddCommand(s.dcOpCmd)
}

func (s *designCenterCmds) dcOpQualifiedName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/operations/%s", dcLocationName(project, s.flagDCOpLocation), id), nil
}

func (s *designCenterCmds) runDCOpCancel(cmd *cobra.Command, args []string) error {
	name, err := s.dcOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if err := s.inv.rest(designCenterRest).do(ctx, http.MethodPost, "/"+name+":cancel", nil, map[string]any{}, nil); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancel request issued for operation %s.\n", args[0])
	return nil
}

func (s *designCenterCmds) runDCOpDelete(cmd *cobra.Command, args []string) error {
	name, err := s.dcOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	if err := s.inv.rest(designCenterRest).do(ctx, http.MethodDelete, "/"+name, nil, nil, nil); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted operation %s.\n", args[0])
	return nil
}

func (s *designCenterCmds) runDCOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.dcOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(designCenterRest).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDCOpFormat)
}

func (s *designCenterCmds) runDCOpList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	base := url.Values{}
	if s.flagDCOpFilter != "" {
		base.Set("filter", s.flagDCOpFilter)
	}
	ctx := cmd.Context()
	parent := fmt.Sprintf("%s/operations", dcLocationName(project, s.flagDCOpLocation))
	items, err := s.inv.rest(designCenterRest).paginate(ctx, "/"+parent, base, "operations", s.flagDCOpPageSize)
	if err != nil {
		return fmt.Errorf("listing operations: %w", err)
	}
	if s.flagDCOpFormat != "" {
		return emitFormattedTo(s.inv.stdout, items, s.flagDCOpFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(items), "%-40s %s\n", "NAME", "DONE")
	for _, o := range items {
		name, _ := o["name"].(string)
		done, _ := o["done"].(bool)
		fmt.Fprintf(s.inv.stdout, "%-40s %v\n", path.Base(name), done)
	}
	return nil
}

func (s *designCenterCmds) runDCOpWait(cmd *cobra.Command, args []string) error {
	name, err := s.dcOpQualifiedName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	op, err := s.inv.rest(designCenterRest).waitOperation(ctx, name, s.flagDCOpTimeout)
	if err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Operation %s completed.\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagDCOpFormat)
}
