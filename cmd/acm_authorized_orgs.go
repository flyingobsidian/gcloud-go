package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	accesscontextmanager "google.golang.org/api/accesscontextmanager/v1"
)

// --- gcloud access-context-manager authorized-orgs (#1441) ---

func (s *accessContextManagerCmds) initAcmAuthorizedOrgs() {
	all := []*cobra.Command{s.acmAOCreateCmd, s.acmAODeleteCmd, s.acmAODescribeCmd, s.acmAOListCmd, s.acmAOUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagACMAOFormat, "format", "", "Output format")
		c.Flags().StringVar(&s.flagACMAOPolicy, "policy", "", "Access policy ID (required)")
		_ = c.MarkFlagRequired("policy")
	}
	for _, c := range []*cobra.Command{s.acmAOCreateCmd, s.acmAOUpdateCmd} {
		c.Flags().StringVar(&s.flagACMAOConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the AuthorizedOrgsDesc body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.acmAOListCmd.Flags().Int64Var(&s.flagACMAOPageSize, "page-size", 0, "Maximum results per page")

	s.acmAOCmd.AddCommand(all...)
	s.accessContextManagerCmd.AddCommand(s.acmAOCmd)
}

func acmAOResource(policy, name string) string {
	return fmt.Sprintf("%s/authorizedOrgsDescs/%s", acmPolicyResource(policy), name)
}

func (s *accessContextManagerCmds) runACMAOCreate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.AuthorizedOrgsDesc{}
	if err := loadYAMLOrJSONInto(s.flagACMAOConfigFile, body); err != nil {
		return err
	}
	body.Name = acmAOResource(s.flagACMAOPolicy, args[0])
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.AuthorizedOrgsDescs.Create(acmPolicyResource(s.flagACMAOPolicy), body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating authorized orgs desc: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for authorized orgs desc [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagACMAOFormat)
}

func (s *accessContextManagerCmds) runACMAODelete(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.AccessPolicies.AuthorizedOrgsDescs.Delete(acmAOResource(s.flagACMAOPolicy, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting authorized orgs desc: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for authorized orgs desc [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagACMAOFormat)
}

func (s *accessContextManagerCmds) runACMAODescribe(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.AccessPolicies.AuthorizedOrgsDescs.Get(acmAOResource(s.flagACMAOPolicy, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing authorized orgs desc: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagACMAOFormat)
}

func (s *accessContextManagerCmds) runACMAOList(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	var all []*accesscontextmanager.AuthorizedOrgsDesc
	pageToken := ""
	for {
		call := svc.AccessPolicies.AuthorizedOrgsDescs.List(acmPolicyResource(s.flagACMAOPolicy)).Context(ctx)
		if s.flagACMAOPageSize > 0 {
			call = call.PageSize(s.flagACMAOPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing authorized orgs descs: %w", err)
		}
		all = append(all, resp.AuthorizedOrgsDescs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagACMAOFormat)
}

func (s *accessContextManagerCmds) runACMAOUpdate(cmd *cobra.Command, args []string) error {
	body := &accesscontextmanager.AuthorizedOrgsDesc{}
	if err := loadYAMLOrJSONInto(s.flagACMAOConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AccessContextManager(ctx)
	if err != nil {
		return err
	}
	call := svc.AccessPolicies.AuthorizedOrgsDescs.Patch(acmAOResource(s.flagACMAOPolicy, args[0]), body).Context(ctx)
	if mask := joinMask(nonEmptyJSONFields(body)); mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating authorized orgs desc: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for authorized orgs desc [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagACMAOFormat)
}
