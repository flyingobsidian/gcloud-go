package cmd

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
	"google.golang.org/api/googleapi"
)

// --- gcloud artifacts files upload (#1909, gcloud-python 584.0.0) ---

func (s *artifactsCmds) initArtifactsFilesUpload() {
	c := s.artifactsFilesUploadCmd
	c.Flags().StringVar(&s.flagArtFilesLocation, "location", "", "Repository location (required)")
	c.Flags().StringVar(&s.flagArtFilesRepository, "repository", "", "Repository name (required)")
	c.Flags().StringVar(&s.flagArtFilesFormat, "format", "", "Output format")
	_ = c.MarkFlagRequired("location")
	_ = c.MarkFlagRequired("repository")
	c.Flags().StringVar(&s.flagArtFilesUpFile, "file", "",
		"The name under which the file will be uploaded. If not specified, the name of the local file is used.")
	c.Flags().StringVar(&s.flagArtFilesUpFileType, "file-type", "attachment",
		"The type of the file to upload: attachment or artifact")
	c.Flags().StringVar(&s.flagArtFilesUpSource, "source", "", "The path to the file you are uploading")
	c.Flags().StringVar(&s.flagArtFilesUpSourceDir, "source-directory", "", "The directory you are uploading")
	c.MarkFlagsMutuallyExclusive("source", "source-directory")
	c.MarkFlagsOneRequired("source", "source-directory")
	c.Flags().BoolVar(&s.flagArtFilesUpSkipExisting, "skip-existing", false,
		"If specified, skip uploading files that already exist in the repository, and continue to upload the remaining files.")
	c.Flags().BoolVar(&s.flagArtFilesUpAsync, "async", false, "Return immediately, without waiting for the operation in progress to complete.")
	s.artifactsFilesCmd.AddCommand(c)
}

// artFilesUploadType maps --file-type onto UploadFileRequest.fileType, as
// arg_utils.ChoiceToEnum does (case-insensitive, dashes to underscores).
func artFilesUploadType(v string) (string, error) {
	t := strings.ToUpper(strings.ReplaceAll(v, "-", "_"))
	switch t {
	case "ATTACHMENT", "ARTIFACT":
		return t, nil
	}
	return "", fmt.Errorf("Invalid choice: '%s'. Valid choices are: [artifact, attachment].", v)
}

// artMimeType mirrors gcloud-python's util.GetMimetype.
func artMimeType(path string) string {
	if t := mime.TypeByExtension(filepath.Ext(path)); t != "" {
		return t
	}
	return "application/octet-stream"
}

func (s *artifactsCmds) runArtifactsFilesUpload(cmd *cobra.Command, args []string) error {
	fileType, err := artFilesUploadType(s.flagArtFilesUpFileType)
	if err != nil {
		return err
	}
	if s.flagArtFilesUpSource != "" && s.flagArtFilesUpSkipExisting {
		return errors.New("Skip existing is not supported for single file uploads.")
	}
	if s.flagArtFilesUpSourceDir != "" && s.flagArtFilesUpAsync {
		return errors.New("Asynchronous uploads not supported for directories.")
	}
	if s.flagArtFilesUpSourceDir != "" && s.flagArtFilesUpFile != "" {
		return errors.New("File name is not supported for directory uploads.")
	}
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	parent := artRepoParent(project, s.flagArtFilesLocation, s.flagArtFilesRepository)

	if s.flagArtFilesUpSource != "" {
		op, err := s.artUploadFile(ctx, svc, parent, s.flagArtFilesUpSource, s.flagArtFilesUpFile, fileType)
		if err != nil {
			return err
		}
		if s.flagArtFilesUpAsync {
			return emitFormattedTo(s.inv.stdout, op, s.flagArtFilesFormat)
		}
		final, err := waitForArtifactRegistryOperation(ctx, svc, op)
		if err != nil {
			return err
		}
		if final.Response != nil {
			return emitFormattedTo(s.inv.stdout, final.Response, s.flagArtFilesFormat)
		}
		return nil
	}

	dir := filepath.Clean(expandHome(s.flagArtFilesUpSourceDir))
	if st, err := os.Stat(dir); err != nil || !st.IsDir() {
		return errors.New("Specified path is not an existing directory.")
	}
	fmt.Fprintf(s.inv.stderr, "Uploading directory: %s\n", s.flagArtFilesUpSourceDir)
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		op, err := s.artUploadFile(ctx, svc, parent, path, "", fileType)
		if err == nil {
			_, err = waitForArtifactRegistryOperation(ctx, svc, op)
		}
		if err != nil && s.flagArtFilesUpSkipExisting && artAlreadyExists(err) {
			fmt.Fprintln(s.inv.stderr, "WARNING: File with the same ID already exists.")
			return nil
		}
		return err
	})
}

// expandHome expands a leading "~" to the user's home directory, like
// Python's os.path.expanduser.
func expandHome(p string) string {
	if p != "~" && !strings.HasPrefix(p, "~/") {
		return p
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return p
	}
	return filepath.Join(home, strings.TrimPrefix(p, "~"))
}

// artAlreadyExists reports whether err is an "already exists" failure from
// the upload call or its operation.
func artAlreadyExists(err error) bool {
	var gerr *googleapi.Error
	if errors.As(err, &gerr) && gerr.Code == 409 {
		return true
	}
	return strings.Contains(err.Error(), "already exists")
}

// artUploadFile starts a media upload of path into the repository parent,
// named fileID (or the local file's base name) and returns its operation.
func (s *artifactsCmds) artUploadFile(ctx context.Context, svc *artifactregistry.Service, parent, path, fileID, fileType string) (*artifactregistry.Operation, error) {
	if fileID == "" {
		fileID = filepath.Base(path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening source: %w", err)
	}
	defer f.Close()
	fmt.Fprintf(s.inv.stderr, "Uploading file: %s\n", fileID)
	resp, err := svc.Projects.Locations.Repositories.Files.
		Upload(parent, &artifactregistry.UploadFileRequest{FileId: fileID, FileType: fileType}).
		Media(f, googleapi.ContentType(artMimeType(path))).
		Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("uploading file: %w", err)
	}
	if resp.Operation == nil {
		return nil, fmt.Errorf("uploading file %s: no operation returned", fileID)
	}
	return resp.Operation, nil
}
