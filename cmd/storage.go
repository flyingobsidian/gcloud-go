package cmd

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash/crc32"
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

// --- cat ---

var storageCatCmd = &cobra.Command{
	Use:   "cat GCS_PATH",
	Short: "Output the contents of a Cloud Storage object to stdout",
	Long: `Output the contents of a Cloud Storage object to stdout.
Examples:
  gcloud storage cat gs://bucket/path/file.txt`,
	Args: cobra.ExactArgs(1),
	RunE: runStorageCat,
}

// --- rm ---

var storageRmCmd = &cobra.Command{
	Use:   "rm GCS_PATH",
	Short: "Delete Cloud Storage objects",
	Long: `Delete objects from Cloud Storage.
Examples:
  gcloud storage rm gs://bucket/path/file.txt
  gcloud storage rm -r gs://bucket/path/`,
	Args: cobra.ExactArgs(1),
	RunE: runStorageRm,
}

var (
	flagStorageRmRecurse       bool
	flagStorageRmAllVersions   bool
	flagStorageRmContinueOnErr bool
)

// --- buckets describe ---

var storageBucketsDescribeCmd = &cobra.Command{
	Use:   "describe gs://BUCKET",
	Short: "Describe a Cloud Storage bucket",
	Args:  cobra.ExactArgs(1),
	RunE:  runStorageBucketsDescribe,
}

// --- mv ---

var storageMvCmd = &cobra.Command{
	Use:   "mv SOURCE DESTINATION",
	Short: "Move/rename Cloud Storage objects",
	Args:  cobra.ExactArgs(2),
	RunE:  runStorageMv,
}

var flagStorageMvRecurse bool

// --- rsync ---

var storageRsyncCmd = &cobra.Command{
	Use:   "rsync SOURCE DESTINATION",
	Short: "Synchronize content of two buckets/directories",
	Args:  cobra.ExactArgs(2),
	RunE:  runStorageRsync,
}

var (
	flagRsyncRecurse       bool
	flagRsyncDeleteUnmatched bool
	flagRsyncDryRun        bool
)

// --- du ---

var storageDuCmd = &cobra.Command{
	Use:   "du [GCS_PATH]",
	Short: "Display disk usage of Cloud Storage objects",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runStorageDu,
}

var (
	flagDuSummarize bool
	flagDuReadable  bool
)

// --- hash ---

var storageHashCmd = &cobra.Command{
	Use:   "hash PATH",
	Short: "Compute hashes for local files or Cloud Storage objects",
	Args:  cobra.ExactArgs(1),
	RunE:  runStorageHash,
}

// --- cp extra flags ---

var (
	flagStorageCpContentType   string
	flagStorageCpCacheControl  string
	flagStorageCpContinueOnErr bool
)

// --- buckets list extra flags ---

var (
	flagBucketsListFilter string
	flagBucketsListPrefix string
	flagBucketsListURI    bool
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
	storageBucketsListCmd.Flags().StringVar(&flagBucketsListFilter, "filter", "", "Filter expression")
	storageBucketsListCmd.Flags().StringVar(&flagBucketsListPrefix, "prefix", "", "Filter buckets by name prefix")
	storageBucketsListCmd.Flags().BoolVar(&flagBucketsListURI, "uri", false, "Print gs:// URIs")
	storageBucketsCmd.AddCommand(storageBucketsListCmd)
	storageBucketsCmd.AddCommand(storageBucketsDescribeCmd)

	storageCpCmd.Flags().BoolVarP(&flagStorageCpRecurse, "recursive", "r", false, "Copy recursively")
	storageCpCmd.Flags().BoolVarP(&flagStorageCpNoClobber, "no-clobber", "n", false, "Do not overwrite existing files")
	storageCpCmd.Flags().StringVar(&flagStorageCpStorageClass, "storage-class", "", "Storage class for uploaded objects")
	storageCpCmd.Flags().StringVar(&flagStorageCpContentType, "content-type", "", "Content-Type for uploaded objects")
	storageCpCmd.Flags().StringVar(&flagStorageCpCacheControl, "cache-control", "", "Cache-Control header for uploaded objects")
	storageCpCmd.Flags().BoolVar(&flagStorageCpContinueOnErr, "continue-on-error", false, "Skip failures and continue")

	storageRmCmd.Flags().BoolVarP(&flagStorageRmRecurse, "recursive", "r", false, "Delete recursively")
	storageRmCmd.Flags().BoolVar(&flagStorageRmAllVersions, "all-versions", false, "Delete all object versions")
	storageRmCmd.Flags().BoolVar(&flagStorageRmContinueOnErr, "continue-on-error", false, "Skip failures and continue")

	storageLsCmd.Flags().BoolVarP(&flagStorageLsRecurse, "recursive", "r", false, "List recursively")
	storageLsCmd.Flags().BoolVarP(&flagStorageLsLong, "long", "l", false, "Show size and creation time")
	storageLsCmd.Flags().BoolVar(&flagStorageLsJSON, "json", false, "Output as JSON")

	storageMvCmd.Flags().BoolVarP(&flagStorageMvRecurse, "recursive", "r", false, "Move recursively")

	storageRsyncCmd.Flags().BoolVarP(&flagRsyncRecurse, "recursive", "r", false, "Sync recursively")
	storageRsyncCmd.Flags().BoolVarP(&flagRsyncDeleteUnmatched, "delete-unmatched-destination-objects", "d", false, "Delete destination objects not in source")
	storageRsyncCmd.Flags().BoolVarP(&flagRsyncDryRun, "dry-run", "n", false, "Show what would be synced")

	storageDuCmd.Flags().BoolVarP(&flagDuSummarize, "summarize", "s", false, "Show total only")
	storageDuCmd.Flags().BoolVarP(&flagDuReadable, "readable", "h", false, "Human-readable sizes")

	storageCmd.AddCommand(storageBucketsCmd)
	storageCmd.AddCommand(storageCatCmd)
	storageCmd.AddCommand(storageCpCmd)
	storageCmd.AddCommand(storageLsCmd)
	storageCmd.AddCommand(storageMvCmd)
	storageCmd.AddCommand(storageRmCmd)
	storageCmd.AddCommand(storageRsyncCmd)
	storageCmd.AddCommand(storageDuCmd)
	storageCmd.AddCommand(storageHashCmd)

	// gcloud-python storage subcommands/subgroups not yet implemented (#547).
	registerStubCommand(storageCmd, "diagnose", "Diagnose Cloud Storage issues")
	registerStubGroup(storageCmd, "insights",
		"Manage Storage Insights",
		"describe", "list")
	registerStubGroup(storageCmd, "intelligence-configs",
		"Manage Storage Intelligence configurations",
		"describe", "update")
	registerStubGroup(storageCmd, "intelligence-findings",
		"Manage Storage Intelligence findings",
		"list", "describe")
	registerStubGroup(storageCmd, "objects",
		"Manage Cloud Storage objects (object-scoped ops)",
		"describe", "list", "update", "compose")
	registerStubCommand(storageCmd, "restore",
		"Restore soft-deleted objects")
	registerStubCommand(storageCmd, "sign-url",
		"Generate a signed URL for an object")

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
		if flagBucketsListPrefix != "" {
			call = call.Prefix(flagBucketsListPrefix)
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

	// Client-side filter (simple substring match on name).
	if flagBucketsListFilter != "" {
		var filtered []*storage.Bucket
		for _, b := range allBuckets {
			if strings.Contains(b.Name, flagBucketsListFilter) {
				filtered = append(filtered, b)
			}
		}
		allBuckets = filtered
	}

	if flagStorageBucketsFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(allBuckets)
	}

	if flagBucketsListURI {
		for _, b := range allBuckets {
			fmt.Printf("gs://%s/\n", b.Name)
		}
		return nil
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
	if flagStorageCpContentType != "" {
		obj.ContentType = flagStorageCpContentType
	}
	if flagStorageCpCacheControl != "" {
		obj.CacheControl = flagStorageCpCacheControl
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

func runStorageCat(cmd *cobra.Command, args []string) error {
	bucket, object, err := parseGCSPath(args[0])
	if err != nil {
		return err
	}
	if object == "" {
		return fmt.Errorf("object path is required: gs://bucket/object")
	}

	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}

	resp, err := svc.Objects.Get(bucket, object).Context(ctx).Download()
	if err != nil {
		return fmt.Errorf("reading object: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(os.Stdout, resp.Body)
	return err
}

func runStorageRm(cmd *cobra.Command, args []string) error {
	bucket, object, err := parseGCSPath(args[0])
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}

	if flagStorageRmRecurse {
		prefix := object
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		pageToken := ""
		var lastErr error
		for {
			call := svc.Objects.List(bucket).Prefix(prefix).Context(ctx)
			if flagStorageRmAllVersions {
				call = call.Versions(true)
			}
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing objects: %w", err)
			}
			for _, obj := range resp.Items {
				delCall := svc.Objects.Delete(bucket, obj.Name).Context(ctx)
				if flagStorageRmAllVersions && obj.Generation != 0 {
					delCall = delCall.Generation(int64(obj.Generation))
				}
				if err := delCall.Do(); err != nil {
					if flagStorageRmContinueOnErr {
						fmt.Fprintf(os.Stderr, "WARNING: failed to delete gs://%s/%s: %v\n", bucket, obj.Name, err)
						lastErr = err
						continue
					}
					return fmt.Errorf("deleting gs://%s/%s: %w", bucket, obj.Name, err)
				}
				fmt.Fprintf(os.Stderr, "Removing gs://%s/%s\n", bucket, obj.Name)
			}
			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}
		if lastErr != nil {
			return fmt.Errorf("some objects could not be deleted")
		}
		return nil
	}

	if object == "" {
		return fmt.Errorf("object path is required (use -r to delete all objects under a prefix)")
	}

	if flagStorageRmAllVersions {
		// Delete all versions of a single object.
		call := svc.Objects.List(bucket).Prefix(object).Versions(true).Context(ctx)
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing object versions: %w", err)
		}
		for _, obj := range resp.Items {
			if obj.Name != object {
				continue
			}
			delCall := svc.Objects.Delete(bucket, obj.Name).Context(ctx)
			if obj.Generation != 0 {
				delCall = delCall.Generation(int64(obj.Generation))
			}
			if err := delCall.Do(); err != nil {
				if flagStorageRmContinueOnErr {
					fmt.Fprintf(os.Stderr, "WARNING: failed to delete gs://%s/%s: %v\n", bucket, obj.Name, err)
					continue
				}
				return fmt.Errorf("deleting gs://%s/%s: %w", bucket, obj.Name, err)
			}
			fmt.Fprintf(os.Stderr, "Removing gs://%s/%s#%d\n", bucket, obj.Name, obj.Generation)
		}
		return nil
	}

	if err := svc.Objects.Delete(bucket, object).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting object: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Removing gs://%s/%s\n", bucket, object)
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

// --- buckets describe (#163) ---

func runStorageBucketsDescribe(cmd *cobra.Command, args []string) error {
	bucketName := args[0]
	bucketName = strings.TrimPrefix(bucketName, "gs://")
	bucketName = strings.TrimSuffix(bucketName, "/")

	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}

	bucket, err := svc.Buckets.Get(bucketName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing bucket: %w", err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(bucket)
}

// --- mv (#164) ---

func runStorageMv(cmd *cobra.Command, args []string) error {
	src, dst := args[0], args[1]

	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}

	switch {
	case isGCSPath(src) && isGCSPath(dst):
		return storageMvGCS(ctx, svc, src, dst)
	case isGCSPath(src) && !isGCSPath(dst):
		// Download then delete source.
		if err := storageDownload(ctx, svc, src, dst); err != nil {
			return err
		}
		return storageDeleteObject(ctx, svc, src)
	case !isGCSPath(src) && isGCSPath(dst):
		// Upload then delete local source.
		if err := storageUpload(ctx, svc, src, dst); err != nil {
			return err
		}
		return os.Remove(src)
	default:
		return fmt.Errorf("at least one of source or destination must be a gs:// path")
	}
}

func storageMvGCS(ctx context.Context, svc *storage.Service, src, dst string) error {
	srcBucket, srcObject, err := parseGCSPath(src)
	if err != nil {
		return err
	}
	dstBucket, dstObject, err := parseGCSPath(dst)
	if err != nil {
		return err
	}

	if flagStorageMvRecurse {
		prefix := srcObject
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		dstPrefix := dstObject
		if dstPrefix != "" && !strings.HasSuffix(dstPrefix, "/") {
			dstPrefix += "/"
		}
		pageToken := ""
		for {
			call := svc.Objects.List(srcBucket).Prefix(prefix).Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing objects: %w", err)
			}
			for _, obj := range resp.Items {
				if strings.HasSuffix(obj.Name, "/") {
					continue
				}
				relPath := strings.TrimPrefix(obj.Name, prefix)
				newName := dstPrefix + relPath
				if _, err := svc.Objects.Copy(srcBucket, obj.Name, dstBucket, newName, nil).Context(ctx).Do(); err != nil {
					return fmt.Errorf("copying gs://%s/%s: %w", srcBucket, obj.Name, err)
				}
				if err := svc.Objects.Delete(srcBucket, obj.Name).Context(ctx).Do(); err != nil {
					return fmt.Errorf("deleting gs://%s/%s: %w", srcBucket, obj.Name, err)
				}
				fmt.Printf("Moved gs://%s/%s to gs://%s/%s\n", srcBucket, obj.Name, dstBucket, newName)
			}
			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}
		return nil
	}

	if dstObject == "" || strings.HasSuffix(dstObject, "/") {
		dstObject += filepath.Base(srcObject)
	}

	if _, err := svc.Objects.Copy(srcBucket, srcObject, dstBucket, dstObject, nil).Context(ctx).Do(); err != nil {
		return fmt.Errorf("copying object: %w", err)
	}
	if err := svc.Objects.Delete(srcBucket, srcObject).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting source: %w", err)
	}
	fmt.Printf("Moved gs://%s/%s to gs://%s/%s\n", srcBucket, srcObject, dstBucket, dstObject)
	return nil
}

func storageDeleteObject(ctx context.Context, svc *storage.Service, gcsPath string) error {
	bucket, object, err := parseGCSPath(gcsPath)
	if err != nil {
		return err
	}
	return svc.Objects.Delete(bucket, object).Context(ctx).Do()
}

// --- rsync (#165) ---

func runStorageRsync(cmd *cobra.Command, args []string) error {
	src, dst := args[0], args[1]

	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}

	// Build source file map.
	srcFiles, err := listFiles(ctx, svc, src, flagRsyncRecurse)
	if err != nil {
		return fmt.Errorf("listing source: %w", err)
	}

	// Build destination file map.
	dstFiles, err := listFiles(ctx, svc, dst, flagRsyncRecurse)
	if err != nil {
		return fmt.Errorf("listing destination: %w", err)
	}

	// Copy files that are new or different.
	for relPath := range srcFiles {
		if _, exists := dstFiles[relPath]; exists {
			continue
		}
		srcPath := joinPath(src, relPath)
		dstPath := joinPath(dst, relPath)
		if flagRsyncDryRun {
			fmt.Printf("Would copy %s to %s\n", srcPath, dstPath)
			continue
		}
		// Use storage cp logic.
		if isGCSPath(src) && isGCSPath(dst) {
			srcBucket, srcObj, _ := parseGCSPath(srcPath)
			dstBucket, dstObj, _ := parseGCSPath(dstPath)
			if _, err := svc.Objects.Copy(srcBucket, srcObj, dstBucket, dstObj, nil).Context(ctx).Do(); err != nil {
				return fmt.Errorf("copying %s: %w", srcPath, err)
			}
		} else if isGCSPath(src) {
			if err := storageDownloadFile(ctx, svc, mustBucket(srcPath), mustObject(srcPath), dstPath); err != nil {
				return err
			}
		} else {
			if err := storageUploadFile(ctx, svc, srcPath, mustBucket(dstPath), mustObject(dstPath)); err != nil {
				return err
			}
		}
		fmt.Printf("Copied %s to %s\n", srcPath, dstPath)
	}

	// Delete unmatched destination objects.
	if flagRsyncDeleteUnmatched {
		for relPath := range dstFiles {
			if _, exists := srcFiles[relPath]; exists {
				continue
			}
			dstPath := joinPath(dst, relPath)
			if flagRsyncDryRun {
				fmt.Printf("Would delete %s\n", dstPath)
				continue
			}
			if isGCSPath(dst) {
				bucket, object, _ := parseGCSPath(dstPath)
				if err := svc.Objects.Delete(bucket, object).Context(ctx).Do(); err != nil {
					return fmt.Errorf("deleting %s: %w", dstPath, err)
				}
			} else {
				if err := os.Remove(dstPath); err != nil {
					return fmt.Errorf("deleting %s: %w", dstPath, err)
				}
			}
			fmt.Printf("Deleted %s\n", dstPath)
		}
	}

	return nil
}

func listFiles(ctx context.Context, svc *storage.Service, path string, recursive bool) (map[string]bool, error) {
	files := make(map[string]bool)
	if isGCSPath(path) {
		bucket, prefix, err := parseGCSPath(path)
		if err != nil {
			return nil, err
		}
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		pageToken := ""
		for {
			call := svc.Objects.List(bucket).Prefix(prefix).Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return nil, err
			}
			for _, obj := range resp.Items {
				if strings.HasSuffix(obj.Name, "/") {
					continue
				}
				relPath := strings.TrimPrefix(obj.Name, prefix)
				files[relPath] = true
			}
			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}
	} else {
		err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				if !recursive && p != path {
					return filepath.SkipDir
				}
				return nil
			}
			relPath, _ := filepath.Rel(path, p)
			files[filepath.ToSlash(relPath)] = true
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

func joinPath(base, rel string) string {
	if isGCSPath(base) {
		base = strings.TrimSuffix(base, "/")
		return base + "/" + rel
	}
	return filepath.Join(base, filepath.FromSlash(rel))
}

func mustBucket(gcsPath string) string {
	b, _, _ := parseGCSPath(gcsPath)
	return b
}

func mustObject(gcsPath string) string {
	_, o, _ := parseGCSPath(gcsPath)
	return o
}

// --- du (#166) ---

func runStorageDu(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	svc, err := gcp.StorageService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var bucket, prefix string
	if len(args) > 0 {
		bucket, prefix, err = parseGCSPath(args[0])
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("GCS path is required (e.g. gs://bucket/prefix)")
	}

	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	var totalSize uint64
	prefixSizes := make(map[string]uint64)
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
		for _, obj := range resp.Items {
			totalSize += obj.Size
			// Group by first-level prefix.
			rel := strings.TrimPrefix(obj.Name, prefix)
			parts := strings.SplitN(rel, "/", 2)
			if len(parts) == 2 {
				prefixSizes[prefix+parts[0]+"/"] += obj.Size
			} else {
				prefixSizes[obj.Name] += obj.Size
			}
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	if flagDuSummarize {
		fmt.Printf("%s  gs://%s/%s\n", formatSize(totalSize, flagDuReadable), bucket, prefix)
		return nil
	}

	for p, size := range prefixSizes {
		fmt.Printf("%s  gs://%s/%s\n", formatSize(size, flagDuReadable), bucket, p)
	}
	fmt.Printf("%s  total\n", formatSize(totalSize, flagDuReadable))
	return nil
}

func formatSize(size uint64, human bool) string {
	if !human {
		return fmt.Sprintf("%d", size)
	}
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
		tb = 1024 * gb
	)
	switch {
	case size >= tb:
		return fmt.Sprintf("%.1f TiB", float64(size)/float64(tb))
	case size >= gb:
		return fmt.Sprintf("%.1f GiB", float64(size)/float64(gb))
	case size >= mb:
		return fmt.Sprintf("%.1f MiB", float64(size)/float64(mb))
	case size >= kb:
		return fmt.Sprintf("%.1f KiB", float64(size)/float64(kb))
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// --- hash (#167) ---

func runStorageHash(cmd *cobra.Command, args []string) error {
	path := args[0]

	if isGCSPath(path) {
		ctx := context.Background()
		svc, err := gcp.StorageService(ctx, flagAccount)
		if err != nil {
			return err
		}
		bucket, object, err := parseGCSPath(path)
		if err != nil {
			return err
		}
		obj, err := svc.Objects.Get(bucket, object).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("getting object metadata: %w", err)
		}
		fmt.Printf("Hash (crc32c): %s\n", obj.Crc32c)
		fmt.Printf("Hash (md5):    %s\n", obj.Md5Hash)
		return nil
	}

	// Local file.
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer f.Close()

	crcHash := crc32.New(crc32.MakeTable(crc32.Castagnoli))
	md5Hash := md5.New()
	w := io.MultiWriter(crcHash, md5Hash)

	if _, err := io.Copy(w, f); err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	fmt.Printf("Hash (crc32c): %08x\n", crcHash.Sum32())
	fmt.Printf("Hash (md5):    %s\n", hex.EncodeToString(md5Hash.Sum(nil)))
	return nil
}
