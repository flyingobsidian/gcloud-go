package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// --- gcloud kms key-handles (#1105) ---

func (s *kmsCmds) initKmsKeyHandles() {
	for _, c := range []*cobra.Command{s.kmsKHCreateCmd, s.kmsKHDescribeCmd, s.kmsKHListCmd} {
		c.Flags().StringVar(&s.flagKmsKHLocation, "location", "", "Location (required)")
		c.Flags().StringVar(&s.flagKmsKHFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
	}

	s.kmsKHCreateCmd.Flags().StringVar(&s.flagKmsKHConfigFile, "config-file", "", "YAML/JSON body for the KeyHandle (required)")
	s.kmsKHCreateCmd.Flags().StringVar(&s.flagKmsKHHandleID, "key-handle-id", "", "Optional KeyHandle id; defaults to a server-generated UUID")
	_ = s.kmsKHCreateCmd.MarkFlagRequired("config-file")

	s.kmsKHListCmd.Flags().StringVar(&s.flagKmsKHFilter, "filter", "", "Filter expression")
	s.kmsKHListCmd.Flags().Int64Var(&s.flagKmsKHPageSize, "page-size", 0, "Page size")

	s.kmsKeyHandlesCmd.AddCommand(s.kmsKHCreateCmd, s.kmsKHDescribeCmd, s.kmsKHListCmd)
	s.kmsCmd.AddCommand(s.kmsKeyHandlesCmd)
}

func kmsKHName(project, location, raw string) string {
	parent := kmsLocationParent(project, location) + "/keyHandles"
	return kmsFullName(parent, raw)
}

func (s *kmsCmds) runKmsKHCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	body := &cloudkms.KeyHandle{}
	if err := loadYAMLOrJSONInto(s.flagKmsKHConfigFile, body); err != nil {
		return err
	}
	parent := kmsLocationParent(project, s.flagKmsKHLocation)
	call := svc.Projects.Locations.KeyHandles.Create(parent, body).Context(ctx)
	if s.flagKmsKHHandleID != "" {
		call = call.KeyHandleId(s.flagKmsKHHandleID)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("creating key handle: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, s.flagKmsKHFormat)
}

func (s *kmsCmds) runKmsKHDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	name := kmsKHName(project, s.flagKmsKHLocation, args[0])
	out, err := svc.Projects.Locations.KeyHandles.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing key handle: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagKmsKHFormat)
}

func (s *kmsCmds) runKmsKHList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().KMS(ctx)
	if err != nil {
		return err
	}
	parent := kmsLocationParent(project, s.flagKmsKHLocation)
	var all []*cloudkms.KeyHandle
	token := ""
	for {
		call := svc.Projects.Locations.KeyHandles.List(parent).Context(ctx)
		if s.flagKmsKHFilter != "" {
			call = call.Filter(s.flagKmsKHFilter)
		}
		if s.flagKmsKHPageSize > 0 {
			call = call.PageSize(s.flagKmsKHPageSize)
		}
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing key handles: %w", err)
		}
		all = append(all, resp.KeyHandles...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagKmsKHFormat)
}
