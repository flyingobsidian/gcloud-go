package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	healthcare "google.golang.org/api/healthcare/v1"
)

// --- gcloud healthcare datasets (#1221) ---

func (s *healthcareCmds) initHealthcareDatasets() {
	all := []*cobra.Command{
		s.healthcareDsCreateCmd, s.healthcareDsDeleteCmd, s.healthcareDsDescribeCmd,
		s.healthcareDsListCmd, s.healthcareDsUpdateCmd, s.healthcareDsGetIamCmd,
		s.healthcareDsSetIamCmd, s.healthcareDsDeidentifyCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagHcDsLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagHcDsFormat, "format", "", "Output format")
	}
	s.healthcareDsCreateCmd.Flags().StringVar(&s.flagHcDsConfigFile, "config-file", "", "Optional YAML/JSON file with the Dataset body")
	s.healthcareDsListCmd.Flags().Int64Var(&s.flagHcDsPageSize, "page-size", 0, "Maximum results per page")
	s.healthcareDsUpdateCmd.Flags().StringVar(&s.flagHcDsConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.healthcareDsUpdateCmd.MarkFlagRequired("config-file")
	s.healthcareDsUpdateCmd.Flags().StringVar(&s.flagHcDsUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	s.healthcareDsDeidentifyCmd.Flags().StringVar(&s.flagHcDsConfigFile, "config-file", "", "YAML/JSON file with the DeidentifyDatasetRequest body (required)")
	_ = s.healthcareDsDeidentifyCmd.MarkFlagRequired("config-file")
	s.healthcareDsDeidentifyCmd.Flags().StringVar(&s.flagHcDsDestDataset, "destination-dataset", "", "Fully qualified destination dataset (overrides destinationDataset in the config file)")

	s.healthcareDatasetsCmd.AddCommand(all...)
	s.healthcareCmd.AddCommand(s.healthcareDatasetsCmd)
}

func (s *healthcareCmds) runHcDsCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.hcLocationParent(s.flagHcDsLocation)
	if err != nil {
		return err
	}
	body := &healthcare.Dataset{}
	if s.flagHcDsConfigFile != "" {
		if err := loadYAMLOrJSONInto(s.flagHcDsConfigFile, body); err != nil {
			return err
		}
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.Create(parent, body).DatasetId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating dataset: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created dataset [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagHcDsFormat)
}

func (s *healthcareCmds) runHcDsDelete(cmd *cobra.Command, args []string) error {
	name, err := s.hcDatasetName(s.flagHcDsLocation, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Datasets.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting dataset: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted dataset [%s].\n", args[0])
	return nil
}

func (s *healthcareCmds) runHcDsDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.hcDatasetName(s.flagHcDsLocation, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing dataset: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagHcDsFormat)
}

func (s *healthcareCmds) runHcDsList(cmd *cobra.Command, args []string) error {
	parent, err := s.hcLocationParent(s.flagHcDsLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	var all []*healthcare.Dataset
	pageToken := ""
	for {
		call := svc.Projects.Locations.Datasets.List(parent).Context(ctx)
		if s.flagHcDsPageSize > 0 {
			call = call.PageSize(s.flagHcDsPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing datasets: %w", err)
		}
		all = append(all, resp.Datasets...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagHcDsFormat)
}

func (s *healthcareCmds) runHcDsUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.hcDatasetName(s.flagHcDsLocation, args[0])
	if err != nil {
		return err
	}
	body := &healthcare.Dataset{}
	if err := loadYAMLOrJSONInto(s.flagHcDsConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagHcDsUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Datasets.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating dataset: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated dataset [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagHcDsFormat)
}

func (s *healthcareCmds) runHcDsGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.hcDatasetName(s.flagHcDsLocation, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Datasets.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagHcDsFormat)
}

func (s *healthcareCmds) runHcDsSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.hcDatasetName(s.flagHcDsLocation, args[0])
	if err != nil {
		return err
	}
	policy := &healthcare.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.Datasets.SetIamPolicy(name, &healthcare.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for dataset [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagHcDsFormat)
}

func (s *healthcareCmds) runHcDsDeidentify(cmd *cobra.Command, args []string) error {
	src, err := s.hcDatasetName(s.flagHcDsLocation, args[0])
	if err != nil {
		return err
	}
	body := &healthcare.DeidentifyDatasetRequest{}
	if err := loadYAMLOrJSONInto(s.flagHcDsConfigFile, body); err != nil {
		return err
	}
	if s.flagHcDsDestDataset != "" {
		body.DestinationDataset = s.flagHcDsDestDataset
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Datasets.Deidentify(src, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deidentifying dataset: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deidentify dataset [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagHcDsFormat)
}
