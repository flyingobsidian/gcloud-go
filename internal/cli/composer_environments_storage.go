package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	storage "google.golang.org/api/storage/v1"
)

// --- gcloud composer environments storage (subgroup of #1502) ---
//
// Python's `gcloud composer environments storage {dags,data,plugins}` mirrors
// files between a local path and the environment's Cloud Storage bucket at
// prefix <bucket>/{dags|data|plugins}/. This implementation talks directly to
// GCS using the environment's config.dagGcsPrefix to derive bucket + prefix.

type composerStorageOp struct {
	kind     string // "dags", "data", "plugins"
	parent   *cobra.Command
	flagsSet bool
}

func (s *composerCmds) initComposerEnvironmentsStorage() {
	for _, k := range s.composerStorageKinds {
		importCmd := &cobra.Command{
			Use: "import", Short: fmt.Sprintf("Upload local files into the environment's %s folder", k.kind),
			Args: cobra.NoArgs, RunE: s.composerStorageImportFn(k.kind),
		}
		exportCmd := &cobra.Command{
			Use: "export", Short: fmt.Sprintf("Download files from the environment's %s folder", k.kind),
			Args: cobra.NoArgs, RunE: s.composerStorageExportFn(k.kind),
		}
		listCmd := &cobra.Command{
			Use: "list", Short: fmt.Sprintf("List files in the environment's %s folder", k.kind),
			Args: cobra.NoArgs, RunE: s.composerStorageListFn(k.kind),
		}
		deleteCmd := &cobra.Command{
			Use: "delete", Short: fmt.Sprintf("Delete files from the environment's %s folder", k.kind),
			Args: cobra.NoArgs, RunE: s.composerStorageDeleteFn(k.kind),
		}

		for _, c := range []*cobra.Command{importCmd, exportCmd, listCmd, deleteCmd} {
			c.Flags().StringVar(&s.flagComposerEnvLocation, "location", "", "Composer location (required)")
			_ = c.MarkFlagRequired("location")
			c.Flags().StringVar(&s.flagComposerStEnvironment, "environment", "",
				"Composer environment ID (required)")
			_ = c.MarkFlagRequired("environment")
			c.Flags().StringVar(&s.flagComposerStSubdir, "subdir", "",
				"Optional subdirectory within the folder")
			c.Flags().StringVar(&s.flagComposerEnvFormat, "format", "", "Output format")
		}
		importCmd.Flags().StringVar(&s.flagComposerStSource, "source", "",
			"Local file or directory to upload (required)")
		_ = importCmd.MarkFlagRequired("source")
		importCmd.Flags().StringVar(&s.flagComposerStDestination, "destination", "",
			"Optional destination path within the folder")

		exportCmd.Flags().StringVar(&s.flagComposerStDestination, "destination", "",
			"Local directory to write downloaded files into (required)")
		_ = exportCmd.MarkFlagRequired("destination")

		k.parent.AddCommand(importCmd, exportCmd, listCmd, deleteCmd)
		s.composerEnvStorageCmd.AddCommand(k.parent)
	}
}

func (s *composerCmds) composerEnvBucketAndPrefix(kind string) (bucket string, prefix string, err error) {
	name, err := s.composerEnvResolvedName(s.flagComposerStEnvironment)
	if err != nil {
		return "", "", err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().Composer(ctx)
	if err != nil {
		return "", "", err
	}
	env, err := svc.Projects.Locations.Environments.Get(name).Context(ctx).Do()
	if err != nil {
		return "", "", fmt.Errorf("describing environment: %w", err)
	}
	if env.Config == nil || env.Config.DagGcsPrefix == "" {
		return "", "", fmt.Errorf("environment [%s] has no DAG GCS prefix set", s.flagComposerStEnvironment)
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
	if s.flagComposerStSubdir != "" {
		prefix = prefix + "/" + strings.TrimPrefix(s.flagComposerStSubdir, "/")
	}
	return bucket, prefix, nil
}

func (s *composerCmds) composerStorageImportFn(kind string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bucket, prefix, err := s.composerEnvBucketAndPrefix(kind)
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		st, err := s.inv.clients().Storage(ctx)
		if err != nil {
			return err
		}
		return s.uploadPathToPrefix(ctx, st, bucket, prefix, s.flagComposerStSource, s.flagComposerStDestination)
	}
}

func (s *composerCmds) uploadPathToPrefix(ctx context.Context, st *storage.Service, bucket, prefix, src, dst string) error {
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
			return s.uploadOne(ctx, st, bucket, objName, p)
		})
	}
	name := filepath.Base(src)
	objName := joinGCS(prefix, dst, name)
	return s.uploadOne(ctx, st, bucket, objName, src)
}

func (s *composerCmds) uploadOne(ctx context.Context, st *storage.Service, bucket, object, localPath string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("opening %s: %w", localPath, err)
	}
	defer f.Close()
	if _, err := st.Objects.Insert(bucket, &storage.Object{Name: object}).Media(f).Context(ctx).Do(); err != nil {
		return fmt.Errorf("uploading %s to gs://%s/%s: %w", localPath, bucket, object, err)
	}
	fmt.Fprintf(s.inv.stdout, "Uploaded %s -> gs://%s/%s\n", localPath, bucket, object)
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

// realPath resolves p like Python's os.path.realpath: the path is made
// absolute and symlinks are resolved for the longest existing prefix, with
// any non-existent remainder appended unchanged.
func realPath(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return filepath.Clean(p)
	}
	rest := ""
	for cur := abs; ; {
		if resolved, err := filepath.EvalSymlinks(cur); err == nil {
			return filepath.Join(resolved, rest)
		}
		parent := filepath.Dir(cur)
		if parent == cur {
			return abs
		}
		rest = filepath.Join(filepath.Base(cur), rest)
		cur = parent
	}
}

// composerPathInsideDir reports whether path resolves strictly inside dir,
// mirroring gcloud-python's storage_util._IsPathInsideDirectory.
func composerPathInsideDir(path, dir string) bool {
	realDir, real := realPath(dir), realPath(path)
	if real == realDir {
		return false
	}
	return strings.HasPrefix(real, strings.TrimSuffix(realDir, string(filepath.Separator))+string(filepath.Separator))
}

func (s *composerCmds) composerStorageExportFn(kind string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bucket, prefix, err := s.composerEnvBucketAndPrefix(kind)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(s.flagComposerStDestination, 0755); err != nil {
			return fmt.Errorf("creating destination: %w", err)
		}
		ctx := cmd.Context()
		st, err := s.inv.clients().Storage(ctx)
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
				target := filepath.Join(s.flagComposerStDestination, rel)
				// Object names are controlled by whoever can write to the
				// bucket; skip any that would land outside --destination
				// via '..' segments or existing symlinks (gcloud-python 589.0.0).
				if !composerPathInsideDir(target, s.flagComposerStDestination) {
					fmt.Fprintf(s.inv.stderr, "WARNING: Skipping export of [gs://%s/%s] because it would be written outside of the destination directory [%s].\n",
						bucket, o.Name, s.flagComposerStDestination)
					continue
				}
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
				fmt.Fprintf(s.inv.stdout, "Downloaded gs://%s/%s -> %s\n", bucket, o.Name, target)
			}
			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}
		return nil
	}
}

func (s *composerCmds) composerStorageListFn(kind string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bucket, prefix, err := s.composerEnvBucketAndPrefix(kind)
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		st, err := s.inv.clients().Storage(ctx)
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
		if s.flagComposerEnvFormat != "" {
			return emitFormattedTo(s.inv.stdout, items, s.flagComposerEnvFormat)
		}
		for _, item := range items {
			fmt.Fprintln(s.inv.stdout, item)
		}
		return nil
	}
}

func (s *composerCmds) composerStorageDeleteFn(kind string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		bucket, prefix, err := s.composerEnvBucketAndPrefix(kind)
		if err != nil {
			return err
		}
		ctx := cmd.Context()
		st, err := s.inv.clients().Storage(ctx)
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
				fmt.Fprintf(s.inv.stdout, "Deleted gs://%s/%s\n", bucket, o.Name)
				count++
			}
			if resp.NextPageToken == "" {
				break
			}
			pageToken = resp.NextPageToken
		}
		fmt.Fprintf(s.inv.stdout, "Deleted %d objects.\n", count)
		return nil
	}
}
