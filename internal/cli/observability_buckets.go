package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	observability "google.golang.org/api/observability/v1"
)

// --- gcloud observability buckets (#1936) ---
//
// Declarative commands in gcloud-python on observability.projects.locations.
// buckets. create and update (added in 586.0.0) return an operation that is
// waited on unless --async.

func (g *observabilityCmds) initObservabilityBuckets() {
	all := []*cobra.Command{g.obBucketCreateCmd, g.obBucketDescribeCmd, g.obBucketListCmd, g.obBucketUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&g.flagObBucketLocation, "location", "", "The location id of the bucket resource")
		c.Flags().StringVar(&g.flagObBucketFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{g.obBucketCreateCmd, g.obBucketUpdateCmd} {
		c.Flags().BoolVar(&g.flagObBucketAsync, "async", false,
			"Return immediately, without waiting for the operation in progress to complete")
		c.Flags().StringVar(&g.flagObBucketDisplayName, "display-name", "", "User friendly display name")
		c.Flags().StringVar(&g.flagObBucketDescription, "description", "", "Description of the bucket")
		c.Flags().StringVar(&g.flagObBucketKmsKey, "cmek-settings-kms-key", "",
			"The resource name for the configured Cloud KMS key")
	}
	g.obBucketUpdateCmd.Flags().BoolVar(&g.flagObBucketClearCmek, "clear-cmek-settings", false,
		"Clear the CMEK settings of the bucket")
	g.obBucketUpdateCmd.MarkFlagsMutuallyExclusive("cmek-settings-kms-key", "clear-cmek-settings")
	g.obBucketListCmd.Flags().BoolVar(&g.flagObBucketShowDeleted, "show-deleted", false,
		"If true, then the response will include deleted buckets")
	g.obBucketListCmd.Flags().Int64Var(&g.flagObBucketPageSize, "page-size", 0, "Maximum results per page")

	g.observabilityBucketsCmd.AddCommand(all...)
	g.observabilityCmd.AddCommand(g.observabilityBucketsCmd)
}

// obBucketName resolves BUCKET (an ID with --location, or a full name) to its
// relative name and ID.
func (g *observabilityCmds) obBucketName(arg string) (string, string, error) {
	if strings.HasPrefix(arg, "projects/") {
		return arg, arg[strings.LastIndex(arg, "/")+1:], nil
	}
	parent, err := g.obBucketParent()
	if err != nil {
		return "", "", err
	}
	return parent + "/buckets/" + arg, arg, nil
}

func (g *observabilityCmds) obBucketParent() (string, error) {
	if g.flagObBucketLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	project, err := g.inv.project()
	if err != nil {
		return "", err
	}
	return obLocationParent(project, g.flagObBucketLocation), nil
}

// obBucketWait prints gcloud's declarative LRO messages and waits unless
// --async. It reports whether it waited.
func (g *observabilityCmds) obBucketWait(ctx context.Context, svc *observability.Service, op *observability.Operation, requestMsg string) (bool, error) {
	fmt.Fprintln(g.inv.stderr, requestMsg)
	if g.flagObBucketAsync {
		fmt.Fprintf(g.inv.stderr, "Check operation [%s] for status.\n", op.Name)
		return false, nil
	}
	fmt.Fprintf(g.inv.stderr, "Waiting for operation [%s] to complete...", op.Name)
	for !op.Done {
		time.Sleep(time.Second)
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
	return true, nil
}

func (g *observabilityCmds) runObBucketCreate(cmd *cobra.Command, args []string) error {
	parent, err := g.obBucketParent()
	if err != nil {
		return err
	}
	body := &observability.Bucket{
		DisplayName: g.flagObBucketDisplayName,
		Description: g.flagObBucketDescription,
	}
	if g.flagObBucketKmsKey != "" {
		body.CmekSettings = &observability.CmekSettings{KmsKey: g.flagObBucketKmsKey}
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Observability(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Buckets.Create(parent, body).BucketId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating bucket: %w", err)
	}
	waited, err := g.obBucketWait(ctx, svc, op, fmt.Sprintf("Create request issued for: [%s]", args[0]))
	if err != nil {
		return err
	}
	if waited {
		fmt.Fprintf(g.inv.stderr, "Created bucket [%s].\n", args[0])
	}
	// gcloud-python's create is a SilentCommand: no stdout unless --format.
	if g.flagObBucketFormat != "" {
		return emitFormattedTo(g.inv.stdout, op, g.flagObBucketFormat)
	}
	return nil
}

func (g *observabilityCmds) runObBucketDescribe(cmd *cobra.Command, args []string) error {
	name, _, err := g.obBucketName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Observability(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Buckets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing bucket: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagObBucketFormat)
}

func (g *observabilityCmds) runObBucketList(cmd *cobra.Command, args []string) error {
	parent, err := g.obBucketParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Observability(ctx)
	if err != nil {
		return err
	}
	all := []*observability.Bucket{}
	pageToken := ""
	for {
		call := svc.Projects.Locations.Buckets.List(parent).Context(ctx)
		if g.flagObBucketShowDeleted {
			call = call.ShowDeleted(true)
		}
		if g.flagObBucketPageSize > 0 {
			call = call.PageSize(g.flagObBucketPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing buckets: %w", err)
		}
		all = append(all, resp.Buckets...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if err := emitFormattedTo(g.inv.stdout, all, g.flagObBucketFormat); err != nil {
		return err
	}
	if g.flagObBucketFormat == "" && len(all) == 0 {
		fmt.Fprintln(g.inv.stderr, "Listed 0 items.")
	}
	return nil
}

// obBucketUpdateMask applies the update flags to b and returns the update
// mask, as the declarative read-modify-update command does.
func (g *observabilityCmds) obBucketUpdateMask(cmd *cobra.Command, b *observability.Bucket) string {
	var mask []string
	if cmd.Flags().Changed("display-name") {
		b.DisplayName = g.flagObBucketDisplayName
		mask = append(mask, "displayName")
	}
	if cmd.Flags().Changed("description") {
		b.Description = g.flagObBucketDescription
		mask = append(mask, "description")
	}
	if cmd.Flags().Changed("cmek-settings-kms-key") {
		if b.CmekSettings == nil {
			b.CmekSettings = &observability.CmekSettings{}
		}
		b.CmekSettings.KmsKey = g.flagObBucketKmsKey
		mask = append(mask, "cmekSettings.kmsKey")
	}
	if g.flagObBucketClearCmek {
		b.CmekSettings = nil
		mask = append(mask, "cmekSettings")
	}
	return strings.Join(mask, ",")
}

func (g *observabilityCmds) runObBucketUpdate(cmd *cobra.Command, args []string) error {
	name, id, err := g.obBucketName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := g.inv.clients().Observability(ctx)
	if err != nil {
		return err
	}
	existing, err := svc.Projects.Locations.Buckets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting bucket: %w", err)
	}
	mask := g.obBucketUpdateMask(cmd, existing)
	if mask == "" {
		return fmt.Errorf("at least one of --display-name, --description, --cmek-settings-kms-key or --clear-cmek-settings is required")
	}
	op, err := svc.Projects.Locations.Buckets.Patch(name, existing).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating bucket: %w", err)
	}
	waited, err := g.obBucketWait(ctx, svc, op, fmt.Sprintf("Request issued for: [%s]", id))
	if err != nil {
		return err
	}
	fmt.Fprintf(g.inv.stderr, "Updated bucket [%s].\n", id)
	if !waited {
		return emitFormattedTo(g.inv.stdout, op, g.flagObBucketFormat)
	}
	got, err := svc.Projects.Locations.Buckets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting bucket: %w", err)
	}
	return emitFormattedTo(g.inv.stdout, got, g.flagObBucketFormat)
}
