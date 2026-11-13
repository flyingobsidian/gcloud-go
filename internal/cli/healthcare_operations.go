package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	healthcare "google.golang.org/api/healthcare/v1"
)

// --- gcloud healthcare operations (#1225) ---

func (s *healthcareCmds) initHealthcareOperations() {
	all := []*cobra.Command{s.healthcareOpCancelCmd, s.healthcareOpDescribeCmd, s.healthcareOpListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagHcOpLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagHcOpDataset, "dataset", "", "Dataset that owns the operation (required)")
		_ = c.MarkFlagRequired("dataset")
		c.Flags().StringVar(&s.flagHcOpFormat, "format", "", "Output format")
	}
	s.healthcareOpListCmd.Flags().StringVar(&s.flagHcOpFilter, "filter", "", "Server-side filter expression")
	s.healthcareOpListCmd.Flags().Int64Var(&s.flagHcOpPageSize, "page-size", 0, "Maximum results per page")

	s.healthcareOperationsCmd.AddCommand(all...)
	s.healthcareCmd.AddCommand(s.healthcareOperationsCmd)
}

func (s *healthcareCmds) hcOpName(id string) (string, error) {
	dataset, err := s.hcDatasetName(s.flagHcOpLocation, s.flagHcOpDataset)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/operations/%s", dataset, id), nil
}

func (s *healthcareCmds) runHcOpCancel(cmd *cobra.Command, args []string) error {
	name, err := s.hcOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Datasets.Operations.Cancel(name, &healthcare.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *healthcareCmds) runHcOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.hcOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagHcOpFormat)
}

func (s *healthcareCmds) runHcOpList(cmd *cobra.Command, args []string) error {
	dataset, err := s.hcDatasetName(s.flagHcOpLocation, s.flagHcOpDataset)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	var all []*healthcare.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Datasets.Operations.List(dataset).Context(ctx)
		if s.flagHcOpFilter != "" {
			call = call.Filter(s.flagHcOpFilter)
		}
		if s.flagHcOpPageSize > 0 {
			call = call.PageSize(s.flagHcOpPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagHcOpFormat)
}
