package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	dataplex "google.golang.org/api/dataplex/v1"
)

// --- gcloud dataplex data-products (GA in gcloud-python 588.0.0, #1926) ---

func (g *dataplexCmds) initDataplexDataProducts() {
	dpAll := []*cobra.Command{g.dpCreateCmd, g.dpUpdateCmd, g.dpDeleteCmd, g.dpDescribeCmd, g.dpListCmd,
		g.dpGetIamCmd, g.dpSetIamCmd, g.dpAddIamCmd, g.dpRemoveIamCmd}
	daAll := []*cobra.Command{g.daCreateCmd, g.daUpdateCmd, g.daDeleteCmd, g.daDescribeCmd, g.daListCmd}
	for _, c := range append(append([]*cobra.Command{}, dpAll...), daAll...) {
		c.Flags().StringVar(&g.flagDPLocation, "location", "", "Dataplex location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&g.flagDPFormat, "format", "", "Output format")
	}
	for _, c := range daAll {
		c.Flags().StringVar(&g.flagDPDataProduct, "data-product", "", "Data product containing the data asset (required)")
		_ = c.MarkFlagRequired("data-product")
	}
	for _, c := range []*cobra.Command{g.dpCreateCmd, g.dpUpdateCmd, g.dpDeleteCmd, g.daCreateCmd, g.daUpdateCmd, g.daDeleteCmd} {
		c.Flags().BoolVar(&g.flagDPAsync, "async", false,
			"Return immediately, without waiting for the operation in progress to complete")
	}
	for _, c := range []*cobra.Command{g.dpCreateCmd, g.dpUpdateCmd} {
		c.Flags().StringVar(&g.flagDPDisplayName, "display-name", "", "Display name of the data product")
		c.Flags().StringVar(&g.flagDPDescription, "description", "", "Description of the data product")
		c.Flags().StringSliceVar(&g.flagDPOwnerEmails, "owner-emails", nil, "Emails of the data product owners")
		c.Flags().StringArrayVar(&g.flagDPAccessGroups, "access-groups", nil,
			`Access group as a JSON string, e.g. {"id": "g", "display_name": "G", "principal": {"google_group": "g@example.com"}}; repeatable`)
	}
	_ = g.dpCreateCmd.MarkFlagRequired("display-name")
	_ = g.dpCreateCmd.MarkFlagRequired("owner-emails")
	g.daCreateCmd.Flags().StringVar(&g.flagDPResource, "resource", "",
		"Full resource name of the cloud resource represented by the data asset (required)")
	_ = g.daCreateCmd.MarkFlagRequired("resource")
	for _, c := range []*cobra.Command{g.daCreateCmd, g.daUpdateCmd} {
		c.Flags().StringArrayVar(&g.flagDPAccessGroups, "access-groups", nil,
			`Access group config as a JSON string, e.g. {"id": "g", "iam_roles": ["roles/x"]}; repeatable`)
	}
	for _, c := range []*cobra.Command{g.dpCreateCmd, g.daCreateCmd} {
		c.Flags().StringToStringVar(&g.flagDPLabels, "labels", nil, "List of label KEY=VALUE pairs to add")
	}
	for _, c := range []*cobra.Command{g.dpUpdateCmd, g.daUpdateCmd} {
		c.Flags().StringToStringVar(&g.flagDPUpdateLabels, "update-labels", nil, "List of label KEY=VALUE pairs to update")
		c.Flags().StringSliceVar(&g.flagDPRemoveLabels, "remove-labels", nil, "List of label keys to remove")
		c.Flags().BoolVar(&g.flagDPClearLabels, "clear-labels", false, "Remove all labels")
		c.MarkFlagsMutuallyExclusive("clear-labels", "remove-labels")
	}
	for _, c := range []*cobra.Command{g.dpDeleteCmd, g.daDeleteCmd} {
		c.Flags().StringVar(&g.flagDPEtag, "etag", "", "Etag of the resource; if given, it must match the server's etag")
	}
	for _, c := range []*cobra.Command{g.dpAddIamCmd, g.dpRemoveIamCmd} {
		c.Flags().StringVar(&g.flagDPMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&g.flagDPRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	for _, c := range []*cobra.Command{g.dpListCmd, g.daListCmd} {
		c.Flags().IntVar(&g.flagDPLimit, "limit", 0, "Maximum number of resources to list (0 = no limit)")
		c.Flags().Int64Var(&g.flagDPPageSize, "page-size", 0, "Number of results per API page")
	}
	g.daListCmd.Flags().BoolVar(&g.flagDPURI, "uri", false, "Print resource names")

	g.dataplexDataAssetsCmd.AddCommand(daAll...)
	g.dataplexDataProductsCmd.AddCommand(dpAll...)
	g.dataplexDataProductsCmd.AddCommand(g.dataplexDataAssetsCmd)
	g.dataplexCmd.AddCommand(g.dataplexDataProductsCmd)
}

// --- shared helpers ---

func (g *dataplexCmds) dpParent() (string, error) {
	project, err := g.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, g.flagDPLocation), nil
}

// dpName expands a short id under parent/collection; full names pass through.
func dpName(parent, collection, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return parent + "/" + collection + "/" + id
}

func (g *dataplexCmds) dpProductName(id string) (string, error) {
	parent, err := g.dpParent()
	if err != nil {
		return "", err
	}
	return dpName(parent, "dataProducts", id), nil
}

func (g *dataplexCmds) dpAssetParent() (string, error) {
	return g.dpProductName(g.flagDPDataProduct)
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
func (g *dataplexCmds) dpFinishOp(ctx context.Context, svc *dataplex.Service, op *dataplex.GoogleLongrunningOperation, verb, kind, id string) (bool, error) {
	fmt.Fprintf(g.inv.stderr, "%s request issued for: [%s]\n", verb, id)
	if g.flagDPAsync {
		fmt.Fprintf(g.inv.stderr, "Check operation [%s] for status.\n", op.Name)
		return false, emitFormattedTo(g.inv.stdout, op, g.flagDPFormat)
	}
	fmt.Fprintf(g.inv.stderr, "Waiting for operation [%s] to complete...", op.Name)
	for !op.Done {
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			fmt.Fprintln(g.inv.stderr, "failed.")
			return false, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		fmt.Fprintln(g.inv.stderr, "failed.")
		return false, fmt.Errorf("%s", op.Error.Message)
	}
	fmt.Fprintln(g.inv.stderr, "done.")
	past := map[string]string{"Create": "Created", "Update": "Updated", "Delete": "Deleted"}[verb]
	fmt.Fprintf(g.inv.stderr, "%s %s [%s].\n", past, kind, id)
	return true, nil
}

func (g *dataplexCmds) dpService(ctx context.Context) (*dataplex.Service, error) {
	return g.inv.clients().Dataplex(ctx)
}

// --- data-products ---

func (g *dataplexCmds) runDPCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.dpParent()
	if err != nil {
		return err
	}
	groups, err := parseDPAccessGroups(g.flagDPAccessGroups)
	if err != nil {
		return err
	}
	body := &dataplex.GoogleCloudDataplexV1DataProduct{
		DisplayName: g.flagDPDisplayName,
		Description: g.flagDPDescription,
		OwnerEmails: g.flagDPOwnerEmails,
		Labels:      g.flagDPLabels,
	}
	if len(groups) > 0 {
		body.AccessGroups = groups
	}
	ctx := cmd.Context()
	svc, err := g.dpService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DataProducts.Create(parent, body).DataProductId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating data product: %w", err)
	}
	done, err := g.dpFinishOp(ctx, svc, op, "Create", "dataProduct", args[0])
	if err != nil || !done {
		return err
	}
	return g.dpEmitIfFormat(ctx, func() (any, error) {
		return svc.Projects.Locations.DataProducts.Get(dpName(parent, "dataProducts", args[0])).Context(ctx).Do()
	})
}

// dpEmitIfFormat prints the finished resource when --format is given.
func (g *dataplexCmds) dpEmitIfFormat(ctx context.Context, get func() (any, error)) error {
	if g.flagDPFormat == "" {
		return nil
	}
	v, err := get()
	if err != nil {
		return err
	}
	return emitFormattedTo(g.inv.stdout, v, g.flagDPFormat)
}

func (g *dataplexCmds) runDPUpdate(cmd *cobra.Command, args []string) error {
	name, err := g.dpProductName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.dpService(ctx)
	if err != nil {
		return err
	}
	body := &dataplex.GoogleCloudDataplexV1DataProduct{}
	var mask []string
	if cmd.Flags().Changed("display-name") {
		body.DisplayName = g.flagDPDisplayName
		mask = append(mask, "displayName")
	}
	if cmd.Flags().Changed("description") {
		body.Description = g.flagDPDescription
		mask = append(mask, "description")
	}
	if cmd.Flags().Changed("owner-emails") {
		body.OwnerEmails = g.flagDPOwnerEmails
		mask = append(mask, "ownerEmails")
	}
	if cmd.Flags().Changed("access-groups") {
		groups, err := parseDPAccessGroups(g.flagDPAccessGroups)
		if err != nil {
			return err
		}
		body.AccessGroups = groups
		mask = append(mask, "accessGroups")
	}
	if g.dpLabelsRequested() {
		cur, err := svc.Projects.Locations.DataProducts.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("describing data product: %w", err)
		}
		if labels, changed := applyLabelsDiff(cur.Labels, g.flagDPClearLabels, g.flagDPRemoveLabels, g.flagDPUpdateLabels); changed {
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
	done, err := g.dpFinishOp(ctx, svc, op, "Update", "dataProduct", args[0])
	if err != nil || !done {
		return err
	}
	return g.dpEmitIfFormat(ctx, func() (any, error) {
		return svc.Projects.Locations.DataProducts.Get(name).Context(ctx).Do()
	})
}

func (g *dataplexCmds) dpLabelsRequested() bool {
	return g.flagDPClearLabels || len(g.flagDPRemoveLabels) > 0 || len(g.flagDPUpdateLabels) > 0
}

func (g *dataplexCmds) runDPDelete(cmd *cobra.Command, args []string) error {
	name, err := g.dpProductName(args[0])
	if err != nil {
		return err
	}
	if err := g.inv.promptContinue(fmt.Sprintf("You are about to delete data product [%s]", args[0]), ""); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.dpService(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.DataProducts.Delete(name).Context(ctx)
	if g.flagDPEtag != "" {
		call = call.Etag(g.flagDPEtag)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting data product: %w", err)
	}
	_, err = g.dpFinishOp(ctx, svc, op, "Delete", "dataProduct", args[0])
	return err
}

func (g *dataplexCmds) runDPDescribe(cmd *cobra.Command, args []string) error {
	name, err := g.dpProductName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.dpService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DataProducts.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing data product: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagDPFormat)
}

func (g *dataplexCmds) runDPList(cmd *cobra.Command, args []string) error {
	parent, err := g.dpParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.dpService(ctx)
	if err != nil {
		return err
	}
	var all []*dataplex.GoogleCloudDataplexV1DataProduct
	call := svc.Projects.Locations.DataProducts.List(parent).Context(ctx)
	if g.flagDPPageSize > 0 {
		call = call.PageSize(g.flagDPPageSize)
	}
	err = call.Pages(ctx, func(r *dataplex.GoogleCloudDataplexV1ListDataProductsResponse) error {
		all = append(all, r.DataProducts...)
		return nil
	})
	if err != nil {
		return fmt.Errorf("listing data products: %w", err)
	}
	if g.flagDPLimit > 0 && len(all) > g.flagDPLimit {
		all = all[:g.flagDPLimit]
	}
	return g.dpEmitList(all, len(all))
}

// dpEmitList prints a list as YAML records (the declarative default) or per
// --format, with "Listed 0 items." on stderr when empty.
func (g *dataplexCmds) dpEmitList(v any, n int) error {
	if n == 0 && g.flagDPFormat == "" {
		fmt.Fprintln(g.inv.stderr, "Listed 0 items.")
		return nil
	}
	return emitFormattedTo(g.inv.stdout, v, g.flagDPFormat)
}

// --- data-products IAM ---

func (g *dataplexCmds) runDPGetIam(cmd *cobra.Command, args []string) error {
	name, err := g.dpProductName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.dpService(ctx)
	if err != nil {
		return err
	}
	pol, err := svc.Projects.Locations.DataProducts.GetIamPolicy(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, pol, g.flagDPFormat)
}

func (g *dataplexCmds) runDPSetIam(cmd *cobra.Command, args []string) error {
	name, err := g.dpProductName(args[0])
	if err != nil {
		return err
	}
	pol := &dataplex.GoogleIamV1Policy{}
	if err := loadYAMLOrJSONInto(args[1], pol); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.dpService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DataProducts.SetIamPolicy(name, &dataplex.GoogleIamV1SetIamPolicyRequest{Policy: pol}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	fmt.Fprintf(g.inv.stderr, "Updated IAM policy for data product [%s].\n", args[0])
	return emitFormattedTo(g.inv.stdout, got, g.flagDPFormat)
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

func (g *dataplexCmds) dpModifyIam(args []string, modify func(*dataplex.GoogleIamV1Policy) error) error {
	name, err := g.dpProductName(args[0])
	if err != nil {
		return err
	}
	ctx := g.inv.context()
	svc, err := g.dpService(ctx)
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
	return emitFormattedTo(g.inv.stdout, got, g.flagDPFormat)
}

func (g *dataplexCmds) runDPAddIam(cmd *cobra.Command, args []string) error {
	return g.dpModifyIam(args, func(p *dataplex.GoogleIamV1Policy) error {
		dpxAddBinding(p, g.flagDPRole, g.flagDPMember)
		return nil
	})
}

func (g *dataplexCmds) runDPRemoveIam(cmd *cobra.Command, args []string) error {
	return g.dpModifyIam(args, func(p *dataplex.GoogleIamV1Policy) error {
		return dpxRemoveBinding(p, g.flagDPRole, g.flagDPMember)
	})
}

// --- data-assets ---

func (g *dataplexCmds) runDACreate(cmd *cobra.Command, args []string) error {
	parent, err := g.dpAssetParent()
	if err != nil {
		return err
	}
	configs, err := parseDAAccessGroupConfigs(g.flagDPAccessGroups)
	if err != nil {
		return err
	}
	body := &dataplex.GoogleCloudDataplexV1DataAsset{Resource: g.flagDPResource, Labels: g.flagDPLabels}
	if len(configs) > 0 {
		body.AccessGroupConfigs = configs
	}
	ctx := cmd.Context()
	svc, err := g.dpService(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.DataProducts.DataAssets.Create(parent, body).DataAssetId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating data asset: %w", err)
	}
	done, err := g.dpFinishOp(ctx, svc, op, "Create", "dataAsset", args[0])
	if err != nil || !done {
		return err
	}
	return g.dpEmitIfFormat(ctx, func() (any, error) {
		return svc.Projects.Locations.DataProducts.DataAssets.Get(dpName(parent, "dataAssets", args[0])).Context(ctx).Do()
	})
}

func (g *dataplexCmds) runDAUpdate(cmd *cobra.Command, args []string) error {
	parent, err := g.dpAssetParent()
	if err != nil {
		return err
	}
	name := dpName(parent, "dataAssets", args[0])
	ctx := cmd.Context()
	svc, err := g.dpService(ctx)
	if err != nil {
		return err
	}
	body := &dataplex.GoogleCloudDataplexV1DataAsset{}
	var mask []string
	if cmd.Flags().Changed("access-groups") {
		configs, err := parseDAAccessGroupConfigs(g.flagDPAccessGroups)
		if err != nil {
			return err
		}
		body.AccessGroupConfigs = configs
		mask = append(mask, "accessGroupConfigs")
	}
	if g.dpLabelsRequested() {
		cur, err := svc.Projects.Locations.DataProducts.DataAssets.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("describing data asset: %w", err)
		}
		if labels, changed := applyLabelsDiff(cur.Labels, g.flagDPClearLabels, g.flagDPRemoveLabels, g.flagDPUpdateLabels); changed {
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
	done, err := g.dpFinishOp(ctx, svc, op, "Update", "dataAsset", args[0])
	if err != nil || !done {
		return err
	}
	return g.dpEmitIfFormat(ctx, func() (any, error) {
		return svc.Projects.Locations.DataProducts.DataAssets.Get(name).Context(ctx).Do()
	})
}

func (g *dataplexCmds) runDADelete(cmd *cobra.Command, args []string) error {
	parent, err := g.dpAssetParent()
	if err != nil {
		return err
	}
	if err := g.inv.promptContinue(fmt.Sprintf("You are about to delete data asset [%s]", args[0]), ""); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.dpService(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.DataProducts.DataAssets.Delete(dpName(parent, "dataAssets", args[0])).Context(ctx)
	if g.flagDPEtag != "" {
		call = call.Etag(g.flagDPEtag)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("deleting data asset: %w", err)
	}
	_, err = g.dpFinishOp(ctx, svc, op, "Delete", "dataAsset", args[0])
	return err
}

func (g *dataplexCmds) runDADescribe(cmd *cobra.Command, args []string) error {
	parent, err := g.dpAssetParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.dpService(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.DataProducts.DataAssets.Get(dpName(parent, "dataAssets", args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing data asset: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagDPFormat)
}

func (g *dataplexCmds) runDAList(cmd *cobra.Command, args []string) error {
	parent, err := g.dpAssetParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.dpService(ctx)
	if err != nil {
		return err
	}
	var all []*dataplex.GoogleCloudDataplexV1DataAsset
	call := svc.Projects.Locations.DataProducts.DataAssets.List(parent).Context(ctx)
	if g.flagDPPageSize > 0 {
		call = call.PageSize(g.flagDPPageSize)
	}
	err = call.Pages(ctx, func(r *dataplex.GoogleCloudDataplexV1ListDataAssetsResponse) error {
		all = append(all, r.DataAssets...)
		return nil
	})
	if err != nil {
		return fmt.Errorf("listing data assets: %w", err)
	}
	if g.flagDPLimit > 0 && len(all) > g.flagDPLimit {
		all = all[:g.flagDPLimit]
	}
	if g.flagDPURI {
		for _, a := range all {
			fmt.Fprintln(g.inv.stdout, a.Name)
		}
		return nil
	}
	return g.dpEmitList(all, len(all))
}
