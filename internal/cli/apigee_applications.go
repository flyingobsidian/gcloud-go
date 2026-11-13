package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee applications (#1376) ---

func (s *apigeeCmds) initApigeeApplications() {
	all := []*cobra.Command{s.apigeeAppDescribeCmd, s.apigeeAppListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagApigeeAppOrganization, "organization", "", "Apigee organization (required)")
		_ = c.MarkFlagRequired("organization")
		c.Flags().StringVar(&s.flagApigeeAppFormat, "format", "", "Output format")
	}
	s.apigeeAppListCmd.Flags().Int64Var(&s.flagApigeeAppPageSize, "page-size", 0, "Maximum results per page")

	s.apigeeApplicationsCmd.AddCommand(all...)
	s.apigeeCmd.AddCommand(s.apigeeApplicationsCmd)
}

func (s *apigeeCmds) apigeeAppName(id string) (string, error) {
	return apigeeResource(s.flagApigeeAppOrganization, "apps", id)
}

func (s *apigeeCmds) runApigeeAppDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.apigeeAppName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, name)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.Apps.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing application: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagApigeeAppFormat)
}

func (s *apigeeCmds) runApigeeAppList(cmd *cobra.Command, args []string) error {
	parent, err := apigeeOrgName(s.flagApigeeAppOrganization)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.apigeeServiceFor(ctx, parent)
	if err != nil {
		return err
	}
	var all []*apigee.GoogleCloudApigeeV1App
	pageToken := ""
	for {
		call := svc.Organizations.Apps.List(parent).Context(ctx)
		if s.flagApigeeAppPageSize > 0 {
			call = call.PageSize(s.flagApigeeAppPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing applications: %w", err)
		}
		all = append(all, resp.App...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagApigeeAppFormat)
}
