package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
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

var (
	flagStorageBucketsFormat      string
	flagStorageBucketsSoftDeleted bool
)

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

var (
	flagStorageCpRecurse      bool
	flagStorageCpNoClobber    bool
	flagStorageCpStorageClass string
)

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

var (
	flagStorageLsRecurse bool
	flagStorageLsLong    bool
	flagStorageLsJSON    bool
)

func init() {
	storageBucketsListCmd.Flags().StringVar(&flagStorageBucketsFormat, "format", "", "Output format (e.g. json)")
	storageBucketsListCmd.Flags().BoolVar(&flagStorageBucketsSoftDeleted, "soft-deleted", false, "Include soft-deleted buckets")
	storageBucketsCmd.AddCommand(storageBucketsListCmd)

	storageCpCmd.Flags().BoolVarP(&flagStorageCpRecurse, "recursive", "r", false, "Copy recursively")
	storageCpCmd.Flags().BoolVarP(&flagStorageCpNoClobber, "no-clobber", "n", false, "Do not overwrite existing files")
	storageCpCmd.Flags().StringVar(&flagStorageCpStorageClass, "storage-class", "", "Storage class for uploaded objects")

	storageLsCmd.Flags().BoolVarP(&flagStorageLsRecurse, "recursive", "r", false, "List recursively")
	storageLsCmd.Flags().BoolVarP(&flagStorageLsLong, "long", "l", false, "Show size and creation time")
	storageLsCmd.Flags().BoolVar(&flagStorageLsJSON, "json", false, "Output as JSON")

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

	var allBuckets []*storage.Bucket
	pageToken := ""
	for {
		call := svc.Buckets.List(project).Context(ctx)
		if flagStorageBucketsSoftDeleted {
			call = call.SoftDeleted(true)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing buckets: %w", err)
		}
		allBuckets = append(allBuckets, resp.Items...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagStorageBucketsFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allBuckets)
	}

	for _, b := range allBuckets {
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

	if flagStorageCpRecurse {
		return storageDownloadRecursive(ctx, svc, bucket, object, dst)
	}

	return storageDownloadFile(ctx, svc, bucket, object, dst)
}

func storageDownloadFile(ctx context.Context, svc *storage.Service, bucket, object, dst string) error {
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

	if flagStorageCpNoClobber {
		if _, err := os.Stat(dst); err == nil {
			fmt.Printf("Skipping existing file: %s\n", dst)
			return nil
		}
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

func storageDownloadRecursive(ctx context.Context, svc *storage.Service, bucket, prefix, dst string) error {
	// Ensure prefix ends with "/" for directory-like listing.
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	var objects []*storage.Object
	pageToken := ""
	for {
		call := svc.Objects.List(bucket).Prefix(prefix).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing objects: %w", err)
		}
		objects = append(objects, resp.Items...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if len(objects) == 0 {
		return fmt.Errorf("no objects found under gs://%s/%s", bucket, prefix)
	}

	for _, obj := range objects {
		// Skip directory marker objects.
		if strings.HasSuffix(obj.Name, "/") {
			continue
		}
		relPath := strings.TrimPrefix(obj.Name, prefix)
		localPath := filepath.Join(dst, filepath.FromSlash(relPath))

		if err := os.MkdirAll(filepath.Dir(localPath), 0755); err != nil {
			return fmt.Errorf("creating directory: %w", err)
		}

		if err := storageDownloadFile(ctx, svc, bucket, obj.Name, localPath); err != nil {
			return err
		}
	}
	return nil
}

func storageUpload(ctx context.Context, svc *storage.Service, src, dst string) error {
	bucket, prefix, err := parseGCSPath(dst)
	if err != nil {
		return err
	}

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat %s: %w", src, err)
	}

	if info.IsDir() {
		if !flagStorageCpRecurse {
			return fmt.Errorf("%s is a directory; use --recursive or -r to copy directories", src)
		}
		return storageUploadRecursive(ctx, svc, src, bucket, prefix)
	}

	return storageUploadFile(ctx, svc, src, bucket, prefix)
}

func storageUploadFile(ctx context.Context, svc *storage.Service, src, bucket, prefix string) error {
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
	if flagStorageCpStorageClass != "" {
		obj.StorageClass = flagStorageCpStorageClass
	}
	call := svc.Objects.Insert(bucket, obj).Media(f).Context(ctx)
	if flagStorageCpNoClobber {
		call = call.IfGenerationMatch(0)
	}
	_, err = call.Do()
	if err != nil {
		return fmt.Errorf("uploading object: %w", err)
	}

	fmt.Printf("Copied %s to gs://%s/%s\n", src, bucket, objectName)
	return nil
}

func storageUploadRecursive(ctx context.Context, svc *storage.Service, srcDir, bucket, prefix string) error {
	// Ensure prefix ends with "/" when set.
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		objectName := prefix + filepath.ToSlash(relPath)

		return storageUploadFile(ctx, svc, path, bucket, objectName)
	})
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

	if flagStorageCpRecurse {
		return storageGCSCopyRecursive(ctx, svc, srcBucket, srcObject, dstBucket, dstObject)
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

func storageGCSCopyRecursive(ctx context.Context, svc *storage.Service, srcBucket, srcPrefix, dstBucket, dstPrefix string) error {
	// Ensure prefixes end with "/" for directory-like listing.
	if srcPrefix != "" && !strings.HasSuffix(srcPrefix, "/") {
		srcPrefix += "/"
	}
	if dstPrefix != "" && !strings.HasSuffix(dstPrefix, "/") {
		dstPrefix += "/"
	}

	var objects []*storage.Object
	pageToken := ""
	for {
		call := svc.Objects.List(srcBucket).Prefix(srcPrefix).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing objects: %w", err)
		}
		objects = append(objects, resp.Items...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if len(objects) == 0 {
		return fmt.Errorf("no objects found under gs://%s/%s", srcBucket, srcPrefix)
	}

	for _, obj := range objects {
		if strings.HasSuffix(obj.Name, "/") {
			continue
		}
		relPath := strings.TrimPrefix(obj.Name, srcPrefix)
		dstObject := dstPrefix + relPath

		_, err := svc.Objects.Copy(srcBucket, obj.Name, dstBucket, dstObject, nil).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("copying gs://%s/%s: %w", srcBucket, obj.Name, err)
		}
		fmt.Printf("Copied gs://%s/%s to gs://%s/%s\n", srcBucket, obj.Name, dstBucket, dstObject)
	}
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
		pageToken := ""
		for {
			call := svc.Buckets.List(project).Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing buckets: %w", err)
			}
			for _, b := range resp.Items {
				fmt.Printf("gs://%s/\n", b.Name)
			}
			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}
		return nil
	}

	bucket, prefix, err := parseGCSPath(args[0])
	if err != nil {
		return err
	}

	var allPrefixes []string
	var allObjects []*storage.Object
	pageToken := ""
	for {
		call := svc.Objects.List(bucket).Prefix(prefix).Context(ctx)
		if !flagStorageLsRecurse {
			call = call.Delimiter("/")
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing objects: %w", err)
		}
		allPrefixes = append(allPrefixes, resp.Prefixes...)
		allObjects = append(allObjects, resp.Items...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagStorageLsJSON {
		type lsEntry struct {
			URL  string `json:"url"`
			Size uint64 `json:"size,omitempty"`
			Time string `json:"timeCreated,omitempty"`
		}
		var entries []lsEntry
		for _, p := range allPrefixes {
			entries = append(entries, lsEntry{URL: fmt.Sprintf("gs://%s/%s", bucket, p)})
		}
		for _, obj := range allObjects {
			entries = append(entries, lsEntry{URL: fmt.Sprintf("gs://%s/%s", bucket, obj.Name), Size: obj.Size, Time: obj.TimeCreated})
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(entries)
	}

	// Print directories (prefixes).
	for _, p := range allPrefixes {
		fmt.Printf("gs://%s/%s\n", bucket, p)
	}
	// Print objects.
	for _, obj := range allObjects {
		if flagStorageLsLong {
			fmt.Printf("%10d  %s  gs://%s/%s\n", obj.Size, obj.TimeCreated, bucket, obj.Name)
		} else {
			fmt.Printf("gs://%s/%s\n", bucket, obj.Name)
		}
	}
	return nil
}
