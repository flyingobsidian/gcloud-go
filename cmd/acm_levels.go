package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
)

// --- gcloud access-context-manager levels (#1443) ---

func (s *accessContextManagerCmds) initAcmLevels() {
	all := []*cobra.Command{
		s.acmLVCreateCmd, s.acmLVDeleteCmd, s.acmLVDescribeCmd, s.acmLVListCmd, s.acmLVUpdateCmd, s.acmLVReplaceAllCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagACMLVFormat, "format", "", "Output format")
		c.Flags().StringVar(&s.flagACMLVPolicy, "policy", "", "Access policy ID (required)")
		_ = c.MarkFlagRequired("policy")
	}
	for _, c := range []*cobra.Command{s.acmLVCreateCmd, s.acmLVUpdateCmd} {
		c.Flags().StringVar(&s.flagACMLVConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the AccessLevel body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.acmLVReplaceAllCmd.Flags().StringVar(&s.flagACMLVSourceFile, "source-file", "",
		"Path to a YAML/JSON file containing a list of AccessLevel bodies (required)")
	_ = s.acmLVReplaceAllCmd.MarkFlagRequired("source-file")
	s.acmLVReplaceAllCmd.Flags().StringVar(&s.flagACMLVEtag, "etag", "", "Expected etag of the parent AccessPolicy")
	s.acmLVListCmd.Flags().Int64Var(&s.flagACMLVPageSize, "page-size", 0, "Maximum results per page")

	condAll := []*cobra.Command{s.acmLVCondCreateCmd, s.acmLVCondDeleteCmd, s.acmLVCondListCmd}
	for _, c := range condAll {
		c.Flags().StringVar(&s.flagACMLVFormat, "format", "", "Output format")
		c.Flags().StringVar(&s.flagACMLVPolicy, "policy", "", "Access policy ID (required)")
		_ = c.MarkFlagRequired("policy")
	}
	s.acmLVCondCreateCmd.Flags().StringVar(&s.flagACMLVCondConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Condition body (required)")
	_ = s.acmLVCondCreateCmd.MarkFlagRequired("config-file")
	s.acmLVCondDeleteCmd.Flags().IntVar(&s.flagACMLVCondIndex, "index", -1,
		"Zero-based index of the condition to remove (required)")
	_ = s.acmLVCondDeleteCmd.MarkFlagRequired("index")

	s.acmLVCondCmd.AddCommand(condAll...)
	s.acmLVCmd.AddCommand(all...)
	s.acmLVCmd.AddCommand(s.acmLVCondCmd)
	s.accessContextManagerCmd.AddCommand(s.acmLVCmd)
}

func acmLVResource(policy, name string) string {
	return fmt.Sprintf("%s/accessLevels/%s", acmPolicyResource(policy), name)
}

func (s *accessContextManagerCmds) runACMLVCreate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.AccessLevel{}
	if err := loadYAMLOrJSONInto(s.flagACMLVConfigFile, body); err != nil {
		return err
	}
	body.Name = acmLVResource(s.flagACMLVPolicy, args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.AccessLevels.Create(acmPolicyResource(s.flagACMLVPolicy), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating access level: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for access level [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagACMLVFormat)
}

func (s *accessContextManagerCmds) runACMLVDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.AccessLevels.Delete(acmLVResource(s.flagACMLVPolicy, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting access level: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for access level [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagACMLVFormat)
}

func (s *accessContextManagerCmds) runACMLVDescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.AccessPolicies.AccessLevels.Get(acmLVResource(s.flagACMLVPolicy, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing access level: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagACMLVFormat)
}

func (s *accessContextManagerCmds) runACMLVList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	var all []*accesscontextmanager.AccessLevel
	pageToken := ""
	for {
		call := svc.AccessPolicies.AccessLevels.List(acmPolicyResource(s.flagACMLVPolicy)).Context(ctx)
		if s.flagACMLVPageSize > 0 {
			call = call.PageSize(s.flagACMLVPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing access levels: %w", err)
		}
		all = append(all, resp.AccessLevels...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagACMLVFormat)
}

func (s *accessContextManagerCmds) runACMLVUpdate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.AccessLevel{}
	if err := loadYAMLOrJSONInto(s.flagACMLVConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	call := svc.AccessPolicies.AccessLevels.Patch(acmLVResource(s.flagACMLVPolicy, args[0]), body).Context(ctx)
	if mask := joinMask(nonEmptyJSONFields(body)); mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating access level: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for access level [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagACMLVFormat)
}

func (s *accessContextManagerCmds) runACMLVReplaceAll(cmd *cobra.Command, args []string) error {
	var levels []*accesscontextmanager.AccessLevel
	if err := loadYAMLOrJSONInto(s.flagACMLVSourceFile, &levels); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	req := &accesscontextmanager.ReplaceAccessLevelsRequest{
		AccessLevels: levels,
		Etag:         s.flagACMLVEtag,
	}
	op, err := svc.AccessPolicies.AccessLevels.ReplaceAll(acmPolicyResource(s.flagACMLVPolicy), req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("replacing access levels: %w", err)
	}
	fmt.Fprintln(s.inv.stdout, "Replace-all request issued for access levels.")
	return emitFormattedTo(s.inv.stdout, op, s.flagACMLVFormat)
}

func acmLVFetchForPatch(ctx context.Context, svc *accesscontextmanager.Service, policy, name string) (*accesscontextmanager.AccessLevel, error) {
	level, err := svc.AccessPolicies.AccessLevels.Get(acmLVResource(policy, name)).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("getting access level: %w", err)
	}
	if level.Basic == nil {
		level.Basic = &accesscontextmanager.BasicLevel{}
	}
	return level, nil
}

func (s *accessContextManagerCmds) runACMLVCondCreate(cmd *cobra.Command, args []string) error {
	cond := &accesscontextmanager.Condition{}
	if err := loadYAMLOrJSONInto(s.flagACMLVCondConfigFile, cond); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	level, err := acmLVFetchForPatch(ctx, svc, s.flagACMLVPolicy, args[0])
	if err != nil {
		return err
	}
	level.Basic.Conditions = append(level.Basic.Conditions, cond)
	patch := &accesscontextmanager.AccessLevel{Basic: level.Basic}
	op, err := svc.AccessPolicies.AccessLevels.Patch(acmLVResource(s.flagACMLVPolicy, args[0]), patch).
		UpdateMask("basic").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("adding condition: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Added condition to access level [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagACMLVFormat)
}

func (s *accessContextManagerCmds) runACMLVCondDelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	level, err := acmLVFetchForPatch(ctx, svc, s.flagACMLVPolicy, args[0])
	if err != nil {
		return err
	}
	conds := level.Basic.Conditions
	if s.flagACMLVCondIndex < 0 || s.flagACMLVCondIndex >= len(conds) {
		return fmt.Errorf("condition index %d out of range [0, %d)", s.flagACMLVCondIndex, len(conds))
	}
	level.Basic.Conditions = append(conds[:s.flagACMLVCondIndex], conds[s.flagACMLVCondIndex+1:]...)
	patch := &accesscontextmanager.AccessLevel{Basic: level.Basic}
	op, err := svc.AccessPolicies.AccessLevels.Patch(acmLVResource(s.flagACMLVPolicy, args[0]), patch).
		UpdateMask("basic").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("removing condition: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Removed condition [%d] from access level [%s].\n", s.flagACMLVCondIndex, args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagACMLVFormat)
}

func (s *accessContextManagerCmds) runACMLVCondList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	level, err := svc.AccessPolicies.AccessLevels.Get(acmLVResource(s.flagACMLVPolicy, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting access level: %w", err)
	}
	var conds []*accesscontextmanager.Condition
	if level.Basic != nil {
		conds = level.Basic.Conditions
	}
	return emitFormattedTo(s.inv.stdout, conds, s.flagACMLVFormat)
}
