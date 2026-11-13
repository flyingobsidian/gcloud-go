package cmd

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	storage "google.golang.org/api/storage/v1"
)

// --- buckets list ---

// --- cp ---

// --- cat ---

// --- rm ---

// --- buckets describe ---

// --- mv ---

// --- rsync ---

// --- du ---

// --- hash ---

// --- cp extra flags ---

// --- buckets list extra flags ---

// --- ls ---

func (s *storageCmds) initStorage() {
	s.storageBucketsListCmd.Flags().StringVar(&s.flagStorageBucketsFormat, "format", "", "Output format (e.g. json)")
	s.storageBucketsListCmd.Flags().BoolVar(&s.flagStorageBucketsSoftDeleted, "soft-deleted", false, "Include soft-deleted buckets")
	s.storageBucketsListCmd.Flags().StringVar(&s.flagBucketsListFilter, "filter", "", "Filter expression")
	s.storageBucketsListCmd.Flags().StringVar(&s.flagBucketsListPrefix, "prefix", "", "Filter buckets by name prefix")
	s.storageBucketsListCmd.Flags().BoolVar(&s.flagBucketsListURI, "uri", false, "Print gs:// URIs")
	s.storageBucketsCmd.AddCommand(s.storageBucketsListCmd)
	s.storageBucketsCmd.AddCommand(s.storageBucketsDescribeCmd)

	s.storageCpCmd.Flags().BoolVarP(&s.flagStorageCpRecurse, "recursive", "r", false, "Copy recursively")
	s.storageCpCmd.Flags().BoolVarP(&s.flagStorageCpNoClobber, "no-clobber", "n", false, "Do not overwrite existing files")
	s.storageCpCmd.Flags().StringVar(&s.flagStorageCpStorageClass, "storage-class", "", "Storage class for uploaded objects")
	s.storageCpCmd.Flags().StringVar(&s.flagStorageCpContentType, "content-type", "", "Content-Type for uploaded objects")
	s.storageCpCmd.Flags().StringVar(&s.flagStorageCpCacheControl, "cache-control", "", "Cache-Control header for uploaded objects")
	s.storageCpCmd.Flags().BoolVar(&s.flagStorageCpContinueOnErr, "continue-on-error", false, "Skip failures and continue")

	s.storageRmCmd.Flags().BoolVarP(&s.flagStorageRmRecurse, "recursive", "r", false, "Delete recursively")
	s.storageRmCmd.Flags().BoolVar(&s.flagStorageRmAllVersions, "all-versions", false, "Delete all object versions")
	s.storageRmCmd.Flags().BoolVar(&s.flagStorageRmContinueOnErr, "continue-on-error", false, "Skip failures and continue")

	s.storageLsCmd.Flags().BoolVarP(&s.flagStorageLsRecurse, "recursive", "r", false, "List recursively")
	s.storageLsCmd.Flags().BoolVarP(&s.flagStorageLsLong, "long", "l", false, "Show size and creation time")
	s.storageLsCmd.Flags().BoolVar(&s.flagStorageLsJSON, "json", false, "Output as JSON")

	s.storageMvCmd.Flags().BoolVarP(&s.flagStorageMvRecurse, "recursive", "r", false, "Move recursively")

	s.storageRsyncCmd.Flags().BoolVarP(&s.flagRsyncRecurse, "recursive", "r", false, "Sync recursively")
	s.storageRsyncCmd.Flags().BoolVarP(&s.flagRsyncDeleteUnmatched, "delete-unmatched-destination-objects", "d", false, "Delete destination objects not in source")
	s.storageRsyncCmd.Flags().BoolVarP(&s.flagRsyncDryRun, "dry-run", "n", false, "Show what would be synced")

	s.storageDuCmd.Flags().BoolVarP(&s.flagDuSummarize, "summarize", "s", false, "Show total only")
	s.storageDuCmd.Flags().BoolVarP(&s.flagDuReadable, "readable-sizes", "r", false, "Human-readable sizes")

	s.storageCmd.AddCommand(s.storageBucketsCmd)
	s.storageCmd.AddCommand(s.storageCatCmd)
	s.storageCmd.AddCommand(s.storageCpCmd)
	s.storageCmd.AddCommand(s.storageLsCmd)
	s.storageCmd.AddCommand(s.storageMvCmd)
	s.storageCmd.AddCommand(s.storageRmCmd)
	s.storageCmd.AddCommand(s.storageRsyncCmd)
	s.storageCmd.AddCommand(s.storageDuCmd)
	s.storageCmd.AddCommand(s.storageHashCmd)

	// gcloud-python storage subcommands/subgroups not yet implemented (#547).
	registerStubCommand(s.storageCmd, "diagnose", "Diagnose Cloud Storage issues")
	registerStubGroup(s.storageCmd, "insights",
		"Manage Storage Insights",
		"describe", "list")
	registerStubGroup(s.storageCmd, "intelligence-configs",
		"Manage Storage Intelligence configurations",
		"describe", "update")
	registerStubGroup(s.storageCmd, "intelligence-findings",
		"Manage Storage Intelligence findings",
		"list", "describe")
	registerStubCommand(s.storageCmd, "sign-url",
		"Generate a signed URL for an object")

}

func (s *storageCmds) runStorageBucketsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}

	var allBuckets []*storage.Bucket
	pageToken := ""
	for {
		call := svc.Buckets.List(project).Context(ctx)
		if s.flagStorageBucketsSoftDeleted {
			call = call.SoftDeleted(true)
		}
		if s.flagBucketsListPrefix != "" {
			call = call.Prefix(s.flagBucketsListPrefix)
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
	if s.flagBucketsListFilter != "" {
		var filtered []*storage.Bucket
		for _, b := range allBuckets {
			if strings.Contains(b.Name, s.flagBucketsListFilter) {
				filtered = append(filtered, b)
			}
		}
		allBuckets = filtered
	}

	if s.flagStorageBucketsFormat == "json" {
		return s.inv.printJSON(allBuckets)
	}

	if s.flagBucketsListURI {
		for _, b := range allBuckets {
			fmt.Fprintf(s.inv.stdout, "gs://%s/\n", b.Name)
		}
		return nil
	}

	for _, b := range allBuckets {
		fmt.Fprintf(s.inv.stdout, "gs://%s/\n", b.Name)
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

func (s *storageCmds) runStorageCp(cmd *cobra.Command, args []string) error {
	src, dst := args[0], args[1]

	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}

	switch {
	case isGCSPath(src) && !isGCSPath(dst):
		// Download: gs://bucket/obj -> local
		return s.storageDownload(ctx, svc, src, dst)
	case !isGCSPath(src) && isGCSPath(dst):
		// Upload: local -> gs://bucket/obj
		return s.storageUpload(ctx, svc, src, dst)
	case isGCSPath(src) && isGCSPath(dst):
		// Copy between GCS paths
		return s.storageGCSCopy(ctx, svc, src, dst)
	default:
		return fmt.Errorf("at least one of source or destination must be a gs:// path")
	}
}

func (s *storageCmds) storageDownload(ctx context.Context, svc *storage.Service, src, dst string) error {
	bucket, object, err := parseGCSPath(src)
	if err != nil {
		return err
	}

	if s.flagStorageCpRecurse {
		return s.storageDownloadRecursive(ctx, svc, bucket, object, dst)
	}

	return s.storageDownloadFile(ctx, svc, bucket, object, dst)
}

func (s *storageCmds) storageDownloadFile(ctx context.Context, svc *storage.Service, bucket, object, dst string) error {
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

	if s.flagStorageCpNoClobber {
		if _, err := os.Stat(dst); err == nil {
			fmt.Fprintf(s.inv.stdout, "Skipping existing file: %s\n", dst)
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

	fmt.Fprintf(s.inv.stdout, "Copied gs://%s/%s to %s (%d bytes)\n", bucket, object, dst, n)
	return nil
}

func (s *storageCmds) storageDownloadRecursive(ctx context.Context, svc *storage.Service, bucket, prefix, dst string) error {
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

		if err := s.storageDownloadFile(ctx, svc, bucket, obj.Name, localPath); err != nil {
			return err
		}
	}
	return nil
}

func (s *storageCmds) storageUpload(ctx context.Context, svc *storage.Service, src, dst string) error {
	bucket, prefix, err := parseGCSPath(dst)
	if err != nil {
		return err
	}

	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat %s: %w", src, err)
	}

	if info.IsDir() {
		if !s.flagStorageCpRecurse {
			return fmt.Errorf("%s is a directory; use --recursive or -r to copy directories", src)
		}
		return s.storageUploadRecursive(ctx, svc, src, bucket, prefix)
	}

	return s.storageUploadFile(ctx, svc, src, bucket, prefix)
}

func (s *storageCmds) storageUploadFile(ctx context.Context, svc *storage.Service, src, bucket, prefix string) error {
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
	if s.flagStorageCpStorageClass != "" {
		obj.StorageClass = s.flagStorageCpStorageClass
	}
	if s.flagStorageCpContentType != "" {
		obj.ContentType = s.flagStorageCpContentType
	}
	if s.flagStorageCpCacheControl != "" {
		obj.CacheControl = s.flagStorageCpCacheControl
	}
	call := svc.Objects.Insert(bucket, obj).Media(f).Context(ctx)
	if s.flagStorageCpNoClobber {
		call = call.IfGenerationMatch(0)
	}
	_, err = call.Do()
	if err != nil {
		return fmt.Errorf("uploading object: %w", err)
	}

	fmt.Fprintf(s.inv.stdout, "Copied %s to gs://%s/%s\n", src, bucket, objectName)
	return nil
}

func (s *storageCmds) storageUploadRecursive(ctx context.Context, svc *storage.Service, srcDir, bucket, prefix string) error {
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

		return s.storageUploadFile(ctx, svc, path, bucket, objectName)
	})
}

func (s *storageCmds) storageGCSCopy(ctx context.Context, svc *storage.Service, src, dst string) error {
	srcBucket, srcObject, err := parseGCSPath(src)
	if err != nil {
		return err
	}
	dstBucket, dstObject, err := parseGCSPath(dst)
	if err != nil {
		return err
	}

	if s.flagStorageCpRecurse {
		return s.storageGCSCopyRecursive(ctx, svc, srcBucket, srcObject, dstBucket, dstObject)
	}

	if dstObject == "" || strings.HasSuffix(dstObject, "/") {
		dstObject += filepath.Base(srcObject)
	}

	_, err = svc.Objects.Copy(srcBucket, srcObject, dstBucket, dstObject, nil).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("copying object: %w", err)
	}

	fmt.Fprintf(s.inv.stdout, "Copied gs://%s/%s to gs://%s/%s\n", srcBucket, srcObject, dstBucket, dstObject)
	return nil
}

func (s *storageCmds) storageGCSCopyRecursive(ctx context.Context, svc *storage.Service, srcBucket, srcPrefix, dstBucket, dstPrefix string) error {
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
		fmt.Fprintf(s.inv.stdout, "Copied gs://%s/%s to gs://%s/%s\n", srcBucket, obj.Name, dstBucket, dstObject)
	}
	return nil
}

func (s *storageCmds) runStorageCat(cmd *cobra.Command, args []string) error {
	bucket, object, err := parseGCSPath(args[0])
	if err != nil {
		return err
	}
	if object == "" {
		return fmt.Errorf("object path is required: gs://bucket/object")
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}

	resp, err := svc.Objects.Get(bucket, object).Context(ctx).Download()
	if err != nil {
		return fmt.Errorf("reading object: %w", err)
	}
	defer resp.Body.Close()

	_, err = io.Copy(s.inv.stdout, resp.Body)
	return err
}

func (s *storageCmds) runStorageRm(cmd *cobra.Command, args []string) error {
	bucket, object, err := parseGCSPath(args[0])
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}

	if s.flagStorageRmRecurse {
		prefix := object
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		pageToken := ""
		var lastErr error
		for {
			call := svc.Objects.List(bucket).Prefix(prefix).Context(ctx)
			if s.flagStorageRmAllVersions {
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
				if s.flagStorageRmAllVersions && obj.Generation != 0 {
					delCall = delCall.Generation(int64(obj.Generation))
				}
				if err := delCall.Do(); err != nil {
					if s.flagStorageRmContinueOnErr {
						fmt.Fprintf(s.inv.stderr, "WARNING: failed to delete gs://%s/%s: %v\n", bucket, obj.Name, err)
						lastErr = err
						continue
					}
					return fmt.Errorf("deleting gs://%s/%s: %w", bucket, obj.Name, err)
				}
				fmt.Fprintf(s.inv.stderr, "Removing gs://%s/%s\n", bucket, obj.Name)
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

	if s.flagStorageRmAllVersions {
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
				if s.flagStorageRmContinueOnErr {
					fmt.Fprintf(s.inv.stderr, "WARNING: failed to delete gs://%s/%s: %v\n", bucket, obj.Name, err)
					continue
				}
				return fmt.Errorf("deleting gs://%s/%s: %w", bucket, obj.Name, err)
			}
			fmt.Fprintf(s.inv.stderr, "Removing gs://%s/%s#%d\n", bucket, obj.Name, obj.Generation)
		}
		return nil
	}

	if err := svc.Objects.Delete(bucket, object).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting object: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Removing gs://%s/%s\n", bucket, object)
	return nil
}

func (s *storageCmds) runStorageLs(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}

	// No args = list buckets.
	if len(args) == 0 {
		project, err := s.inv.project()
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
				fmt.Fprintf(s.inv.stdout, "gs://%s/\n", b.Name)
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
		if !s.flagStorageLsRecurse {
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

	if s.flagStorageLsJSON {
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
		return s.inv.printJSON(entries)
	}

	// Print directories (prefixes).
	for _, p := range allPrefixes {
		fmt.Fprintf(s.inv.stdout, "gs://%s/%s\n", bucket, p)
	}
	// Print objects.
	for _, obj := range allObjects {
		if s.flagStorageLsLong {
			fmt.Fprintf(s.inv.stdout, "%10d  %s  gs://%s/%s\n", obj.Size, obj.TimeCreated, bucket, obj.Name)
		} else {
			fmt.Fprintf(s.inv.stdout, "gs://%s/%s\n", bucket, obj.Name)
		}
	}
	return nil
}

// --- buckets describe (#163) ---

func (s *storageCmds) runStorageBucketsDescribe(cmd *cobra.Command, args []string) error {
	bucketName := args[0]
	bucketName = strings.TrimPrefix(bucketName, "gs://")
	bucketName = strings.TrimSuffix(bucketName, "/")

	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}

	bucket, err := svc.Buckets.Get(bucketName).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing bucket: %w", err)
	}

	return s.inv.printJSON(bucket)
}

// --- mv (#164) ---

func (s *storageCmds) runStorageMv(cmd *cobra.Command, args []string) error {
	src, dst := args[0], args[1]

	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}

	switch {
	case isGCSPath(src) && isGCSPath(dst):
		return s.storageMvGCS(ctx, svc, src, dst)
	case isGCSPath(src) && !isGCSPath(dst):
		// Download then delete source.
		if err := s.storageDownload(ctx, svc, src, dst); err != nil {
			return err
		}
		return storageDeleteObject(ctx, svc, src)
	case !isGCSPath(src) && isGCSPath(dst):
		// Upload then delete local source.
		if err := s.storageUpload(ctx, svc, src, dst); err != nil {
			return err
		}
		return os.Remove(src)
	default:
		return fmt.Errorf("at least one of source or destination must be a gs:// path")
	}
}

func (s *storageCmds) storageMvGCS(ctx context.Context, svc *storage.Service, src, dst string) error {
	srcBucket, srcObject, err := parseGCSPath(src)
	if err != nil {
		return err
	}
	dstBucket, dstObject, err := parseGCSPath(dst)
	if err != nil {
		return err
	}

	if s.flagStorageMvRecurse {
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
				fmt.Fprintf(s.inv.stdout, "Moved gs://%s/%s to gs://%s/%s\n", srcBucket, obj.Name, dstBucket, newName)
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
	fmt.Fprintf(s.inv.stdout, "Moved gs://%s/%s to gs://%s/%s\n", srcBucket, srcObject, dstBucket, dstObject)
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

func (s *storageCmds) runStorageRsync(cmd *cobra.Command, args []string) error {
	src, dst := args[0], args[1]

	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}

	// Build source file map.
	srcFiles, err := listFiles(ctx, svc, src, s.flagRsyncRecurse)
	if err != nil {
		return fmt.Errorf("listing source: %w", err)
	}

	// Build destination file map.
	dstFiles, err := listFiles(ctx, svc, dst, s.flagRsyncRecurse)
	if err != nil {
		return fmt.Errorf("listing destination: %w", err)
	}

	// Copy files that are new or different.
	for relPath := range srcFiles {
		if _, exists := dstFiles[relPath]; exists {
			continue
		}
		if !rsyncSafeRelPath(relPath) {
			fmt.Fprintf(s.inv.stderr, "Skipping %q: path escapes destination directory\n", relPath)
			continue
		}
		srcPath := joinPath(src, relPath)
		dstPath := joinPath(dst, relPath)
		if s.flagRsyncDryRun {
			fmt.Fprintf(s.inv.stdout, "Would copy %s to %s\n", srcPath, dstPath)
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
			if err := s.storageDownloadFile(ctx, svc, mustBucket(srcPath), mustObject(srcPath), dstPath); err != nil {
				return err
			}
		} else {
			if err := s.storageUploadFile(ctx, svc, srcPath, mustBucket(dstPath), mustObject(dstPath)); err != nil {
				return err
			}
		}
		fmt.Fprintf(s.inv.stdout, "Copied %s to %s\n", srcPath, dstPath)
	}

	// Delete unmatched destination objects.
	if s.flagRsyncDeleteUnmatched {
		for relPath := range dstFiles {
			if _, exists := srcFiles[relPath]; exists {
				continue
			}
			dstPath := joinPath(dst, relPath)
			if s.flagRsyncDryRun {
				fmt.Fprintf(s.inv.stdout, "Would delete %s\n", dstPath)
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
			fmt.Fprintf(s.inv.stdout, "Deleted %s\n", dstPath)
		}
	}

	return nil
}

// rsyncSafeRelPath reports whether a relative path derived from the source
// listing stays within the destination directory when joined. Reject absolute
// paths and any component that resolves to `..`, matching gcloud-python's
// 580.0.0 behaviour of skipping files that escape the destination.
func rsyncSafeRelPath(rel string) bool {
	if rel == "" {
		return false
	}
	if strings.HasPrefix(rel, "/") || strings.HasPrefix(rel, `\`) {
		return false
	}
	// Reject rooted paths on Windows too (e.g. `C:\foo`).
	if len(rel) >= 2 && rel[1] == ':' {
		return false
	}
	normalized := strings.ReplaceAll(rel, `\`, "/")
	for _, part := range strings.Split(normalized, "/") {
		if part == ".." {
			return false
		}
	}
	return true
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

func (s *storageCmds) runStorageDu(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
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

	if s.flagDuSummarize {
		fmt.Fprintf(s.inv.stdout, "%s  gs://%s/%s\n", formatSize(totalSize, s.flagDuReadable), bucket, prefix)
		return nil
	}

	for p, size := range prefixSizes {
		fmt.Fprintf(s.inv.stdout, "%s  gs://%s/%s\n", formatSize(size, s.flagDuReadable), bucket, p)
	}
	fmt.Fprintf(s.inv.stdout, "%s  total\n", formatSize(totalSize, s.flagDuReadable))
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

func (s *storageCmds) runStorageHash(cmd *cobra.Command, args []string) error {
	path := args[0]

	if isGCSPath(path) {
		ctx := cmd.Context()
		svc, err := s.inv.clients().Storage(ctx)
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
		fmt.Fprintf(s.inv.stdout, "Hash (crc32c): %s\n", obj.Crc32c)
		fmt.Fprintf(s.inv.stdout, "Hash (md5):    %s\n", obj.Md5Hash)
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

	fmt.Fprintf(s.inv.stdout, "Hash (crc32c): %08x\n", crcHash.Sum32())
	fmt.Fprintf(s.inv.stdout, "Hash (md5):    %s\n", hex.EncodeToString(md5Hash.Sum(nil)))
	return nil
}

// storageCmds holds the commands and flag values of one invocation's "storage"
// command group, so invocations don't share state (#2070).
type storageCmds struct {
	inv *invocation

	storageCmd                    *cobra.Command
	storageBucketsCmd             *cobra.Command
	storageBucketsListCmd         *cobra.Command
	flagStorageBucketsFormat      string
	flagStorageBucketsSoftDeleted bool
	storageCpCmd                  *cobra.Command
	flagStorageCpRecurse          bool
	flagStorageCpNoClobber        bool
	flagStorageCpStorageClass     string
	storageCatCmd                 *cobra.Command
	storageRmCmd                  *cobra.Command
	flagStorageRmRecurse          bool
	flagStorageRmAllVersions      bool
	flagStorageRmContinueOnErr    bool
	storageBucketsDescribeCmd     *cobra.Command
	storageMvCmd                  *cobra.Command
	flagStorageMvRecurse          bool
	storageRsyncCmd               *cobra.Command
	flagRsyncRecurse              bool
	flagRsyncDeleteUnmatched      bool
	flagRsyncDryRun               bool
	storageDuCmd                  *cobra.Command
	flagDuSummarize               bool
	flagDuReadable                bool
	storageHashCmd                *cobra.Command
	flagStorageCpContentType      string
	flagStorageCpCacheControl     string
	flagStorageCpContinueOnErr    bool
	flagBucketsListFilter         string
	flagBucketsListPrefix         string
	flagBucketsListURI            bool
	storageLsCmd                  *cobra.Command
	flagStorageLsRecurse          bool
	flagStorageLsLong             bool
	flagStorageLsJSON             bool
	storageBatchOperationsCmd     *cobra.Command
	storageBatchOpsJobsCmd        *cobra.Command
	storageBatchOpsBucketOpsCmd   *cobra.Command
	flagStBoLocation              string
	flagStBoFormat                string
	flagStBoConfigFile            string
	flagStBoJob                   string
	flagStBoPageSize              int64
	// jobs
	storageBatchOpsJobsCancelCmd   *cobra.Command
	storageBatchOpsJobsCreateCmd   *cobra.Command
	storageBatchOpsJobsDeleteCmd   *cobra.Command
	storageBatchOpsJobsDescribeCmd *cobra.Command
	storageBatchOpsJobsListCmd     *cobra.Command
	// bucket-operations
	storageBatchOpsBucketOpsDescribeCmd *cobra.Command
	storageBatchOpsBucketOpsListCmd     *cobra.Command
	storageFoldersCmd                   *cobra.Command
	flagStFolBucket                     string
	flagStFolFormat                     string
	flagStFolPrefix                     string
	flagStFolPageSize                   int64
	flagStFolDest                       string
	storageFoldersCreateCmd             *cobra.Command
	storageFoldersDeleteCmd             *cobra.Command
	storageFoldersDescribeCmd           *cobra.Command
	storageFoldersListCmd               *cobra.Command
	storageFoldersRenameCmd             *cobra.Command
	storageHmacCmd                      *cobra.Command
	flagStHmacFormat                    string
	flagStHmacServiceAccount            string
	flagStHmacState                     string
	storageHmacCreateCmd                *cobra.Command
	storageHmacDeleteCmd                *cobra.Command
	storageHmacDescribeCmd              *cobra.Command
	storageHmacListCmd                  *cobra.Command
	storageHmacUpdateCmd                *cobra.Command
	storageManagedFoldersCmd            *cobra.Command
	flagStMfBucket                      string
	flagStMfFormat                      string
	flagStMfPageSize                    int64
	flagStMfPrefix                      string
	storageMfCreateCmd                  *cobra.Command
	storageMfDeleteCmd                  *cobra.Command
	storageMfDescribeCmd                *cobra.Command
	storageMfListCmd                    *cobra.Command
	storageMfGetIamCmd                  *cobra.Command
	storageMfSetIamCmd                  *cobra.Command
	storageObjectsCmd                   *cobra.Command
	flagStObjBucket                     string
	flagStObjFormat                     string
	flagStObjConfigFile                 string
	flagStObjPrefix                     string
	flagStObjDelimiter                  string
	flagStObjPageSize                   int64
	storageObjectsDescribeCmd           *cobra.Command
	storageObjectsListCmd               *cobra.Command
	storageObjectsUpdateCmd             *cobra.Command
	storageObjectsComposeCmd            *cobra.Command
	storageOperationsCmd                *cobra.Command
	flagStOpBucket                      string
	flagStOpFormat                      string
	flagStOpFilter                      string
	flagStOpPageSize                    int64
	storageOpCancelCmd                  *cobra.Command
	storageOpDescribeCmd                *cobra.Command
	storageOpListCmd                    *cobra.Command
	flagStRestBucket                    string
	flagStRestGeneration                int64
	flagStRestFormat                    string
	storageRestoreCmd                   *cobra.Command
	storageServiceAgentCmd              *cobra.Command
	flagStSaFormat                      string
	storageSaDescribeCmd                *cobra.Command
}

// newStorageCmd returns the "storage" command group for inv.
func newStorageCmd(inv *invocation) *cobra.Command {
	s := &storageCmds{inv: inv}
	s.storageCmd = &cobra.Command{
		Use:   "storage",
		Short: "Manage Cloud Storage",
	}
	s.storageBucketsCmd = &cobra.Command{
		Use:   "buckets",
		Short: "Manage Cloud Storage buckets",
	}
	s.storageBucketsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Cloud Storage buckets",
		Args:  cobra.NoArgs,
		RunE:  s.runStorageBucketsList,
	}
	s.storageCpCmd = &cobra.Command{
		Use:   "cp SOURCE DESTINATION",
		Short: "Copy files to/from Cloud Storage",
		Long: `Copy files between local filesystem and Cloud Storage.
Examples:
  gcloud storage cp file.txt gs://bucket/path/
  gcloud storage cp gs://bucket/path/file.txt ./local/
  gcloud storage cp -r ./dir gs://bucket/path/`,
		Args: cobra.ExactArgs(2),
		RunE: s.runStorageCp,
	}
	s.storageCatCmd = &cobra.Command{
		Use:   "cat GCS_PATH",
		Short: "Output the contents of a Cloud Storage object to stdout",
		Long: `Output the contents of a Cloud Storage object to stdout.
Examples:
  gcloud storage cat gs://bucket/path/file.txt`,
		Args: cobra.ExactArgs(1),
		RunE: s.runStorageCat,
	}
	s.storageRmCmd = &cobra.Command{
		Use:   "rm GCS_PATH",
		Short: "Delete Cloud Storage objects",
		Long: `Delete objects from Cloud Storage.
Examples:
  gcloud storage rm gs://bucket/path/file.txt
  gcloud storage rm -r gs://bucket/path/`,
		Args: cobra.ExactArgs(1),
		RunE: s.runStorageRm,
	}
	s.storageBucketsDescribeCmd = &cobra.Command{
		Use:   "describe gs://BUCKET",
		Short: "Describe a Cloud Storage bucket",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runStorageBucketsDescribe,
	}
	s.storageMvCmd = &cobra.Command{
		Use:   "mv SOURCE DESTINATION",
		Short: "Move/rename Cloud Storage objects",
		Args:  cobra.ExactArgs(2),
		RunE:  s.runStorageMv,
	}
	s.storageRsyncCmd = &cobra.Command{
		Use:   "rsync SOURCE DESTINATION",
		Short: "Synchronize content of two buckets/directories",
		Args:  cobra.ExactArgs(2),
		RunE:  s.runStorageRsync,
	}
	s.storageDuCmd = &cobra.Command{
		Use:   "du [GCS_PATH]",
		Short: "Display disk usage of Cloud Storage objects",
		Args:  cobra.MaximumNArgs(1),
		RunE:  s.runStorageDu,
	}
	s.storageHashCmd = &cobra.Command{
		Use:   "hash PATH",
		Short: "Compute hashes for local files or Cloud Storage objects",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runStorageHash,
	}
	s.storageLsCmd = &cobra.Command{
		Use:   "ls [GCS_PATH]",
		Short: "List Cloud Storage objects",
		Long: `List objects and buckets in Cloud Storage.
Examples:
  gcloud storage ls gs://bucket
  gcloud storage ls gs://bucket/prefix/`,
		Args: cobra.MaximumNArgs(1),
		RunE: s.runStorageLs,
	}
	s.storageBatchOperationsCmd = &cobra.Command{Use: "batch-operations", Short: "Manage long-running batch storage jobs"}
	s.storageBatchOpsJobsCmd = &cobra.Command{Use: "jobs", Short: "Manage batch storage jobs"}
	s.storageBatchOpsBucketOpsCmd = &cobra.Command{Use: "bucket-operations", Short: "Inspect per-bucket sub-operations of a batch job"}
	s.storageBatchOpsJobsCancelCmd = &cobra.Command{
		Use: "cancel JOB", Short: "Cancel a batch storage job",
		Args: cobra.ExactArgs(1), RunE: s.runStBoJobCancel,
	}
	s.storageBatchOpsJobsCreateCmd = &cobra.Command{
		Use: "create JOB", Short: "Create a batch storage job",
		Args: cobra.ExactArgs(1), RunE: s.runStBoJobCreate,
	}
	s.storageBatchOpsJobsDeleteCmd = &cobra.Command{
		Use: "delete JOB", Short: "Delete a batch storage job",
		Args: cobra.ExactArgs(1), RunE: s.runStBoJobDelete,
	}
	s.storageBatchOpsJobsDescribeCmd = &cobra.Command{
		Use: "describe JOB", Short: "Describe a batch storage job",
		Args: cobra.ExactArgs(1), RunE: s.runStBoJobDescribe,
	}
	s.storageBatchOpsJobsListCmd = &cobra.Command{
		Use: "list", Short: "List batch storage jobs in a location",
		Args: cobra.NoArgs, RunE: s.runStBoJobList,
	}
	s.storageBatchOpsBucketOpsDescribeCmd = &cobra.Command{
		Use: "describe BUCKET_OPERATION", Short: "Describe a per-bucket sub-operation of a batch job",
		Args: cobra.ExactArgs(1), RunE: s.runStBoBucketOpDescribe,
	}
	s.storageBatchOpsBucketOpsListCmd = &cobra.Command{
		Use: "list", Short: "List per-bucket sub-operations of a batch job",
		Args: cobra.NoArgs, RunE: s.runStBoBucketOpList,
	}
	s.storageFoldersCmd = &cobra.Command{Use: "folders", Short: "Manage hierarchical-namespace folders"}
	s.storageFoldersCreateCmd = &cobra.Command{
		Use: "create FOLDER", Short: "Create a folder in a bucket",
		Args: cobra.ExactArgs(1), RunE: s.runStFolCreate,
	}
	s.storageFoldersDeleteCmd = &cobra.Command{
		Use: "delete FOLDER", Short: "Delete a folder in a bucket",
		Args: cobra.ExactArgs(1), RunE: s.runStFolDelete,
	}
	s.storageFoldersDescribeCmd = &cobra.Command{
		Use: "describe FOLDER", Short: "Describe a folder in a bucket",
		Args: cobra.ExactArgs(1), RunE: s.runStFolDescribe,
	}
	s.storageFoldersListCmd = &cobra.Command{
		Use: "list", Short: "List folders in a bucket",
		Args: cobra.NoArgs, RunE: s.runStFolList,
	}
	s.storageFoldersRenameCmd = &cobra.Command{
		Use: "rename SOURCE_FOLDER", Short: "Rename a folder within a bucket",
		Args: cobra.ExactArgs(1), RunE: s.runStFolRename,
	}
	s.storageHmacCmd = &cobra.Command{Use: "hmac", Short: "Manage HMAC keys for service accounts"}
	s.storageHmacCreateCmd = &cobra.Command{
		Use: "create", Short: "Create an HMAC key for a service account (loads request from --service-account)",
		Args: cobra.NoArgs, RunE: s.runStHmacCreate,
	}
	s.storageHmacDeleteCmd = &cobra.Command{
		Use: "delete ACCESS_ID", Short: "Delete an HMAC key",
		Args: cobra.ExactArgs(1), RunE: s.runStHmacDelete,
	}
	s.storageHmacDescribeCmd = &cobra.Command{
		Use: "describe ACCESS_ID", Short: "Describe an HMAC key",
		Args: cobra.ExactArgs(1), RunE: s.runStHmacDescribe,
	}
	s.storageHmacListCmd = &cobra.Command{
		Use: "list", Short: "List HMAC keys in the current project",
		Args: cobra.NoArgs, RunE: s.runStHmacList,
	}
	s.storageHmacUpdateCmd = &cobra.Command{
		Use: "update ACCESS_ID", Short: "Update an HMAC key state (ACTIVE or INACTIVE)",
		Args: cobra.ExactArgs(1), RunE: s.runStHmacUpdate,
	}
	s.storageManagedFoldersCmd = &cobra.Command{Use: "managed-folders", Short: "Manage managed folders inside a bucket"}
	s.storageMfCreateCmd = &cobra.Command{
		Use: "create MANAGED_FOLDER", Short: "Create a managed folder in a bucket",
		Args: cobra.ExactArgs(1), RunE: s.runStMfCreate,
	}
	s.storageMfDeleteCmd = &cobra.Command{
		Use: "delete MANAGED_FOLDER", Short: "Delete a managed folder in a bucket",
		Args: cobra.ExactArgs(1), RunE: s.runStMfDelete,
	}
	s.storageMfDescribeCmd = &cobra.Command{
		Use: "describe MANAGED_FOLDER", Short: "Describe a managed folder in a bucket",
		Args: cobra.ExactArgs(1), RunE: s.runStMfDescribe,
	}
	s.storageMfListCmd = &cobra.Command{
		Use: "list", Short: "List managed folders in a bucket",
		Args: cobra.NoArgs, RunE: s.runStMfList,
	}
	s.storageMfGetIamCmd = &cobra.Command{
		Use: "get-iam-policy MANAGED_FOLDER", Short: "Get the IAM policy for a managed folder",
		Args: cobra.ExactArgs(1), RunE: s.runStMfGetIam,
	}
	s.storageMfSetIamCmd = &cobra.Command{
		Use: "set-iam-policy MANAGED_FOLDER POLICY_FILE", Short: "Set the IAM policy for a managed folder",
		Args: cobra.ExactArgs(2), RunE: s.runStMfSetIam,
	}
	s.storageObjectsCmd = &cobra.Command{Use: "objects", Short: "Manage Cloud Storage objects (metadata operations)"}
	s.storageObjectsDescribeCmd = &cobra.Command{
		Use: "describe OBJECT", Short: "Describe an object (metadata only)",
		Args: cobra.ExactArgs(1), RunE: s.runStObjDescribe,
	}
	s.storageObjectsListCmd = &cobra.Command{
		Use: "list", Short: "List objects in a bucket",
		Args: cobra.NoArgs, RunE: s.runStObjList,
	}
	s.storageObjectsUpdateCmd = &cobra.Command{
		Use: "update OBJECT", Short: "Update object metadata (loads request from --config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runStObjUpdate,
	}
	s.storageObjectsComposeCmd = &cobra.Command{
		Use: "compose DESTINATION_OBJECT", Short: "Compose an object from source objects (loads request from --config-file)",
		Args: cobra.ExactArgs(1), RunE: s.runStObjCompose,
	}
	s.storageOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage bucket-scoped storage long-running operations"}
	s.storageOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a bucket-scoped storage operation",
		Args: cobra.ExactArgs(1), RunE: s.runStOpCancel,
	}
	s.storageOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a bucket-scoped storage operation",
		Args: cobra.ExactArgs(1), RunE: s.runStOpDescribe,
	}
	s.storageOpListCmd = &cobra.Command{
		Use: "list", Short: "List bucket-scoped storage operations",
		Args: cobra.NoArgs, RunE: s.runStOpList,
	}
	s.storageRestoreCmd = &cobra.Command{
		Use:   "restore OBJECT",
		Short: "Restore a soft-deleted object at a specific generation",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runStRestore,
	}
	s.storageServiceAgentCmd = &cobra.Command{Use: "service-agent", Short: "Manage the Cloud Storage service agent"}
	s.storageSaDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the Cloud Storage service agent for the current project",
		Args: cobra.NoArgs, RunE: s.runStSaDescribe,
	}
	s.initStorage()
	s.initStorageBatchOperations()
	s.initStorageFolders()
	s.initStorageHmac()
	s.initStorageManagedFolders()
	s.initStorageObjects()
	s.initStorageOperations()
	s.initStorageRestore()
	s.initStorageServiceAgent()
	return s.storageCmd
}
