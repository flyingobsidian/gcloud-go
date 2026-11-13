package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee operations (#1381) ---

func (s *apigeeCmds) initApigeeOperations() {
	all := []*cobra.Command{s.apigeeOpDescribeCmd, s.apigeeOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagApigeeOpOrganization, "organization", "", "Apigee organization (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&s.flagApigeeOpFormat, "format", "", "Output format")
	}
	s.apigeeOpListCmd.Flags().Int64Var(&s.flagApigeeOpPageSize, "page-size", 0, "Maximum results per page")

	s.apigeeOperationsCmd.AddCommand(all...)
	s.apigeeCmd.AddCommand(s.apigeeOperationsCmd)
}

func (s *apigeeCmds) apigeeOpName(id string) (string, error) {
	return apigeeResource(s.flagApigeeOpOrganization, "operations", id)
}

func (s *apigeeCmds) runApigeeOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.apigeeOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, name)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagApigeeOpFormat)
}

func (s *apigeeCmds) runApigeeOpList(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(s.flagApigeeOpOrganization)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, parent)
	if err != nil {
		return err
	}
	var all []*apigee.GoogleLongrunningOperation
	pageToken := ""
	for {
		call := svc.Organizations.Operations.List(parent).Context(ctx)
		if s.flagApigeeOpPageSize > 0 {
			call = call.PageSize(s.flagApigeeOpPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagApigeeOpFormat)
}
