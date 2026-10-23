package cmd

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	observability "google.golang.org/api/observability/v1"
)

// --- gcloud observability buckets (#1936) ---
//
// Declarative commands in gcloud-python on observability.projects.locations.
// buckets. create and update (added in 586.0.0) return an operation that is
// waited on unless --async.

var observabilityBucketsCmd = &cobra.Command{Use: "buckets", Short: "Manage observability buckets"}

var (
	obBucketCreateCmd = &cobra.Command{
		Use: "create BUCKET", Short: "Create an observability bucket",
		Args: cobra.ExactArgs(1), RunE: runObBucketCreate,
	}
	obBucketDescribeCmd = &cobra.Command{
		Use: "describe BUCKET", Short: "Describe an observability bucket",
		Args: cobra.ExactArgs(1), RunE: runObBucketDescribe,
	}
	obBucketListCmd = &cobra.Command{
		Use: "list", Short: "List observability buckets",
		Args: cobra.NoArgs, RunE: runObBucketList,
	}
	obBucketUpdateCmd = &cobra.Command{
		Use: "update BUCKET", Short: "Update an observability bucket",
		Args: cobra.ExactArgs(1), RunE: runObBucketUpdate,
	}
)

var (
	flagObBucketLocation    string
	flagObBucketFormat      string
	flagObBucketAsync       bool
	flagObBucketDisplayName string
	flagObBucketDescription string
	flagObBucketKmsKey      string
	flagObBucketClearCmek   bool
	flagObBucketShowDeleted bool
	flagObBucketPageSize    int64
)

func init() {
	all := []*cobra.Command{obBucketCreateCmd, obBucketDescribeCmd, obBucketListCmd, obBucketUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagObBucketLocation, "location", "", "The location id of the bucket resource")
		c.Flags().StringVar(&flagObBucketFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{obBucketCreateCmd, obBucketUpdateCmd} {
		c.Flags().BoolVar(&flagObBucketAsync, "async", false,
			"Return immediately, without waiting for the operation in progress to complete")
		c.Flags().StringVar(&flagObBucketDisplayName, "display-name", "", "User friendly display name")
		c.Flags().StringVar(&flagObBucketDescription, "description", "", "Description of the bucket")
		c.Flags().StringVar(&flagObBucketKmsKey, "cmek-settings-kms-key", "",
			"The resource name for the configured Cloud KMS key")
	}
	obBucketUpdateCmd.Flags().BoolVar(&flagObBucketClearCmek, "clear-cmek-settings", false,
		"Clear the CMEK settings of the bucket")
	obBucketUpdateCmd.MarkFlagsMutuallyExclusive("cmek-settings-kms-key", "clear-cmek-settings")
	obBucketListCmd.Flags().BoolVar(&flagObBucketShowDeleted, "show-deleted", false,
		"If true, then the response will include deleted buckets")
	obBucketListCmd.Flags().Int64Var(&flagObBucketPageSize, "page-size", 0, "Maximum results per page")

	observabilityBucketsCmd.AddCommand(all...)
	observabilityCmd.AddCommand(observabilityBucketsCmd)
}

// obBucketName resolves BUCKET (an ID with --location, or a full name) to its
// relative name and ID.
func obBucketName(arg string) (string, string, error) {
	if strings.HasPrefix(arg, "projects/") {
		return arg, arg[strings.LastIndex(arg, "/")+1:], nil
	}
	parent, err := obBucketParent()
	if err != nil {
		return "", "", err
	}
	return parent + "/buckets/" + arg, arg, nil
}

func obBucketParent() (string, error) {
	if flagObBucketLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return obLocationParent(project, flagObBucketLocation), nil
}

// obBucketWait prints gcloud's declarative LRO messages and waits unless
// --async. It reports whether it waited.
func obBucketWait(ctx context.Context, svc *observability.Service, op *observability.Operation, requestMsg string) (bool, error) {
	fmt.Fprintln(os.Stderr, requestMsg)
	if flagObBucketAsync {
		fmt.Fprintf(os.Stderr, "Check operation [%s] for status.\n", op.Name)
		return false, nil
	}
	fmt.Fprintf(os.Stderr, "Waiting for operation [%s] to complete...", op.Name)
	for !op.Done {
		time.Sleep(time.Second)
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
	return true, nil
}

func runObBucketCreate(cmd *cobra.Command, args []string) error {
	parent, err := obBucketParent()
	if err != nil {
		return err
	}
	body := &observability.Bucket{
		DisplayName: flagObBucketDisplayName,
		Description: flagObBucketDescription,
	}
	if flagObBucketKmsKey != "" {
		body.CmekSettings = &observability.CmekSettings{KmsKey: flagObBucketKmsKey}
	}
	ctx := context.Background()
	svc, err := gcp.ObservabilityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Buckets.Create(parent, body).BucketId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating bucket: %w", err)
	}
	waited, err := obBucketWait(ctx, svc, op, fmt.Sprintf("Create request issued for: [%s]", args[0]))
	if err != nil {
		return err
	}
	if waited {
		fmt.Fprintf(os.Stderr, "Created bucket [%s].\n", args[0])
	}
	// gcloud-python's create is a SilentCommand: no stdout unless --format.
	if flagObBucketFormat != "" {
		return emitFormatted(op, flagObBucketFormat)
	}
	return nil
}

func runObBucketDescribe(cmd *cobra.Command, args []string) error {
	name, _, err := obBucketName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ObservabilityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Buckets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing bucket: %w", err)
	}
	return emitFormatted(got, flagObBucketFormat)
}

func runObBucketList(cmd *cobra.Command, args []string) error {
	parent, err := obBucketParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ObservabilityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	all := []*observability.Bucket{}
	pageToken := ""
	for {
		call := svc.Projects.Locations.Buckets.List(parent).Context(ctx)
		if flagObBucketShowDeleted {
			call = call.ShowDeleted(true)
		}
		if flagObBucketPageSize > 0 {
			call = call.PageSize(flagObBucketPageSize)
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
	if err := emitFormatted(all, flagObBucketFormat); err != nil {
		return err
	}
	if flagObBucketFormat == "" && len(all) == 0 {
		fmt.Fprintln(os.Stderr, "Listed 0 items.")
	}
	return nil
}

// obBucketUpdateMask applies the update flags to b and returns the update
// mask, as the declarative read-modify-update command does.
func obBucketUpdateMask(cmd *cobra.Command, b *observability.Bucket) string {
	var mask []string
	if cmd.Flags().Changed("display-name") {
		b.DisplayName = flagObBucketDisplayName
		mask = append(mask, "displayName")
	}
	if cmd.Flags().Changed("description") {
		b.Description = flagObBucketDescription
		mask = append(mask, "description")
	}
	if cmd.Flags().Changed("cmek-settings-kms-key") {
		if b.CmekSettings == nil {
			b.CmekSettings = &observability.CmekSettings{}
		}
		b.CmekSettings.KmsKey = flagObBucketKmsKey
		mask = append(mask, "cmekSettings.kmsKey")
	}
	if flagObBucketClearCmek {
		b.CmekSettings = nil
		mask = append(mask, "cmekSettings")
	}
	return strings.Join(mask, ",")
}

func runObBucketUpdate(cmd *cobra.Command, args []string) error {
	name, id, err := obBucketName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ObservabilityService(ctx, flagAccount)
	if err != nil {
		return err
	}
	existing, err := svc.Projects.Locations.Buckets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting bucket: %w", err)
	}
	mask := obBucketUpdateMask(cmd, existing)
	if mask == "" {
		return fmt.Errorf("at least one of --display-name, --description, --cmek-settings-kms-key or --clear-cmek-settings is required")
	}
	op, err := svc.Projects.Locations.Buckets.Patch(name, existing).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating bucket: %w", err)
	}
	waited, err := obBucketWait(ctx, svc, op, fmt.Sprintf("Request issued for: [%s]", id))
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Updated bucket [%s].\n", id)
	if !waited {
		return emitFormatted(op, flagObBucketFormat)
	}
	got, err := svc.Projects.Locations.Buckets.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting bucket: %w", err)
	}
	return emitFormatted(got, flagObBucketFormat)
}
