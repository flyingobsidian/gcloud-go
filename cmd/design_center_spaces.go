package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

// --- gcloud design-center spaces (#1535) ---

var dcSpaceCmd = &cobra.Command{Use: "spaces", Short: "Manage Design Center spaces"}

var (
	flagDCSpaceLocation   string
	flagDCSpaceFormat     string
	flagDCSpaceConfigFile string
	flagDCSpaceUpdateMask string
	flagDCSpaceFilter     string
	flagDCSpacePageSize   int64
	flagDCSpaceForce      bool

	flagDCSpaceTestPerms []string
)

var (
	dcSpaceCreateCmd = &cobra.Command{
		Use: "create SPACE", Short: "Create a Design Center space",
		Args: cobra.ExactArgs(1), RunE: runDCSpaceCreate,
	}
	dcSpaceDeleteCmd = &cobra.Command{
		Use: "delete SPACE", Short: "Delete a Design Center space",
		Args: cobra.ExactArgs(1), RunE: runDCSpaceDelete,
	}
	dcSpaceDescribeCmd = &cobra.Command{
		Use: "describe SPACE", Short: "Describe a Design Center space",
		Args: cobra.ExactArgs(1), RunE: runDCSpaceDescribe,
	}
	dcSpaceListCmd = &cobra.Command{
		Use: "list", Short: "List Design Center spaces",
		Args: cobra.NoArgs, RunE: runDCSpaceList,
	}
	dcSpaceUpdateCmd = &cobra.Command{
		Use: "update SPACE", Short: "Update a Design Center space",
		Args: cobra.ExactArgs(1), RunE: runDCSpaceUpdate,
	}
	dcSpaceGetIamCmd = &cobra.Command{
		Use: "get-iam-policy SPACE", Short: "Get the IAM policy for a space",
		Args: cobra.ExactArgs(1), RunE: runDCSpaceGetIam,
	}
	dcSpaceSetIamCmd = &cobra.Command{
		Use: "set-iam-policy SPACE POLICY_FILE", Short: "Set the IAM policy for a space",
		Args: cobra.ExactArgs(2), RunE: runDCSpaceSetIam,
	}
	dcSpaceTestIamCmd = &cobra.Command{
		Use: "test-iam-permissions SPACE", Short: "Test the caller's permissions on a space",
		Args: cobra.ExactArgs(1), RunE: runDCSpaceTestIam,
	}
)

func init() {
	all := []*cobra.Command{
		dcSpaceCreateCmd, dcSpaceDeleteCmd, dcSpaceDescribeCmd, dcSpaceListCmd, dcSpaceUpdateCmd,
		dcSpaceGetIamCmd, dcSpaceSetIamCmd, dcSpaceTestIamCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&flagDCSpaceLocation, "location", "", "Design Center location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagDCSpaceFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{dcSpaceCreateCmd, dcSpaceUpdateCmd} {
		c.Flags().StringVar(&flagDCSpaceConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the Space body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	dcSpaceUpdateCmd.Flags().StringVar(&flagDCSpaceUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	dcSpaceListCmd.Flags().StringVar(&flagDCSpaceFilter, "filter", "", "Server-side filter")
	dcSpaceListCmd.Flags().Int64Var(&flagDCSpacePageSize, "page-size", 0, "Maximum results per page")
	dcSpaceDeleteCmd.Flags().BoolVar(&flagDCSpaceForce, "force", false, "Force delete of a non-empty space")
	dcSpaceTestIamCmd.Flags().StringSliceVar(&flagDCSpaceTestPerms, "permissions", nil,
		"Permissions to test (comma-separated or repeated) (required)")
	_ = dcSpaceTestIamCmd.MarkFlagRequired("permissions")

	dcSpaceCmd.AddCommand(all...)
	designCenterCmd.AddCommand(dcSpaceCmd)
}

func dcSpaceParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return dcLocationName(project, flagDCSpaceLocation), nil
}

func dcSpaceName(id string) (string, error) {
	parent, err := dcSpaceParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/spaces/%s", parent, id), nil
}

func runDCSpaceCreate(cmd *cobra.Command, args []string) error {
	parent, err := dcSpaceParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagDCSpaceConfigFile, &body); err != nil {
		return err
	}
	ctx := context.Background()
	q := url.Values{}
	q.Set("spaceId", args[0])
	var op map[string]any
	if err := dcDo(ctx, http.MethodPost, "/"+parent+"/spaces", q, body, &op); err != nil {
		return fmt.Errorf("creating space: %w", err)
	}
	fmt.Printf("Create request issued for space [%s].\n", args[0])
	return emitFormatted(op, flagDCSpaceFormat)
}

func runDCSpaceDelete(cmd *cobra.Command, args []string) error {
	name, err := dcSpaceName(args[0])
	if err != nil {
		return err
	}
	q := url.Values{}
	if flagDCSpaceForce {
		q.Set("force", "true")
	}
	ctx := context.Background()
	var op map[string]any
	if err := dcDo(ctx, http.MethodDelete, "/"+name, q, nil, &op); err != nil {
		return fmt.Errorf("deleting space: %w", err)
	}
	fmt.Printf("Delete request issued for space [%s].\n", args[0])
	return emitFormatted(op, flagDCSpaceFormat)
}

func runDCSpaceDescribe(cmd *cobra.Command, args []string) error {
	name, err := dcSpaceName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := dcDo(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing space: %w", err)
	}
	return emitFormatted(got, flagDCSpaceFormat)
}

func runDCSpaceList(cmd *cobra.Command, args []string) error {
	parent, err := dcSpaceParent()
	if err != nil {
		return err
	}
	base := url.Values{}
	if flagDCSpaceFilter != "" {
		base.Set("filter", flagDCSpaceFilter)
	}
	ctx := context.Background()
	items, err := dcPaginate(ctx, "/"+parent+"/spaces", base, "spaces", flagDCSpacePageSize)
	if err != nil {
		return fmt.Errorf("listing spaces: %w", err)
	}
	return emitFormatted(items, flagDCSpaceFormat)
}

func runDCSpaceUpdate(cmd *cobra.Command, args []string) error {
	name, err := dcSpaceName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagDCSpaceConfigFile, &body); err != nil {
		return err
	}
	mask := flagDCSpaceUpdateMask
	if mask == "" {
		mask = joinMask(dcTopLevelKeys(body))
	}
	q := url.Values{}
	if mask != "" {
		q.Set("updateMask", mask)
	}
	ctx := context.Background()
	var op map[string]any
	if err := dcDo(ctx, http.MethodPatch, "/"+name, q, body, &op); err != nil {
		return fmt.Errorf("updating space: %w", err)
	}
	fmt.Printf("Update request issued for space [%s].\n", args[0])
	return emitFormatted(op, flagDCSpaceFormat)
}

func runDCSpaceGetIam(cmd *cobra.Command, args []string) error {
	name, err := dcSpaceName(args[0])
	if err != nil {
		return err
	}
	q := url.Values{}
	q.Set("options.requestedPolicyVersion", "3")
	ctx := context.Background()
	var got map[string]any
	if err := dcDo(ctx, http.MethodGet, "/"+name+":getIamPolicy", q, nil, &got); err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(got, flagDCSpaceFormat)
}

func runDCSpaceSetIam(cmd *cobra.Command, args []string) error {
	name, err := dcSpaceName(args[0])
	if err != nil {
		return err
	}
	policy := map[string]any{}
	if err := loadYAMLOrJSONInto(args[1], &policy); err != nil {
		return err
	}
	policy["version"] = 3
	ctx := context.Background()
	var got map[string]any
	if err := dcDo(ctx, http.MethodPost, "/"+name+":setIamPolicy", nil, map[string]any{"policy": policy}, &got); err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, flagDCSpaceFormat)
}

func runDCSpaceTestIam(cmd *cobra.Command, args []string) error {
	name, err := dcSpaceName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := dcDo(ctx, http.MethodPost, "/"+name+":testIamPermissions", nil, map[string]any{
		"permissions": flagDCSpaceTestPerms,
	}, &got); err != nil {
		return fmt.Errorf("testing IAM permissions: %w", err)
	}
	return emitFormatted(got, flagDCSpaceFormat)
}

func dcTopLevelKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
