package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai models (#1458) ---

func (s *aiCmds) initAiModels() {
	all := []*cobra.Command{
		s.aiMDeleteCmd, s.aiMDescribeCmd, s.aiMListCmd,
		s.aiMCopyCmd, s.aiMUploadCmd,
		s.aiMListVersionsCmd, s.aiMDeleteVersionCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagAIMRegion, "region", "", "Region where the model lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagAIMFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.aiMCopyCmd, s.aiMUploadCmd} {
		c.Flags().StringVar(&s.flagAIMConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.aiMListCmd.Flags().StringVar(&s.flagAIMFilter, "filter", "", "Server-side filter expression")
	s.aiMListCmd.Flags().StringVar(&s.flagAIMOrderBy, "order-by", "", "Order-by expression")
	s.aiMListCmd.Flags().Int64Var(&s.flagAIMPageSize, "page-size", 0, "Maximum results per page")
	s.aiMListCmd.Flags().StringVar(&s.flagAIMReadMask, "read-mask", "", "Field mask for reads")
	s.aiMListVersionsCmd.Flags().StringVar(&s.flagAIMFilter, "filter", "", "Server-side filter expression")
	s.aiMListVersionsCmd.Flags().StringVar(&s.flagAIMOrderBy, "order-by", "", "Order-by expression")
	s.aiMListVersionsCmd.Flags().Int64Var(&s.flagAIMPageSize, "page-size", 0, "Maximum results per page")
	s.aiMListVersionsCmd.Flags().StringVar(&s.flagAIMReadMask, "read-mask", "", "Field mask for reads")
	s.aiMDeleteVersionCmd.Flags().StringVar(&s.flagAIMVersion, "version", "",
		"Model version ID (may also be supplied inline as MODEL@VERSION)")

	s.aiModelsCmd.AddCommand(all...)
	s.aiCmd.AddCommand(s.aiModelsCmd)
}

func (s *aiCmds) aiMParent() (string, error) { return s.aiParent(s.flagAIMRegion) }

func (s *aiCmds) aiMName(id string) (string, error) {
	parent, err := s.aiMParent()
	if err != nil {
		return "", err
	}
	return aiChild("models", id, parent), nil
}

// aiMVersionedName resolves MODEL[@VERSION] plus --version into a full
// "projects/.../models/<id>@<version>" resource name. Either an inline
// "@VERSION" suffix on MODEL or a --version flag is required; specifying both
// values with a conflict is rejected.
func (s *aiCmds) aiMVersionedName(rawModel, versionFlag string) (string, error) {
	base := rawModel
	inline := ""
	if i := strings.Index(rawModel, "@"); i >= 0 {
		base = rawModel[:i]
		inline = rawModel[i+1:]
	}
	version := inline
	if versionFlag != "" {
		if inline != "" && inline != versionFlag {
			return "", fmt.Errorf("model version specified twice: %q inline and %q via --version", inline, versionFlag)
		}
		version = versionFlag
	}
	if version == "" {
		return "", fmt.Errorf("model version required (use MODEL@VERSION or --version)")
	}
	parent, err := s.aiMParent()
	if err != nil {
		return "", err
	}
	full := aiChild("models", base, parent)
	return full + "@" + version, nil
}

func (s *aiCmds) runAIMDelete(cmd *cobra.Command, args []string) error {
	name, err := s.aiMName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIMRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Models.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting model: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Delete request issued for model [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIMFormat)
}

func (s *aiCmds) runAIMDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.aiMName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIMRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Models.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing model: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAIMFormat)
}

func (s *aiCmds) runAIMList(cmd *cobra.Command, args []string) error {
	parent, err := s.aiMParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIMRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1Model
	pageToken := ""
	for {
		call := svc.Projects.Locations.Models.List(parent).Context(ctx)
		if s.flagAIMFilter != "" {
			call = call.Filter(s.flagAIMFilter)
		}
		if s.flagAIMOrderBy != "" {
			call = call.OrderBy(s.flagAIMOrderBy)
		}
		if s.flagAIMPageSize > 0 {
			call = call.PageSize(s.flagAIMPageSize)
		}
		if s.flagAIMReadMask != "" {
			call = call.ReadMask(s.flagAIMReadMask)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing models: %w", err)
		}
		all = append(all, resp.Models...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagAIMFormat)
}

func (s *aiCmds) runAIMCopy(cmd *cobra.Command, args []string) error {
	parent, err := s.aiMParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1CopyModelRequest{}
	if err := loadYAMLOrJSONInto(s.flagAIMConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIMRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Models.Copy(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("copying model: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Copy request issued (operation: %s).\n", op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIMFormat)
}

func (s *aiCmds) runAIMUpload(cmd *cobra.Command, args []string) error {
	parent, err := s.aiMParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1UploadModelRequest{}
	if err := loadYAMLOrJSONInto(s.flagAIMConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIMRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Models.Upload(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("uploading model: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Upload request issued (operation: %s).\n", op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIMFormat)
}

func (s *aiCmds) runAIMListVersions(cmd *cobra.Command, args []string) error {
	name, err := s.aiMName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIMRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1Model
	pageToken := ""
	for {
		call := svc.Projects.Locations.Models.ListVersions(name).Context(ctx)
		if s.flagAIMFilter != "" {
			call = call.Filter(s.flagAIMFilter)
		}
		if s.flagAIMOrderBy != "" {
			call = call.OrderBy(s.flagAIMOrderBy)
		}
		if s.flagAIMPageSize > 0 {
			call = call.PageSize(s.flagAIMPageSize)
		}
		if s.flagAIMReadMask != "" {
			call = call.ReadMask(s.flagAIMReadMask)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing model versions: %w", err)
		}
		all = append(all, resp.Models...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagAIMFormat)
}

func (s *aiCmds) runAIMDeleteVersion(cmd *cobra.Command, args []string) error {
	name, err := s.aiMVersionedName(args[0], s.flagAIMVersion)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAIMRegion)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Models.DeleteVersion(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting model version: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Delete-version request issued for [%s] (operation: %s).\n", name, op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagAIMFormat)
}
