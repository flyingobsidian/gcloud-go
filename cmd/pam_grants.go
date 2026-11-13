package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud pam grants (#963) ---

func (s *pamCmds) initPamGrants() {
	all := []*cobra.Command{
		s.pamGrantApproveCmd, s.pamGrantCreateCmd, s.pamGrantDenyCmd, s.pamGrantDescribeCmd,
		s.pamGrantListCmd, s.pamGrantRevokeCmd, s.pamGrantSearchCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagPAMGrantLocation, "location", "",
			"Location of the parent entitlement (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagPAMGrantFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.pamGrantCreateCmd, s.pamGrantListCmd, s.pamGrantSearchCmd} {
		c.Flags().StringVar(&s.flagPAMGrantEntitlement, "entitlement", "",
			"Parent entitlement ID (required)")
		_ = c.MarkFlagRequired("entitlement")
	}
	// GRANT-taking commands also accept --entitlement so a bare grant ID can be
	// combined with the location + entitlement to form the full resource path.
	for _, c := range []*cobra.Command{
		s.pamGrantApproveCmd, s.pamGrantDenyCmd, s.pamGrantDescribeCmd, s.pamGrantRevokeCmd,
	} {
		c.Flags().StringVar(&s.flagPAMGrantEntitlement, "entitlement", "",
			"Parent entitlement ID (required unless a full grant resource name is passed)")
	}
	s.pamGrantCreateCmd.Flags().StringVar(&s.flagPAMGrantConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the grant body (required)")
	_ = s.pamGrantCreateCmd.MarkFlagRequired("config-file")

	for _, c := range []*cobra.Command{s.pamGrantApproveCmd, s.pamGrantDenyCmd, s.pamGrantRevokeCmd} {
		c.Flags().StringVar(&s.flagPAMGrantReason, "reason", "",
			"Optional reason to record on the request")
	}
	s.pamGrantSearchCmd.Flags().StringVar(&s.flagPAMGrantCallerRel, "caller-relationship", "HAD_CREATED",
		"Caller relationship: HAD_CREATED, CAN_APPROVE, or HAD_APPROVED")
	s.pamGrantListCmd.Flags().Int64Var(&s.flagPAMGrantPageSize, "page-size", 0, "Maximum results per page")
	s.pamGrantSearchCmd.Flags().Int64Var(&s.flagPAMGrantPageSize, "page-size", 0, "Maximum results per page")

	s.pamGrantsCmd.AddCommand(all...)
	s.pamCmd.AddCommand(s.pamGrantsCmd)
}

func (s *pamCmds) pamGrantEntitlementParent() (string, error) {
	if strings.HasPrefix(s.flagPAMGrantEntitlement, "projects/") ||
		strings.HasPrefix(s.flagPAMGrantEntitlement, "folders/") ||
		strings.HasPrefix(s.flagPAMGrantEntitlement, "organizations/") {
		return s.flagPAMGrantEntitlement, nil
	}
	if s.flagPAMGrantEntitlement == "" {
		return "", fmt.Errorf("--entitlement is required")
	}
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s/entitlements/%s", project, s.flagPAMGrantLocation, s.flagPAMGrantEntitlement), nil
}

func (s *pamCmds) pamGrantName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") ||
		strings.HasPrefix(id, "folders/") ||
		strings.HasPrefix(id, "organizations/") {
		return id, nil
	}
	parent, err := s.pamGrantEntitlementParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/grants/%s", parent, id), nil
}

func (s *pamCmds) pamGrantReasonBody() map[string]any {
	body := map[string]any{}
	if s.flagPAMGrantReason != "" {
		body["reason"] = s.flagPAMGrantReason
	}
	return body
}

func (s *pamCmds) runPAMGrantApprove(cmd *cobra.Command, args []string) error {
	name, err := s.pamGrantName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(pamRest()).do(ctx, http.MethodPost, "/"+name+":approve", nil, s.pamGrantReasonBody(), &got); err != nil {
		return fmt.Errorf("approving grant: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Approved grant [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagPAMGrantFormat)
}

func (s *pamCmds) runPAMGrantCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.pamGrantEntitlementParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagPAMGrantConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("grantId", args[0])
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(pamRest()).do(ctx, http.MethodPost, "/"+parent+"/grants", q, body, &got); err != nil {
		return fmt.Errorf("creating grant: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created grant [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagPAMGrantFormat)
}

func (s *pamCmds) runPAMGrantDeny(cmd *cobra.Command, args []string) error {
	name, err := s.pamGrantName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(pamRest()).do(ctx, http.MethodPost, "/"+name+":deny", nil, s.pamGrantReasonBody(), &got); err != nil {
		return fmt.Errorf("denying grant: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Denied grant [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagPAMGrantFormat)
}

func (s *pamCmds) runPAMGrantDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.pamGrantName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(pamRest()).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing grant: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagPAMGrantFormat)
}

func (s *pamCmds) runPAMGrantList(cmd *cobra.Command, args []string) error {
	parent, err := s.pamGrantEntitlementParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(pamRest()).paginate(ctx, "/"+parent+"/grants", nil, "grants", s.flagPAMGrantPageSize)
	if err != nil {
		return fmt.Errorf("listing grants: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagPAMGrantFormat)
}

func (s *pamCmds) runPAMGrantRevoke(cmd *cobra.Command, args []string) error {
	name, err := s.pamGrantName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(pamRest()).do(ctx, http.MethodPost, "/"+name+":revoke", nil, s.pamGrantReasonBody(), &got); err != nil {
		return fmt.Errorf("revoking grant: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Revoked grant [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagPAMGrantFormat)
}

func (s *pamCmds) runPAMGrantSearch(cmd *cobra.Command, args []string) error {
	parent, err := s.pamGrantEntitlementParent()
	if err != nil {
		return err
	}
	rel := s.flagPAMGrantCallerRel
	if rel == "" {
		rel = "HAD_CREATED"
	}
	ctx := cmd.Context()
	base := url.Values{}
	base.Set("callerRelationship", rel)
	items, err := s.inv.rest(pamRest()).paginate(ctx, "/"+parent+":searchGrants", base, "grants", s.flagPAMGrantPageSize)
	if err != nil {
		return fmt.Errorf("searching grants: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagPAMGrantFormat)
}
