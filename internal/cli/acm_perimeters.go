package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
)

// --- gcloud access-context-manager perimeters (#1444) ---

func (s *accessContextManagerCmds) initAcmPerimeters() {
	all := []*cobra.Command{
		s.acmPMCreateCmd, s.acmPMDeleteCmd, s.acmPMDescribeCmd, s.acmPMListCmd, s.acmPMUpdateCmd, s.acmPMReplaceAllCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagACMPMFormat, "format", "", "Output format")
		c.Flags().StringVar(&s.flagACMPMPolicy, "policy", "", "Access policy ID (required)")
		_ = c.MarkFlagRequired("policy")
	}
	for _, c := range []*cobra.Command{s.acmPMCreateCmd, s.acmPMUpdateCmd} {
		c.Flags().StringVar(&s.flagACMPMConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the ServicePerimeter body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.acmPMReplaceAllCmd.Flags().StringVar(&s.flagACMPMSourceFile, "source-file", "",
		"Path to a YAML/JSON file containing a list of ServicePerimeter bodies (required)")
	_ = s.acmPMReplaceAllCmd.MarkFlagRequired("source-file")
	s.acmPMReplaceAllCmd.Flags().StringVar(&s.flagACMPMEtag, "etag", "", "Expected etag of the parent AccessPolicy")
	s.acmPMListCmd.Flags().Int64Var(&s.flagACMPMPageSize, "page-size", 0, "Maximum results per page")

	dryAll := []*cobra.Command{
		s.acmPMDryCreateCmd, s.acmPMDryDescribeCmd, s.acmPMDryEnforceCmd, s.acmPMDryResetCmd, s.acmPMDryDropCmd,
	}
	for _, c := range dryAll {
		c.Flags().StringVar(&s.flagACMPMFormat, "format", "", "Output format")
		c.Flags().StringVar(&s.flagACMPMPolicy, "policy", "", "Access policy ID (required)")
		_ = c.MarkFlagRequired("policy")
	}
	s.acmPMDryCreateCmd.Flags().StringVar(&s.flagACMPMConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the dry-run ServicePerimeter body (required)")
	_ = s.acmPMDryCreateCmd.MarkFlagRequired("config-file")
	s.acmPMDryEnforceCmd.Flags().StringVar(&s.flagACMPMEtag, "etag", "", "Expected etag of the parent AccessPolicy")

	s.acmPMDryRunCmd.AddCommand(dryAll...)
	s.acmPMCmd.AddCommand(all...)
	s.acmPMCmd.AddCommand(s.acmPMDryRunCmd)
	s.accessContextManagerCmd.AddCommand(s.acmPMCmd)
}

func acmPMResource(policy, name string) string {
	return fmt.Sprintf("%s/servicePerimeters/%s", acmPolicyResource(policy), name)
}

func (s *accessContextManagerCmds) runACMPMCreate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.ServicePerimeter{}
	if err := loadYAMLOrJSONInto(s.flagACMPMConfigFile, body); err != nil {
		return err
	}
	body.Name = acmPMResource(s.flagACMPMPolicy, args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.ServicePerimeters.Create(acmPolicyResource(s.flagACMPMPolicy), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating service perimeter: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for service perimeter [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagACMPMFormat)
}

func (s *accessContextManagerCmds) runACMPMDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.ServicePerimeters.Delete(acmPMResource(s.flagACMPMPolicy, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting service perimeter: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for service perimeter [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagACMPMFormat)
}

func (s *accessContextManagerCmds) runACMPMDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.AccessPolicies.ServicePerimeters.Get(acmPMResource(s.flagACMPMPolicy, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service perimeter: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagACMPMFormat)
}

func (s *accessContextManagerCmds) runACMPMList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	var all []*accesscontextmanager.ServicePerimeter
	pageToken := ""
	for {
		call := svc.AccessPolicies.ServicePerimeters.List(acmPolicyResource(s.flagACMPMPolicy)).Context(ctx)
		if s.flagACMPMPageSize > 0 {
			call = call.PageSize(s.flagACMPMPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing service perimeters: %w", err)
		}
		all = append(all, resp.ServicePerimeters...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagACMPMFormat)
}

func (s *accessContextManagerCmds) runACMPMUpdate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.ServicePerimeter{}
	if err := loadYAMLOrJSONInto(s.flagACMPMConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	call := svc.AccessPolicies.ServicePerimeters.Patch(acmPMResource(s.flagACMPMPolicy, args[0]), body).Context(ctx)
	if mask := joinMask(nonEmptyJSONFields(body)); mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating service perimeter: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for service perimeter [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagACMPMFormat)
}

func (s *accessContextManagerCmds) runACMPMReplaceAll(cmd *cobra.Command, args []string) error {
	var perimeters []*accesscontextmanager.ServicePerimeter
	if err := loadYAMLOrJSONInto(s.flagACMPMSourceFile, &perimeters); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	req := &accesscontextmanager.ReplaceServicePerimetersRequest{
		ServicePerimeters: perimeters,
		Etag:              s.flagACMPMEtag,
	}
	op, err := svc.AccessPolicies.ServicePerimeters.ReplaceAll(acmPolicyResource(s.flagACMPMPolicy), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("replacing service perimeters: %w", err)
	}
	fmt.Fprintln(s.inv.stdout, "Replace-all request issued for service perimeters.")
	return emitFormattedTo(s.inv.stdout, op, s.flagACMPMFormat)
}

func (s *accessContextManagerCmds) runACMPMDryCreate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.ServicePerimeter{}
	if err := loadYAMLOrJSONInto(s.flagACMPMConfigFile, body); err != nil {
		return err
	}
	body.UseExplicitDryRunSpec = true
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.ServicePerimeters.
		Patch(acmPMResource(s.flagACMPMPolicy, args[0]), body).
		UpdateMask("spec,useExplicitDryRunSpec").
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating dry-run spec: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Dry-run spec created for service perimeter [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagACMPMFormat)
}

func (s *accessContextManagerCmds) runACMPMDryDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.AccessPolicies.ServicePerimeters.Get(acmPMResource(s.flagACMPMPolicy, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing service perimeter: %w", err)
	}
	view := struct {
		Name                  string                                       `json:"name"`
		Status                *accesscontextmanager.ServicePerimeterConfig `json:"status,omitempty"`
		Spec                  *accesscontextmanager.ServicePerimeterConfig `json:"spec,omitempty"`
		UseExplicitDryRunSpec bool                                         `json:"useExplicitDryRunSpec"`
	}{
		Name:                  got.Name,
		Status:                got.Status,
		Spec:                  got.Spec,
		UseExplicitDryRunSpec: got.UseExplicitDryRunSpec,
	}
	return emitFormattedTo(s.inv.stdout, view, s.flagACMPMFormat)
}

func (s *accessContextManagerCmds) runACMPMDryEnforce(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	req := &accesscontextmanager.CommitServicePerimetersRequest{Etag: s.flagACMPMEtag}
	op, err := svc.AccessPolicies.ServicePerimeters.Commit(acmPolicyResource(s.flagACMPMPolicy), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("committing dry-run specs: %w", err)
	}
	fmt.Fprintln(s.inv.stdout, "Enforce request issued for dry-run specs.")
	return emitFormattedTo(s.inv.stdout, op, s.flagACMPMFormat)
}

func (s *accessContextManagerCmds) runACMPMDryDrop(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	body := &accesscontextmanager.ServicePerimeter{
		UseExplicitDryRunSpec: false,
		ForceSendFields:       []string{"UseExplicitDryRunSpec"},
		NullFields:            []string{"Spec"},
	}
	op, err := svc.AccessPolicies.ServicePerimeters.
		Patch(acmPMResource(s.flagACMPMPolicy, args[0]), body).
		UpdateMask("spec,useExplicitDryRunSpec").
		Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("clearing dry-run spec: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Dry-run spec cleared for service perimeter [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagACMPMFormat)
}
