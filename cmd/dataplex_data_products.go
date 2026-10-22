package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	dataplex "google.golang.org/api/dataplex/v1"
)

// --- gcloud dataplex data-products (GA in gcloud-python 588.0.0, #1926) ---

var dataplexDataProductsCmd = &cobra.Command{Use: "data-products", Short: "Manage Dataplex data products"}
var dataplexDataAssetsCmd = &cobra.Command{Use: "data-assets", Short: "Manage Dataplex data assets"}

var (
	dpCreateCmd = &cobra.Command{Use: "create DATA_PRODUCT", Short: "Create a Dataplex data product",
		Args: cobra.ExactArgs(1), RunE: runDPCreate}
	dpUpdateCmd = &cobra.Command{Use: "update DATA_PRODUCT", Short: "Update a Dataplex data product",
		Args: cobra.ExactArgs(1), RunE: runDPUpdate}
	dpDeleteCmd = &cobra.Command{Use: "delete DATA_PRODUCT", Short: "Delete a Dataplex data product",
		Args: cobra.ExactArgs(1), RunE: runDPDelete}
	dpDescribeCmd = &cobra.Command{Use: "describe DATA_PRODUCT", Short: "Describe a Dataplex data product",
		Args: cobra.ExactArgs(1), RunE: runDPDescribe}
	dpListCmd = &cobra.Command{Use: "list", Short: "List Dataplex data products",
		Args: cobra.NoArgs, RunE: runDPList}
	dpGetIamCmd = &cobra.Command{Use: "get-iam-policy DATA_PRODUCT", Short: "Get the IAM policy of a Dataplex data product",
		Args: cobra.ExactArgs(1), RunE: runDPGetIam}
	dpSetIamCmd = &cobra.Command{Use: "set-iam-policy DATA_PRODUCT POLICY_FILE", Short: "Set the IAM policy of a Dataplex data product",
		Args: cobra.ExactArgs(2), RunE: runDPSetIam}
	dpAddIamCmd = &cobra.Command{Use: "add-iam-policy-binding DATA_PRODUCT", Short: "Add an IAM policy binding to a Dataplex data product",
		Args: cobra.ExactArgs(1), RunE: runDPAddIam}
	dpRemoveIamCmd = &cobra.Command{Use: "remove-iam-policy-binding DATA_PRODUCT", Short: "Remove an IAM policy binding from a Dataplex data product",
		Args: cobra.ExactArgs(1), RunE: runDPRemoveIam}

	daCreateCmd = &cobra.Command{Use: "create DATA_ASSET", Short: "Create a Dataplex data asset",
		Args: cobra.ExactArgs(1), RunE: runDACreate}
	daUpdateCmd = &cobra.Command{Use: "update DATA_ASSET", Short: "Update a Dataplex data asset",
		Args: cobra.ExactArgs(1), RunE: runDAUpdate}
	daDeleteCmd = &cobra.Command{Use: "delete DATA_ASSET", Short: "Delete a Dataplex data asset",
		Args: cobra.ExactArgs(1), RunE: runDADelete}
	daDescribeCmd = &cobra.Command{Use: "describe DATA_ASSET", Short: "Describe a Dataplex data asset",
		Args: cobra.ExactArgs(1), RunE: runDADescribe}
	daListCmd = &cobra.Command{Use: "list", Short: "List Dataplex data assets in a data product",
		Args: cobra.NoArgs, RunE: runDAList}
)

var (
	flagDPLocation     string
	flagDPDataProduct  string
	flagDPFormat       string
	flagDPAsync        bool
	flagDPDisplayName  string
	flagDPDescription  string
	flagDPOwnerEmails  []string
	flagDPAccessGroups []string
	flagDPLabels       map[string]string
	flagDPUpdateLabels map[string]string
	flagDPRemoveLabels []string
	flagDPClearLabels  bool
	flagDPEtag         string
	flagDPResource     string
	flagDPMember       string
	flagDPRole         string
	flagDPLimit        int
	flagDPPageSize     int64
	flagDPURI          bool
)

func init() {
	dpAll := []*cobra.Command{dpCreateCmd, dpUpdateCmd, dpDeleteCmd, dpDescribeCmd, dpListCmd,
		dpGetIamCmd, dpSetIamCmd, dpAddIamCmd, dpRemoveIamCmd}
	daAll := []*cobra.Command{daCreateCmd, daUpdateCmd, daDeleteCmd, daDescribeCmd, daListCmd}
	for _, c := range append(append([]*cobra.Command{}, dpAll...), daAll...) {
		c.Flags().StringVar(&flagDPLocation, "location", "", "Dataplex location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagDPFormat, "format", "", "Output format")
	}
	for _, c := range daAll {
		c.Flags().StringVar(&flagDPDataProduct, "data-product", "", "Data product containing the data asset (required)")
		_ = c.MarkFlagRequired("data-product")
	}
	for _, c := range []*cobra.Command{dpCreateCmd, dpUpdateCmd, dpDeleteCmd, daCreateCmd, daUpdateCmd, daDeleteCmd} {
		c.Flags().BoolVar(&flagDPAsync, "async", false,
			"Return immediately, without waiting for the operation in progress to complete")
	}
	for _, c := range []*cobra.Command{dpCreateCmd, dpUpdateCmd} {
		c.Flags().StringVar(&flagDPDisplayName, "display-name", "", "Display name of the data product")
		c.Flags().StringVar(&flagDPDescription, "description", "", "Description of the data product")
		c.Flags().StringSliceVar(&flagDPOwnerEmails, "owner-emails", nil, "Emails of the data product owners")
		c.Flags().StringArrayVar(&flagDPAccessGroups, "access-groups", nil,
			`Access group as a JSON string, e.g. {"id": "g", "display_name": "G", "principal": {"google_group": "g@example.com"}}; repeatable`)
	}
	_ = dpCreateCmd.MarkFlagRequired("display-name")
	_ = dpCreateCmd.MarkFlagRequired("owner-emails")
	daCreateCmd.Flags().StringVar(&flagDPResource, "resource", "",
		"Full resource name of the cloud resource represented by the data asset (required)")
	_ = daCreateCmd.MarkFlagRequired("resource")
	for _, c := range []*cobra.Command{daCreateCmd, daUpdateCmd} {
		c.Flags().StringArrayVar(&flagDPAccessGroups, "access-groups", nil,
			`Access group config as a JSON string, e.g. {"id": "g", "iam_roles": ["roles/x"]}; repeatable`)
	}
	for _, c := range []*cobra.Command{dpCreateCmd, daCreateCmd} {
		c.Flags().StringToStringVar(&flagDPLabels, "labels", nil, "List of label KEY=VALUE pairs to add")
	}
	for _, c := range []*cobra.Command{dpUpdateCmd, daUpdateCmd} {
		c.Flags().StringToStringVar(&flagDPUpdateLabels, "update-labels", nil, "List of label KEY=VALUE pairs to update")
		c.Flags().StringSliceVar(&flagDPRemoveLabels, "remove-labels", nil, "List of label keys to remove")
		c.Flags().BoolVar(&flagDPClearLabels, "clear-labels", false, "Remove all labels")
		c.MarkFlagsMutuallyExclusive("clear-labels", "remove-labels")
	}
	for _, c := range []*cobra.Command{dpDeleteCmd, daDeleteCmd} {
		c.Flags().StringVar(&flagDPEtag, "etag", "", "Etag of the resource; if given, it must match the server's etag")
	}
	for _, c := range []*cobra.Command{dpAddIamCmd, dpRemoveIamCmd} {
		c.Flags().StringVar(&flagDPMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&flagDPRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	for _, c := range []*cobra.Command{dpListCmd, daListCmd} {
		c.Flags().IntVar(&flagDPLimit, "limit", 0, "Maximum number of resources to list (0 = no limit)")
		c.Flags().Int64Var(&flagDPPageSize, "page-size", 0, "Number of results per API page")
	}
	daListCmd.Flags().BoolVar(&flagDPURI, "uri", false, "Print resource names")

	dataplexDataAssetsCmd.AddCommand(daAll...)
	dataplexDataProductsCmd.AddCommand(dpAll...)
	dataplexDataProductsCmd.AddCommand(dataplexDataAssetsCmd)
	dataplexCmd.AddCommand(dataplexDataProductsCmd)
}

// --- shared helpers ---

func dpParent() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, flagDPLocation), nil
}

// dpName expands a short id under parent/collection; full names pass through.
func dpName(parent, collection, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return parent + "/" + collection + "/" + id
}

func dpProductName(id string) (string, error) {
	parent, err := dpParent()
	if err != nil {
		return "", err
	}
	return dpName(parent, "dataProducts", id), nil
}

func dpAssetParent() (string, error) {
	return dpProductName(flagDPDataProduct)
}

// snakeToCamelJSON rewrites object keys from snake_case to camelCase,
// recursively, as gcloud-python's dataplex_util.SnakeToCamelDict does.
func snakeToCamelJSON(v any) any {
	switch t := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, val := range t {
			out[snakeToCamel(k)] = snakeToCamelJSON(val)
		}
		return out
	case []any:
		for i := range t {
			t[i] = snakeToCamelJSON(t[i])
		}
		return t
	}
	return v
}

func snakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if parts[i] != "" {
			r := []rune(parts[i])
			r[0] = unicode.ToUpper(r[0])
			parts[i] = string(r)
		}
	}
	return strings.Join(parts, "")
}

// decodeStrictJSON decodes a snake- or camel-case JSON object into out,
// rejecting unknown fields like gcloud's DictToMessageWithErrorCheck.
func decodeStrictJSON(obj map[string]any, out any) error {
	raw, err := json.Marshal(snakeToCamelJSON(obj))
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	return dec.Decode(out)
}

// parseDPAccessGroups mirrors parsers.ParseAccessGroups: each JSON string is
// an access group keyed by its required "id".
func parseDPAccessGroups(specs []string) (map[string]dataplex.GoogleCloudDataplexV1DataProductAccessGroup, error) {
	out := map[string]dataplex.GoogleCloudDataplexV1DataProductAccessGroup{}
	for _, spec := range specs {
		var obj map[string]any
		if err := json.Unmarshal([]byte(spec), &obj); err != nil {
			return nil, fmt.Errorf("argument --access-groups: %v", err)
		}
		var g dataplex.GoogleCloudDataplexV1DataProductAccessGroup
		if err := decodeStrictJSON(obj, &g); err != nil {
			return nil, fmt.Errorf("argument --access-groups: %v", err)
		}
		if g.Id == "" {
			return nil, fmt.Errorf(`argument --access-groups: Access group must have an "id" field.`)
		}
		out[g.Id] = g
	}
	return out, nil
}

// parseDAAccessGroupConfigs mirrors parsers.ParseAccessGroupConfigs: "id"
// is popped as the map key and the rest decodes as an AccessGroupConfig.
func parseDAAccessGroupConfigs(specs []string) (map[string]dataplex.GoogleCloudDataplexV1DataAssetAccessGroupConfig, error) {
	out := map[string]dataplex.GoogleCloudDataplexV1DataAssetAccessGroupConfig{}
	for _, spec := range specs {
		var obj map[string]any
		if err := json.Unmarshal([]byte(spec), &obj); err != nil {
			return nil, fmt.Errorf("argument --access-groups: %v", err)
		}
		id, _ := obj["id"].(string)
		delete(obj, "id")
		if id == "" {
			return nil, fmt.Errorf(`argument --access-groups: Access group config must have an "id" field.`)
		}
		var c dataplex.GoogleCloudDataplexV1DataAssetAccessGroupConfig
		if err := decodeStrictJSON(obj, &c); err != nil {
			return nil, fmt.Errorf("argument --access-groups: %v", err)
		}
		out[id] = c
	}
	return out, nil
}

// dpFinishOp prints gcloud's declarative LRO messages and waits unless
// --async. kind is e.g. "dataProduct"; verb is Create/Update/Delete.
func dpFinishOp(ctx context.Context, svc *dataplex.Service, op *dataplex.GoogleLongrunningOperation, verb, kind, id string) (bool, error) {
	fmt.Fprintf(os.Stderr, "%s request issued for: [%s]\n", verb, id)
	if flagDPAsync {
		fmt.Fprintf(os.Stderr, "Check operation [%s] for status.\n", op.Name)
		return false, emitFormatted(op, flagDPFormat)
	}
	fmt.Fprintf(os.Stderr, "Waiting for operation [%s] to complete...", op.Name)
	for !op.Done {
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			fmt.Fprintln(os.Stderr, "failed.")
			return false, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		fmt.Fprintln(os.Stderr, "failed.")
		return false, fmt.Errorf("%s", op.Error.Message)
	}
	fmt.Fprintln(os.Stderr, "done.")
	past := map[string]string{"Create": "Created", "Update": "Updated", "Delete": "Deleted"}[verb]
	fmt.Fprintf(os.Stderr, "%s %s [%s].\n", past, kind, id)
	return true, nil
}

func dpService(ctx context.Context) (*dataplex.Service, error) {
	return gcp.DataplexService(ctx, flagAccount)
}

// --- data-products ---

func runDPCreate(cmd *cobra.Command, args []string) error {
	parent, err := dpParent()
	if err != nil {
		return err
	}
	groups, err := parseDPAccessGroups(flagDPAccessGroups)
	if err != nil {
		return err
	}
	body := &dataplex.GoogleCloudDataplexV1DataProduct{
		DisplayName: flagDPDisplayName,
		Description: flagDPDescription,
		OwnerEmails: flagDPOwnerEmails,
		Labels:      flagDPLabels,
	}
	if len(groups) > 0 {
		body.AccessGroups = groups
	}
	ctx := context.Background()
	svc, err := dpService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DataProducts.Create(parent, body).DataProductId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating data product: %w", err)
	}
	done, err := dpFinishOp(ctx, svc, op, "Create", "dataProduct", args[0])
	if err != nil || !done {
		return err
	}
	return dpEmitIfFormat(ctx, func() (any, error) {
		return svc.Projects.Locations.DataProducts.Get(dpName(parent, "dataProducts", args[0])).Context(ctx).Do()
	})
}

// dpEmitIfFormat prints the finished resource when --format is given.
func dpEmitIfFormat(ctx context.Context, get func() (any, error)) error {
	if flagDPFormat == "" {
		return nil
	}
	v, err := get()
	if err != nil {
		return err
	}
	return emitFormatted(v, flagDPFormat)
}

func runDPUpdate(cmd *cobra.Command, args []string) error {
	name, err := dpProductName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := dpService(ctx)
	if err != nil {
		return err
	}
	body := &dataplex.GoogleCloudDataplexV1DataProduct{}
	var mask []string
	if cmd.Flags().Changed("display-name") {
		body.DisplayName = flagDPDisplayName
		mask = append(mask, "displayName")
	}
	if cmd.Flags().Changed("description") {
		body.Description = flagDPDescription
		mask = append(mask, "description")
	}
	if cmd.Flags().Changed("owner-emails") {
		body.OwnerEmails = flagDPOwnerEmails
		mask = append(mask, "ownerEmails")
	}
	if cmd.Flags().Changed("access-groups") {
		groups, err := parseDPAccessGroups(flagDPAccessGroups)
		if err != nil {
			return err
		}
		body.AccessGroups = groups
		mask = append(mask, "accessGroups")
	}
	if dpLabelsRequested() {
		cur, err := svc.Projects.Locations.DataProducts.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("describing data product: %w", err)
		}
		if labels, changed := applyLabelsDiff(cur.Labels, flagDPClearLabels, flagDPRemoveLabels, flagDPUpdateLabels); changed {
			body.Labels = labels
			if len(labels) == 0 {
				body.NullFields = append(body.NullFields, "Labels")
			}
			mask = append(mask, "labels")
		}
	}
	if len(mask) == 0 {
		return fmt.Errorf("No fields to update. Specify at least one flag to update.")
	}
	op, err := svc.Projects.Locations.DataProducts.Patch(name, body).UpdateMask(strings.Join(mask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating data product: %w", err)
	}
	done, err := dpFinishOp(ctx, svc, op, "Update", "dataProduct", args[0])
	if err != nil || !done {
		return err
	}
	return dpEmitIfFormat(ctx, func() (any, error) {
		return svc.Projects.Locations.DataProducts.Get(name).Context(ctx).Do()
	})
}

func dpLabelsRequested() bool {
	return flagDPClearLabels || len(flagDPRemoveLabels) > 0 || len(flagDPUpdateLabels) > 0
}

func runDPDelete(cmd *cobra.Command, args []string) error {
	name, err := dpProductName(args[0])
	if err != nil {
		return err
	}
	if err := promptContinue(fmt.Sprintf("You are about to delete data product [%s]", args[0]), ""); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := dpService(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.DataProducts.Delete(name).Context(ctx)
	if flagDPEtag != "" {
		call = call.Etag(flagDPEtag)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting data product: %w", err)
	}
	_, err = dpFinishOp(ctx, svc, op, "Delete", "dataProduct", args[0])
	return err
}

func runDPDescribe(cmd *cobra.Command, args []string) error {
	name, err := dpProductName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := dpService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DataProducts.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing data product: %w", err)
	}
	return emitFormatted(got, flagDPFormat)
}

func runDPList(cmd *cobra.Command, args []string) error {
	parent, err := dpParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := dpService(ctx)
	if err != nil {
		return err
	}
	var all []*dataplex.GoogleCloudDataplexV1DataProduct
	call := svc.Projects.Locations.DataProducts.List(parent).Context(ctx)
	if flagDPPageSize > 0 {
		call = call.PageSize(flagDPPageSize)
	}
	err = call.Pages(ctx, func(r *dataplex.GoogleCloudDataplexV1ListDataProductsResponse) error {
		all = append(all, r.DataProducts...)
		return nil
	})
	if err != nil {
		return fmt.Errorf("listing data products: %w", err)
	}
	if flagDPLimit > 0 && len(all) > flagDPLimit {
		all = all[:flagDPLimit]
	}
	return dpEmitList(all, len(all))
}

// dpEmitList prints a list as YAML records (the declarative default) or per
// --format, with "Listed 0 items." on stderr when empty.
func dpEmitList(v any, n int) error {
	if n == 0 && flagDPFormat == "" {
		fmt.Fprintln(os.Stderr, "Listed 0 items.")
		return nil
	}
	return emitFormatted(v, flagDPFormat)
}

// --- data-products IAM ---

func runDPGetIam(cmd *cobra.Command, args []string) error {
	name, err := dpProductName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := dpService(ctx)
	if err != nil {
		return err
	}
	pol, err := svc.Projects.Locations.DataProducts.GetIamPolicy(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(pol, flagDPFormat)
}

func runDPSetIam(cmd *cobra.Command, args []string) error {
	name, err := dpProductName(args[0])
	if err != nil {
		return err
	}
	pol := &dataplex.GoogleIamV1Policy{}
	if err := loadYAMLOrJSONInto(args[1], pol); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := dpService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DataProducts.SetIamPolicy(name, &dataplex.GoogleIamV1SetIamPolicyRequest{Policy: pol}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updated IAM policy for data product [%s].\n", args[0])
	return emitFormatted(got, flagDPFormat)
}

// dpxAddBinding / dpxRemoveBinding mirror iam_util Add/RemoveBindingFromIamPolicy
// for unconditional bindings.
func dpxAddBinding(pol *dataplex.GoogleIamV1Policy, role, member string) {
	for _, b := range pol.Bindings {
		if b.Role == role && b.Condition == nil {
			for _, m := range b.Members {
				if m == member {
					return
				}
			}
			b.Members = append(b.Members, member)
			return
		}
	}
	pol.Bindings = append(pol.Bindings, &dataplex.GoogleIamV1Binding{Role: role, Members: []string{member}})
}

func dpxRemoveBinding(pol *dataplex.GoogleIamV1Policy, role, member string) error {
	for i, b := range pol.Bindings {
		if b.Role != role || b.Condition != nil {
			continue
		}
		for j, m := range b.Members {
			if m == member {
				b.Members = append(b.Members[:j], b.Members[j+1:]...)
				if len(b.Members) == 0 {
					pol.Bindings = append(pol.Bindings[:i], pol.Bindings[i+1:]...)
				}
				return nil
			}
		}
	}
	return fmt.Errorf("Policy binding with the specified principal, role, and condition not found!")
}

func dpModifyIam(args []string, modify func(*dataplex.GoogleIamV1Policy) error) error {
	name, err := dpProductName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := dpService(ctx)
	if err != nil {
		return err
	}
	pol, err := svc.Projects.Locations.DataProducts.GetIamPolicy(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if err := modify(pol); err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DataProducts.SetIamPolicy(name, &dataplex.GoogleIamV1SetIamPolicyRequest{Policy: pol}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, flagDPFormat)
}

func runDPAddIam(cmd *cobra.Command, args []string) error {
	return dpModifyIam(args, func(p *dataplex.GoogleIamV1Policy) error {
		dpxAddBinding(p, flagDPRole, flagDPMember)
		return nil
	})
}

func runDPRemoveIam(cmd *cobra.Command, args []string) error {
	return dpModifyIam(args, func(p *dataplex.GoogleIamV1Policy) error {
		return dpxRemoveBinding(p, flagDPRole, flagDPMember)
	})
}

// --- data-assets ---

func runDACreate(cmd *cobra.Command, args []string) error {
	parent, err := dpAssetParent()
	if err != nil {
		return err
	}
	configs, err := parseDAAccessGroupConfigs(flagDPAccessGroups)
	if err != nil {
		return err
	}
	body := &dataplex.GoogleCloudDataplexV1DataAsset{Resource: flagDPResource, Labels: flagDPLabels}
	if len(configs) > 0 {
		body.AccessGroupConfigs = configs
	}
	ctx := context.Background()
	svc, err := dpService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DataProducts.DataAssets.Create(parent, body).DataAssetId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating data asset: %w", err)
	}
	done, err := dpFinishOp(ctx, svc, op, "Create", "dataAsset", args[0])
	if err != nil || !done {
		return err
	}
	return dpEmitIfFormat(ctx, func() (any, error) {
		return svc.Projects.Locations.DataProducts.DataAssets.Get(dpName(parent, "dataAssets", args[0])).Context(ctx).Do()
	})
}

func runDAUpdate(cmd *cobra.Command, args []string) error {
	parent, err := dpAssetParent()
	if err != nil {
		return err
	}
	name := dpName(parent, "dataAssets", args[0])
	ctx := context.Background()
	svc, err := dpService(ctx)
	if err != nil {
		return err
	}
	body := &dataplex.GoogleCloudDataplexV1DataAsset{}
	var mask []string
	if cmd.Flags().Changed("access-groups") {
		configs, err := parseDAAccessGroupConfigs(flagDPAccessGroups)
		if err != nil {
			return err
		}
		body.AccessGroupConfigs = configs
		mask = append(mask, "accessGroupConfigs")
	}
	if dpLabelsRequested() {
		cur, err := svc.Projects.Locations.DataProducts.DataAssets.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("describing data asset: %w", err)
		}
		if labels, changed := applyLabelsDiff(cur.Labels, flagDPClearLabels, flagDPRemoveLabels, flagDPUpdateLabels); changed {
			body.Labels = labels
			if len(labels) == 0 {
				body.NullFields = append(body.NullFields, "Labels")
			}
			mask = append(mask, "labels")
		}
	}
	if len(mask) == 0 {
		return fmt.Errorf("No fields to update. Specify at least one flag to update.")
	}
	op, err := svc.Projects.Locations.DataProducts.DataAssets.Patch(name, body).UpdateMask(strings.Join(mask, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating data asset: %w", err)
	}
	done, err := dpFinishOp(ctx, svc, op, "Update", "dataAsset", args[0])
	if err != nil || !done {
		return err
	}
	return dpEmitIfFormat(ctx, func() (any, error) {
		return svc.Projects.Locations.DataProducts.DataAssets.Get(name).Context(ctx).Do()
	})
}

func runDADelete(cmd *cobra.Command, args []string) error {
	parent, err := dpAssetParent()
	if err != nil {
		return err
	}
	if err := promptContinue(fmt.Sprintf("You are about to delete data asset [%s]", args[0]), ""); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := dpService(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.DataProducts.DataAssets.Delete(dpName(parent, "dataAssets", args[0])).Context(ctx)
	if flagDPEtag != "" {
		call = call.Etag(flagDPEtag)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting data asset: %w", err)
	}
	_, err = dpFinishOp(ctx, svc, op, "Delete", "dataAsset", args[0])
	return err
}

func runDADescribe(cmd *cobra.Command, args []string) error {
	parent, err := dpAssetParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := dpService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DataProducts.DataAssets.Get(dpName(parent, "dataAssets", args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing data asset: %w", err)
	}
	return emitFormatted(got, flagDPFormat)
}

func runDAList(cmd *cobra.Command, args []string) error {
	parent, err := dpAssetParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := dpService(ctx)
	if err != nil {
		return err
	}
	var all []*dataplex.GoogleCloudDataplexV1DataAsset
	call := svc.Projects.Locations.DataProducts.DataAssets.List(parent).Context(ctx)
	if flagDPPageSize > 0 {
		call = call.PageSize(flagDPPageSize)
	}
	err = call.Pages(ctx, func(r *dataplex.GoogleCloudDataplexV1ListDataAssetsResponse) error {
		all = append(all, r.DataAssets...)
		return nil
	})
	if err != nil {
		return fmt.Errorf("listing data assets: %w", err)
	}
	if flagDPLimit > 0 && len(all) > flagDPLimit {
		all = all[:flagDPLimit]
	}
	if flagDPURI {
		for _, a := range all {
			fmt.Println(a.Name)
		}
		return nil
	}
	return dpEmitList(all, len(all))
}
