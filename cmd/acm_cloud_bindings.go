package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
)

// --- gcloud access-context-manager cloud-bindings (#1442) ---

func (s *accessContextManagerCmds) initAcmCloudBindings() {
	all := []*cobra.Command{s.acmCBCreateCmd, s.acmCBDeleteCmd, s.acmCBDescribeCmd, s.acmCBListCmd, s.acmCBUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagACMCBFormat, "format", "", "Output format")
		c.Flags().StringVar(&s.flagACMCBOrganization, "organization", "", "Organization ID (required)")
		_ = c.MarkFlagRequired("organization")
	}
	for _, c := range []*cobra.Command{s.acmCBCreateCmd, s.acmCBUpdateCmd} {
		c.Flags().StringVar(&s.flagACMCBConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the GcpUserAccessBinding body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.acmCBListCmd.Flags().Int64Var(&s.flagACMCBPageSize, "page-size", 0, "Maximum results per page")

	s.acmCBCmd.AddCommand(all...)
	s.accessContextManagerCmd.AddCommand(s.acmCBCmd)
}

func acmCBParent(org string) string {
	return "organizations/" + org
}

func acmCBResource(org, binding string) string {
	return fmt.Sprintf("%s/gcpUserAccessBindings/%s", acmCBParent(org), binding)
}

func (s *accessContextManagerCmds) runACMCBCreate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.GcpUserAccessBinding{}
	if err := loadYAMLOrJSONInto(s.flagACMCBConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.GcpUserAccessBindings.Create(acmCBParent(s.flagACMCBOrganization), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating cloud access binding: %w", err)
	}
	fmt.Fprintln(s.inv.stdout, "Create request issued for cloud access binding.")
	return emitFormattedTo(s.inv.stdout, op, s.flagACMCBFormat)
}

func (s *accessContextManagerCmds) runACMCBDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Organizations.GcpUserAccessBindings.Delete(acmCBResource(s.flagACMCBOrganization, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting cloud access binding: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for cloud access binding [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagACMCBFormat)
}

func (s *accessContextManagerCmds) runACMCBDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Organizations.GcpUserAccessBindings.Get(acmCBResource(s.flagACMCBOrganization, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing cloud access binding: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagACMCBFormat)
}

func (s *accessContextManagerCmds) runACMCBList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	var all []*accesscontextmanager.GcpUserAccessBinding
	pageToken := ""
	for {
		call := svc.Organizations.GcpUserAccessBindings.List(acmCBParent(s.flagACMCBOrganization)).Context(ctx)
		if s.flagACMCBPageSize > 0 {
			call = call.PageSize(s.flagACMCBPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing cloud access bindings: %w", err)
		}
		all = append(all, resp.GcpUserAccessBindings...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagACMCBFormat)
}

func (s *accessContextManagerCmds) runACMCBUpdate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.GcpUserAccessBinding{}
	if err := loadYAMLOrJSONInto(s.flagACMCBConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	call := svc.Organizations.GcpUserAccessBindings.Patch(acmCBResource(s.flagACMCBOrganization, args[0]), body).Context(ctx)
	if mask := joinMask(nonEmptyJSONFields(body)); mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating cloud access binding: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for cloud access binding [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagACMCBFormat)
}
