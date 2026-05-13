package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/flyingobsidian/gcloud-golang-cli/internal/gcp"
	"github.com/spf13/cobra"
	storage "google.golang.org/api/storage/v1"
)

var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Manage Cloud Storage",
}

// --- buckets list ---

var storageBucketsCmd = &cobra.Command{
	Use:   "buckets",
	Short: "Manage Cloud Storage buckets",
}

var storageBucketsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Cloud Storage buckets",
	Args:  cobra.NoArgs,
	RunE:  runStorageBucketsList,
}

var flagStorageBucketsFormat string

// --- cp ---

var storageCpCmd = &cobra.Command{
	Use:   "cp SOURCE DESTINATION",
	Short: "Copy files to/from Cloud Storage",
	Long: `Copy files between local filesystem and Cloud Storage.
Examples:
  gcloud storage cp file.txt gs://bucket/path/
  gcloud storage cp gs://bucket/path/file.txt ./local/
  gcloud storage cp -r ./dir gs://bucket/path/`,
	Args: cobra.ExactArgs(2),
	RunE: runStorageCp,
}

var flagStorageCpRecurse bool

// --- ls ---

var storageLsCmd = &cobra.Command{
	Use:   "ls [GCS_PATH]",
	Short: "List Cloud Storage objects",
	Long: `List objects and buckets in Cloud Storage.
Examples:
  gcloud storage ls gs://bucket
  gcloud storage ls gs://bucket/prefix/`,
	Args: cobra.MaximumNArgs(1),
	RunE: runStorageLs,
}

var flagStorageLsRecurse bool

func init() {
	storageBucketsListCmd.Flags().StringVar(&flagStorageBucketsFormat, "format", "", "Output format (e.g. json)")
	storageBucketsCmd.AddCommand(storageBucketsListCmd)

	storageCpCmd.Flags().BoolVarP(&flagStorageCpRecurse, "recursive", "r", false, "Copy recursively")

	storageLsCmd.Flags().BoolVarP(&flagStorageLsRecurse, "recursive", "r", false, "List recursively")

	storageCmd.AddCommand(storageBucketsCmd)
	storageCmd.AddCommand(storageCpCmd)
	storageCmd.AddCommand(storageLsCmd)
	rootCmd.AddCommand(storageCmd)
}

func runStorageBucketsList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}

	resp, err := svc.Buckets.List(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing buckets: %w", err)
	}

	if flagStorageBucketsFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(resp.Items)
	}

	for _, b := range resp.Items {
		fmt.Printf("gs://%s/\n", b.Name)
	}
	return nil
}

// parseGCSPath splits "gs://bucket/prefix" into bucket and prefix.
func parseGCSPath(p string) (bucket, prefix string, err error) {
	if !strings.HasPrefix(p, "gs://") {
		return "", "", fmt.Errorf("not a GCS path: %s", p)
	}
	p = strings.TrimPrefix(p, "gs://")
	if i := strings.IndexByte(p, '/'); i >= 0 {
		return p[:i], p[i+1:], nil
	}
	return p, "", nil
}

func isGCSPath(p string) bool {
	return strings.HasPrefix(p, "gs://")
}

func runStorageCp(cmd *cobra.Command, args []string) error {
	src, dst := args[0], args[1]

	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}

	switch {
	case isGCSPath(src) && !isGCSPath(dst):
		// Download: gs://bucket/obj -> local
		return storageDownload(ctx, svc, src, dst)
	case !isGCSPath(src) && isGCSPath(dst):
		// Upload: local -> gs://bucket/obj
		return storageUpload(ctx, svc, src, dst)
	case isGCSPath(src) && isGCSPath(dst):
		// Copy between GCS paths
		return storageGCSCopy(ctx, svc, src, dst)
	default:
		return fmt.Errorf("at least one of source or destination must be a gs:// path")
	}
}

func storageDownload(ctx context.Context, svc *storage.Service, src, dst string) error {
	bucket, object, err := parseGCSPath(src)
	if err != nil {
		return err
	}

	resp, err := svc.Objects.Get(bucket, object).Context(ctx).Download()
	if err != nil {
		return fmt.Errorf("downloading object: %w", err)
	}
	defer resp.Body.Close()

	// If dst is a directory, use the object's base name.
	info, statErr := os.Stat(dst)
	if statErr == nil && info.IsDir() {
		dst = filepath.Join(dst, filepath.Base(object))
	}

	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating local file: %w", err)
	}
	defer f.Close()

	n, err := io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("writing to local file: %w", err)
	}

	fmt.Printf("Copied gs://%s/%s to %s (%d bytes)\n", bucket, object, dst, n)
	return nil
}

func storageUpload(ctx context.Context, svc *storage.Service, src, dst string) error {
	bucket, prefix, err := parseGCSPath(dst)
	if err != nil {
		return err
	}

	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening local file: %w", err)
	}
	defer f.Close()

	objectName := prefix
	if objectName == "" || strings.HasSuffix(objectName, "/") {
		objectName += filepath.Base(src)
	}

	obj := &storage.Object{Name: objectName}
	_, err = svc.Objects.Insert(bucket, obj).Media(f).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("uploading object: %w", err)
	}

	fmt.Printf("Copied %s to gs://%s/%s\n", src, bucket, objectName)
	return nil
}

func storageGCSCopy(ctx context.Context, svc *storage.Service, src, dst string) error {
	srcBucket, srcObject, err := parseGCSPath(src)
	if err != nil {
		return err
	}
	dstBucket, dstObject, err := parseGCSPath(dst)
	if err != nil {
		return err
	}

	if dstObject == "" || strings.HasSuffix(dstObject, "/") {
		dstObject += filepath.Base(srcObject)
	}

	_, err = svc.Objects.Copy(srcBucket, srcObject, dstBucket, dstObject, nil).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("copying object: %w", err)
	}

	fmt.Printf("Copied gs://%s/%s to gs://%s/%s\n", srcBucket, srcObject, dstBucket, dstObject)
	return nil
}

func runStorageLs(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}

	// No args = list buckets.
	if len(args) == 0 {
		project, err := resolveProject()
		if err != nil {
			return err
		}
		resp, err := svc.Buckets.List(project).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("listing buckets: %w", err)
		}
		for _, b := range resp.Items {
			fmt.Printf("gs://%s/\n", b.Name)
		}
		return nil
	}

	bucket, prefix, err := parseGCSPath(args[0])
	if err != nil {
		return err
	}

	call := svc.Objects.List(bucket).Prefix(prefix).Context(ctx)
	if !flagStorageLsRecurse {
		call = call.Delimiter("/")
	}

	resp, err := call.Do()
	if err != nil {
		return fmt.Errorf("listing objects: %w", err)
	}

	// Print directories (prefixes).
	for _, p := range resp.Prefixes {
		fmt.Printf("gs://%s/%s\n", bucket, p)
	}
	// Print objects.
	for _, obj := range resp.Items {
		fmt.Printf("gs://%s/%s\n", bucket, obj.Name)
	}
	return nil
}
