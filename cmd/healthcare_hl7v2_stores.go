package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	healthcare "google.golang.org/api/healthcare/v1"
)

// --- gcloud healthcare hl7v2-stores (#1224) ---

func (s *healthcareCmds) initHealthcareHl7v2Stores() {
	all := []*cobra.Command{
		s.hcH7CreateCmd, s.hcH7DeleteCmd, s.hcH7DescribeCmd,
		s.hcH7ListCmd, s.hcH7UpdateCmd, s.hcH7GetIamCmd, s.hcH7SetIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagHcH7Location, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagHcH7Dataset, "dataset", "", "Parent dataset (required)")
		_ = c.MarkFlagRequired("dataset")
		c.Flags().StringVar(&s.flagHcH7Format, "format", "", "Output format")
	}
	s.hcH7CreateCmd.Flags().StringVar(&s.flagHcH7ConfigFile, "config-file", "", "YAML/JSON file with the Hl7V2Store body (required)")
	_ = s.hcH7CreateCmd.MarkFlagRequired("config-file")
	s.hcH7ListCmd.Flags().Int64Var(&s.flagHcH7PageSize, "page-size", 0, "Maximum results per page")
	s.hcH7UpdateCmd.Flags().StringVar(&s.flagHcH7ConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.hcH7UpdateCmd.MarkFlagRequired("config-file")
	s.hcH7UpdateCmd.Flags().StringVar(&s.flagHcH7UpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")

	s.healthcareHl7v2StoresCmd.AddCommand(all...)
	s.healthcareCmd.AddCommand(s.healthcareHl7v2StoresCmd)
}

// hcHl7V2StoreName returns projects/.../datasets/DATASET/hl7V2Stores/ID.
func (s *healthcareCmds) hcHl7V2StoreName(location, dataset, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	ds, err := s.hcDatasetName(location, dataset)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/hl7V2Stores/%s", ds, id), nil
}

func (s *healthcareCmds) runHcH7Create(cmd *cobra.Command, args []string) error {
	parent, err := s.hcDatasetName(s.flagHcH7Location, s.flagHcH7Dataset)
	if err != nil {
		return err
	}
	body := &healthcare.Hl7V2Store{}
	if err := loadYAMLOrJSONInto(s.flagHcH7ConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.Hl7V2Stores.Create(parent, body).Hl7V2StoreId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating HL7v2 store: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created HL7v2 store [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagHcH7Format)
}

func (s *healthcareCmds) runHcH7Delete(cmd *cobra.Command, args []string) error {
	name, err := s.hcHl7V2StoreName(s.flagHcH7Location, s.flagHcH7Dataset, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Datasets.Hl7V2Stores.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting HL7v2 store: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted HL7v2 store [%s].\n", args[0])
	return nil
}

func (s *healthcareCmds) runHcH7Describe(cmd *cobra.Command, args []string) error {
	name, err := s.hcHl7V2StoreName(s.flagHcH7Location, s.flagHcH7Dataset, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Datasets.Hl7V2Stores.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing HL7v2 store: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagHcH7Format)
}

func (s *healthcareCmds) runHcH7List(cmd *cobra.Command, args []string) error {
	parent, err := s.hcDatasetName(s.flagHcH7Location, s.flagHcH7Dataset)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	var all []*healthcare.Hl7V2Store
	pageToken := ""
	for {
		call := svc.Projects.Locations.Datasets.Hl7V2Stores.List(parent).Context(ctx)
		if s.flagHcH7PageSize > 0 {
			call = call.PageSize(s.flagHcH7PageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing HL7v2 stores: %w", err)
		}
		all = append(all, resp.Hl7V2Stores...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagHcH7Format)
}

func (s *healthcareCmds) runHcH7Update(cmd *cobra.Command, args []string) error {
	name, err := s.hcHl7V2StoreName(s.flagHcH7Location, s.flagHcH7Dataset, args[0])
	if err != nil {
		return err
	}
	body := &healthcare.Hl7V2Store{}
	if err := loadYAMLOrJSONInto(s.flagHcH7ConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagHcH7UpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Datasets.Hl7V2Stores.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	got, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating HL7v2 store: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated HL7v2 store [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagHcH7Format)
}

func (s *healthcareCmds) runHcH7GetIam(cmd *cobra.Command, args []string) error {
	name, err := s.hcHl7V2StoreName(s.flagHcH7Location, s.flagHcH7Dataset, args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Healthcare(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Datasets.Hl7V2Stores.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagHcH7Format)
}

func (s *healthcareCmds) runHcH7SetIam(cmd *cobra.Command, args []string) error {
	name, err := s.hcHl7V2StoreName(s.flagHcH7Location, s.flagHcH7Dataset, args[0])
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
	updated, err := svc.Projects.Locations.Datasets.Hl7V2Stores.SetIamPolicy(name, &healthcare.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for HL7v2 store [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, updated, s.flagHcH7Format)
}
