package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud memorystore acl-policies (#976) ---

var memstoreACLCmd = &cobra.Command{Use: "acl-policies", Short: "Manage Memorystore ACL policies"}

var (
	flagMemstoreACLLocation   string
	flagMemstoreACLFormat     string
	flagMemstoreACLConfigFile string
	flagMemstoreACLUpdateMask string
	flagMemstoreACLPageSize   int64
)

var (
	memstoreACLCreateCmd = &cobra.Command{
		Use: "create ACL_POLICY", Short: "Create a Memorystore ACL policy",
		Args: cobra.ExactArgs(1), RunE: runMemstoreACLCreate,
	}
	memstoreACLDeleteCmd = &cobra.Command{
		Use: "delete ACL_POLICY", Short: "Delete a Memorystore ACL policy",
		Args: cobra.ExactArgs(1), RunE: runMemstoreACLDelete,
	}
	memstoreACLDescribeCmd = &cobra.Command{
		Use: "describe ACL_POLICY", Short: "Describe a Memorystore ACL policy",
		Args: cobra.ExactArgs(1), RunE: runMemstoreACLDescribe,
	}
	memstoreACLListCmd = &cobra.Command{
		Use: "list", Short: "List Memorystore ACL policies",
		Args: cobra.NoArgs, RunE: runMemstoreACLList,
	}
	memstoreACLUpdateCmd = &cobra.Command{
		Use: "update ACL_POLICY", Short: "Update a Memorystore ACL policy",
		Args: cobra.ExactArgs(1), RunE: runMemstoreACLUpdate,
	}
)

func init() {
	all := []*cobra.Command{memstoreACLCreateCmd, memstoreACLDeleteCmd, memstoreACLDescribeCmd, memstoreACLListCmd, memstoreACLUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagMemstoreACLLocation, "location", "", "Memorystore location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagMemstoreACLFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{memstoreACLCreateCmd, memstoreACLUpdateCmd} {
		c.Flags().StringVar(&flagMemstoreACLConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the ACL policy body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	memstoreACLUpdateCmd.Flags().StringVar(&flagMemstoreACLUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	memstoreACLListCmd.Flags().Int64Var(&flagMemstoreACLPageSize, "page-size", 0, "Maximum results per page")

	memstoreACLCmd.AddCommand(all...)
	memorystoreCmd.AddCommand(memstoreACLCmd)
}

func memstoreACLParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagMemstoreACLLocation), nil
}

func memstoreACLName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := memstoreACLParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/aclPolicies/%s", parent, id), nil
}

func runMemstoreACLCreate(cmd *cobra.Command, args []string) error {
	parent, err := memstoreACLParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagMemstoreACLConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("aclPolicyId", args[0])
	ctx := context.Background()
	var op map[string]any
	if err := memorystoreRest.do(ctx, http.MethodPost, "/"+parent+"/aclPolicies", q, body, &op); err != nil {
		return fmt.Errorf("creating ACL policy: %w", err)
	}
	fmt.Printf("Create request issued for ACL policy [%s].\n", args[0])
	return emitFormatted(op, flagMemstoreACLFormat)
}

func runMemstoreACLDelete(cmd *cobra.Command, args []string) error {
	name, err := memstoreACLName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var op map[string]any
	if err := memorystoreRest.do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting ACL policy: %w", err)
	}
	fmt.Printf("Delete request issued for ACL policy [%s].\n", args[0])
	return emitFormatted(op, flagMemstoreACLFormat)
}

func runMemstoreACLDescribe(cmd *cobra.Command, args []string) error {
	name, err := memstoreACLName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	var got map[string]any
	if err := memorystoreRest.do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing ACL policy: %w", err)
	}
	return emitFormatted(got, flagMemstoreACLFormat)
}

func runMemstoreACLList(cmd *cobra.Command, args []string) error {
	parent, err := memstoreACLParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	items, err := memorystoreRest.paginate(ctx, "/"+parent+"/aclPolicies", nil, "aclPolicies", flagMemstoreACLPageSize)
	if err != nil {
		return fmt.Errorf("listing ACL policies: %w", err)
	}
	return emitFormatted(items, flagMemstoreACLFormat)
}

func runMemstoreACLUpdate(cmd *cobra.Command, args []string) error {
	name, err := memstoreACLName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(flagMemstoreACLConfigFile, &body); err != nil {
		return err
	}
	mask := flagMemstoreACLUpdateMask
	if mask == "" {
		mask = joinMask(dcTopLevelKeys(body))
	}
	q := url.Values{}
	if mask != "" {
		q.Set("updateMask", mask)
	}
	ctx := context.Background()
	var op map[string]any
	if err := memorystoreRest.do(ctx, http.MethodPatch, "/"+name, q, body, &op); err != nil {
		return fmt.Errorf("updating ACL policy: %w", err)
	}
	fmt.Printf("Update request issued for ACL policy [%s].\n", args[0])
	return emitFormatted(op, flagMemstoreACLFormat)
}
