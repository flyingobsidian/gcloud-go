package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms retired-resources (#1110) ---

func (s *kmsCmds) initKmsRetiredResources() {
	for _, c := range []*cobra.Command{s.kmsRRDescribeCmd, s.kmsRRListCmd} {
		c.Flags().StringVar(&s.flagKmsRRLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&s.flagKmsRRFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}
	s.kmsRRListCmd.Flags().Int64Var(&s.flagKmsRRPageSize, "page-size", 0, "Page size")

	s.kmsRetiredCmd.AddCommand(s.kmsRRDescribeCmd, s.kmsRRListCmd)
	s.kmsCmd.AddCommand(s.kmsRetiredCmd)
}

func kmsRRName(project, location, raw string) string {
	parent := kmsLocationParent(project, location) + "/retiredResources"
	return kmsFullName(parent, raw)
}

func (s *kmsCmds) runKmsRRDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsRRName(project, s.flagKmsRRLocation, args[0])
	out, err := svc.Projects.Locations.RetiredResources.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing retired resource: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsRRFormat)
}

func (s *kmsCmds) runKmsRRList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	parent := kmsLocationParent(project, s.flagKmsRRLocation)
	var all []*cloudkms.RetiredResource
	token := ""
	for {
		call := svc.Projects.Locations.RetiredResources.List(parent).Context(ctx)
		if s.flagKmsRRPageSize > 0 {
			call = call.PageSize(s.flagKmsRRPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing retired resources: %w", err)
		}
		all = append(all, resp.RetiredResources...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagKmsRRFormat)
}
