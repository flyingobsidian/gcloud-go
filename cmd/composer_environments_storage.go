package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	storage "google.golang.org/api/storage/v1"
)

// --- gcloud composer environments storage (subgroup of #1502) ---
//
// Python's `gcloud composer environments storage {dags,data,plugins}` mirrors
// files between a local path and the environment's Cloud Storage bucket at
// prefix <bucket>/{dags|data|plugins}/. This implementation talks directly to
// GCS using the environment's config.dagGcsPrefix to derive bucket + prefix.

var composerEnvStorageCmd = &cobra.Command{Use: "storage", Short: "Manage files in a Composer environment's Cloud Storage bucket"}

var (
	composerEnvStorageDagsCmd    = &cobra.Command{Use: "dags", Short: "Manage the DAGs folder"}
	composerEnvStorageDataCmd    = &cobra.Command{Use: "data", Short: "Manage the data folder"}
	composerEnvStoragePluginsCmd = &cobra.Command{Use: "plugins", Short: "Manage the plugins folder"}
)

var (
	flagComposerStEnvironment  string
	flagComposerStSource       string
	flagComposerStDestination  string
	flagComposerStSubdir       string
)

type composerStorageOp struct {
	kind     string // "dags", "data", "plugins"
	parent   *cobra.Command
	flagsSet bool
}

var composerStorageKinds = []*composerStorageOp{
	{kind: "dags", parent: composerEnvStorageDagsCmd},
	{kind: "data", parent: composerEnvStorageDataCmd},
	{kind: "plugins", parent: composerEnvStoragePluginsCmd},
}

func init() {
	for _, k := range composerStorageKinds {
		importCmd := &cobra.Command{
			Use: "import", Short: fmt.Sprintf("Upload local files into the environment's %s folder", k.kind),
			Args: cobra.NoArgs, RunE: composerStorageImportFn(k.kind),
		}
		exportCmd := &cobra.Command{
			Use: "export", Short: fmt.Sprintf("Download files from the environment's %s folder", k.kind),
			Args: cobra.NoArgs, RunE: composerStorageExportFn(k.kind),
		}
		listCmd := &cobra.Command{
			Use: "list", Short: fmt.Sprintf("List files in the environment's %s folder", k.kind),
			Args: cobra.NoArgs, RunE: composerStorageListFn(k.kind),
		}
		deleteCmd := &cobra.Command{
			Use: "delete", Short: fmt.Sprintf("Delete files from the environment's %s folder", k.kind),
			Args: cobra.NoArgs, RunE: composerStorageDeleteFn(k.kind),
		}

		for _, c := range []*cobra.Command{importCmd, exportCmd, listCmd, deleteCmd} {
			c.Flags().StringVar(&flagComposerEnvLocation, "location", "", "Composer location (required)")
			_ = c.MarkFlagRequired("location")
			c.Flags().StringVar(&flagComposerStEnvironment, "environment", "",
				"Composer environment ID (required)")
			_ = c.MarkFlagRequired("environment")
			c.Flags().StringVar(&flagComposerStSubdir, "subdir", "",
				"Optional subdirectory within the folder")
			c.Flags().StringVar(&flagComposerEnvFormat, "format", "", "Output format")
		}
		importCmd.Flags().StringVar(&flagComposerStSource, "source", "",
			"Local file or directory to upload (required)")
		_ = importCmd.MarkFlagRequired("source")
		importCmd.Flags().StringVar(&flagComposerStDestination, "destination", "",
			"Optional destination path within the folder")

		exportCmd.Flags().StringVar(&flagComposerStDestination, "destination", "",
			"Local directory to write downloaded files into (required)")
		_ = exportCmd.MarkFlagRequired("destination")

		k.parent.AddCommand(importCmd, exportCmd, listCmd, deleteCmd)
		composerEnvStorageCmd.AddCommand(k.parent)
	}
}

func composerEnvBucketAndPrefix(kind string) (bucket string, prefix string, err error) {
	name, err := composerEnvResolvedName(flagComposerStEnvironment)
	if err != nil {
		return "", "", err
	}
	ctx := context.Background()
	svc, err := gcp.ComposerService(ctx, flagAccount)
	if err != nil {
		return "", "", err
	}
	env, err := svc.Projects.Locations.Environments.Get(name).Context(ctx).Do()
	if err != nil {
		return "", "", fmt.Errorf("describing environment: %w", err)
	}
	if env.Config == nil || env.Config.DagGcsPrefix == "" {
		return "", "", fmt.Errorf("environment [%s] has no DAG GCS prefix set", flagComposerStEnvironment)
	}
	// dagGcsPrefix format: gs://BUCKET/dags
	raw := strings.TrimPrefix(env.Config.DagGcsPrefix, "gs://")
	slash := strings.Index(raw, "/")
	if slash < 0 {
		return "", "", fmt.Errorf("unexpected dagGcsPrefix format: %s", env.Config.DagGcsPrefix)
	}
	bucket = raw[:slash]
	// substitute the kind for the trailing "dags" segment
	prefix = kind
	if flagComposerStSubdir != "" {
		prefix = prefix + "/" + strings.TrimPrefix(flagComposerStSubdir, "/")
	}
	return bucket, prefix, nil
}

func composerStorageImportFn(kind string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bucket, prefix, err := composerEnvBucketAndPrefix(kind)
		if err != nil {
			return err
		}
		ctx := context.Background()
		st, err := gcp.StorageService(ctx, flagAccount)
		if err != nil {
			return err
		}
		return uploadPathToPrefix(ctx, st, bucket, prefix, flagComposerStSource, flagComposerStDestination)
	}
}

func uploadPathToPrefix(ctx context.Context, st *storage.Service, bucket, prefix, src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("reading %s: %w", src, err)
	}
	if info.IsDir() {
		return filepath.Walk(src, func(p string, fi os.FileInfo, werr error) error {
			if werr != nil {
				return werr
			}
			if fi.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(src, p)
			if err != nil {
				return err
			}
			objName := joinGCS(prefix, dst, rel)
			return uploadOne(ctx, st, bucket, objName, p)
		})
	}
	name := filepath.Base(src)
	objName := joinGCS(prefix, dst, name)
	return uploadOne(ctx, st, bucket, objName, src)
}

func uploadOne(ctx context.Context, st *storage.Service, bucket, object, localPath string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("opening %s: %w", localPath, err)
	}
	defer f.Close()
	if _, err := st.Objects.Insert(bucket, &storage.Object{Name: object}).Media(f).Context(ctx).Do(); err != nil {
		return fmt.Errorf("uploading %s to gs://%s/%s: %w", localPath, bucket, object, err)
	}
	fmt.Printf("Uploaded %s -> gs://%s/%s\n", localPath, bucket, object)
	return nil
}

func joinGCS(parts ...string) string {
	out := ""
	for _, p := range parts {
		if p == "" {
			continue
		}
		p = strings.Trim(p, "/")
		if out == "" {
			out = p
		} else {
			out = out + "/" + p
		}
	}
	return out
}

func composerStorageExportFn(kind string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bucket, prefix, err := composerEnvBucketAndPrefix(kind)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(flagComposerStDestination, 0755); err != nil {
			return fmt.Errorf("creating destination: %w", err)
		}
		ctx := context.Background()
		st, err := gcp.StorageService(ctx, flagAccount)
		if err != nil {
			return err
		}
		pageToken := ""
		for {
			call := st.Objects.List(bucket).Prefix(prefix + "/").Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing bucket: %w", err)
			}
			for _, o := range resp.Items {
				rel := strings.TrimPrefix(o.Name, prefix+"/")
				if rel == "" {
					continue
				}
				target := filepath.Join(flagComposerStDestination, rel)
				if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
					return err
				}
				resp, err := st.Objects.Get(bucket, o.Name).Context(ctx).Download()
				if err != nil {
					return fmt.Errorf("downloading %s: %w", o.Name, err)
				}
				func() {
					defer resp.Body.Close()
					f, ferr := os.Create(target)
					if ferr != nil {
						err = ferr
						return
					}
					defer f.Close()
					if _, werr := io.Copy(f, resp.Body); werr != nil {
						err = werr
					}
				}()
				if err != nil {
					return fmt.Errorf("writing %s: %w", target, err)
				}
				fmt.Printf("Downloaded gs://%s/%s -> %s\n", bucket, o.Name, target)
			}
			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}
		return nil
	}
}

func composerStorageListFn(kind string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bucket, prefix, err := composerEnvBucketAndPrefix(kind)
		if err != nil {
			return err
		}
		ctx := context.Background()
		st, err := gcp.StorageService(ctx, flagAccount)
		if err != nil {
			return err
		}
		var items []string
		pageToken := ""
		for {
			call := st.Objects.List(bucket).Prefix(prefix + "/").Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing bucket: %w", err)
			}
			for _, o := range resp.Items {
				items = append(items, fmt.Sprintf("gs://%s/%s", bucket, o.Name))
			}
			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}
		if flagComposerEnvFormat != "" {
			return emitFormatted(items, flagComposerEnvFormat)
		}
		for _, item := range items {
			fmt.Println(item)
		}
		return nil
	}
}

func composerStorageDeleteFn(kind string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bucket, prefix, err := composerEnvBucketAndPrefix(kind)
		if err != nil {
			return err
		}
		ctx := context.Background()
		st, err := gcp.StorageService(ctx, flagAccount)
		if err != nil {
			return err
		}
		pageToken := ""
		count := 0
		for {
			call := st.Objects.List(bucket).Prefix(prefix + "/").Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			resp, err := call.Do()
			if err != nil {
				return fmt.Errorf("listing bucket: %w", err)
			}
			for _, o := range resp.Items {
				if err := st.Objects.Delete(bucket, o.Name).Context(ctx).Do(); err != nil {
					return fmt.Errorf("deleting %s: %w", o.Name, err)
				}
				fmt.Printf("Deleted gs://%s/%s\n", bucket, o.Name)
				count++
			}
			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}
		fmt.Printf("Deleted %d objects.\n", count)
		return nil
	}
}
