package cmd

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

// --- gcloud design-center spaces (#1535) ---

func (s *designCenterCmds) initDesignCenterSpaces() {
	all := []*cobra.Command{
		s.dcSpaceCreateCmd, s.dcSpaceDeleteCmd, s.dcSpaceDescribeCmd, s.dcSpaceListCmd, s.dcSpaceUpdateCmd,
		s.dcSpaceGetIamCmd, s.dcSpaceSetIamCmd, s.dcSpaceTestIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagDCSpaceLocation, "location", "", "Design Center location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagDCSpaceFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.dcSpaceCreateCmd, s.dcSpaceUpdateCmd} {
		c.Flags().StringVar(&s.flagDCSpaceConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Space body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.dcSpaceUpdateCmd.Flags().StringVar(&s.flagDCSpaceUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.dcSpaceListCmd.Flags().StringVar(&s.flagDCSpaceFilter, "filter", "", "Server-side filter")
	s.dcSpaceListCmd.Flags().Int64Var(&s.flagDCSpacePageSize, "page-size", 0, "Maximum results per page")
	s.dcSpaceDeleteCmd.Flags().BoolVar(&s.flagDCSpaceForce, "force", false, "Force delete of a non-empty space")
	s.dcSpaceTestIamCmd.Flags().StringSliceVar(&s.flagDCSpaceTestPerms, "permissions", nil,
		"Permissions to test (comma-separated or repeated) (required)")
	_ = s.dcSpaceTestIamCmd.MarkFlagRequired("permissions")

	s.dcSpaceCmd.AddCommand(all...)
	s.designCenterCmd.AddCommand(s.dcSpaceCmd)
}

func (s *designCenterCmds) dcSpaceParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return dcLocationName(project, s.flagDCSpaceLocation), nil
}

func (s *designCenterCmds) dcSpaceName(id string) (string, error) {
	parent, err := s.dcSpaceParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/spaces/%s", parent, id), nil
}

func (s *designCenterCmds) runDCSpaceCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.dcSpaceParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagDCSpaceConfigFile, &body); err != nil {
		return err
	}
	ctx := cmd.Context()
	q := url.Values{}
	q.Set("spaceId", args[0])
	var op map[string]any
	if err := s.inv.rest(designCenterRest).do(ctx, http.MethodPost, "/"+parent+"/spaces", q, body, &op); err != nil {
		return fmt.Errorf("creating space: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for space [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagDCSpaceFormat)
}

func (s *designCenterCmds) runDCSpaceDelete(cmd *cobra.Command, args []string) error {
	name, err := s.dcSpaceName(args[0])
	if err != nil {
		return err
	}
	q := url.Values{}
	if s.flagDCSpaceForce {
		q.Set("force", "true")
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(designCenterRest).do(ctx, http.MethodDelete, "/"+name, q, nil, &op); err != nil {
		return fmt.Errorf("deleting space: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for space [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagDCSpaceFormat)
}

func (s *designCenterCmds) runDCSpaceDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.dcSpaceName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(designCenterRest).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing space: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDCSpaceFormat)
}

func (s *designCenterCmds) runDCSpaceList(cmd *cobra.Command, args []string) error {
	parent, err := s.dcSpaceParent()
	if err != nil {
		return err
	}
	base := url.Values{}
	if s.flagDCSpaceFilter != "" {
		base.Set("filter", s.flagDCSpaceFilter)
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(designCenterRest).paginate(ctx, "/"+parent+"/spaces", base, "spaces", s.flagDCSpacePageSize)
	if err != nil {
		return fmt.Errorf("listing spaces: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagDCSpaceFormat)
}

func (s *designCenterCmds) runDCSpaceUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.dcSpaceName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagDCSpaceConfigFile, &body); err != nil {
		return err
	}
	mask := s.flagDCSpaceUpdateMask
	if mask == "" {
		mask = joinMask(dcTopLevelKeys(body))
	}
	q := url.Values{}
	if mask != "" {
		q.Set("updateMask", mask)
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(designCenterRest).do(ctx, http.MethodPatch, "/"+name, q, body, &op); err != nil {
		return fmt.Errorf("updating space: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for space [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagDCSpaceFormat)
}

func (s *designCenterCmds) runDCSpaceGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcSpaceName(args[0])
	if err != nil {
		return err
	}
	q := url.Values{}
	q.Set("options.requestedPolicyVersion", "3")
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(designCenterRest).do(ctx, http.MethodGet, "/"+name+":getIamPolicy", q, nil, &got); err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDCSpaceFormat)
}

func (s *designCenterCmds) runDCSpaceSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcSpaceName(args[0])
	if err != nil {
		return err
	}
	policy := map[string]any{}
	if err := loadYAMLOrJSONInto(args[1], &policy); err != nil {
		return err
	}
	policy["version"] = 3
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(designCenterRest).do(ctx, http.MethodPost, "/"+name+":setIamPolicy", nil, map[string]any{"policy": policy}, &got); err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDCSpaceFormat)
}

func (s *designCenterCmds) runDCSpaceTestIam(cmd *cobra.Command, args []string) error {
	name, err := s.dcSpaceName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(designCenterRest).do(ctx, http.MethodPost, "/"+name+":testIamPermissions", nil, map[string]any{
		"permissions": s.flagDCSpaceTestPerms,
	}, &got); err != nil {
		return fmt.Errorf("testing IAM permissions: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagDCSpaceFormat)
}

func dcTopLevelKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
