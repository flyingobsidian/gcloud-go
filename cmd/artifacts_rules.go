package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts rules (#1080) ---

var artifactsRulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Manage Artifact Registry cleanup rules",
}

var (
	flagArtRuleLocation   string
	flagArtRuleRepository string
	flagArtRuleFormat     string
	flagArtRuleFilter     string
	flagArtRuleConfigFile string
	flagArtRuleMask       string
)

var artifactsRulesCreateCmd = &cobra.Command{
	Use:   "create RULE",
	Short: "Create a cleanup rule",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsRulesCreate,
}

var artifactsRulesDeleteCmd = &cobra.Command{
	Use:   "delete RULE",
	Short: "Delete a cleanup rule",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsRulesDelete,
}

var artifactsRulesDescribeCmd = &cobra.Command{
	Use:   "describe RULE",
	Short: "Describe a cleanup rule",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsRulesDescribe,
}

var artifactsRulesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cleanup rules",
	Args:  cobra.NoArgs,
	RunE:  runArtifactsRulesList,
}

var artifactsRulesUpdateCmd = &cobra.Command{
	Use:   "update RULE",
	Short: "Update a cleanup rule",
	Args:  cobra.ExactArgs(1),
	RunE:  runArtifactsRulesUpdate,
}

func init() {
	for _, c := range []*cobra.Command{
		artifactsRulesCreateCmd, artifactsRulesDeleteCmd,
		artifactsRulesDescribeCmd, artifactsRulesListCmd, artifactsRulesUpdateCmd,
	} {
		c.Flags().StringVar(&flagArtRuleLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&flagArtRuleRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&flagArtRuleFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
	}
	artifactsRulesCreateCmd.Flags().StringVar(&flagArtRuleConfigFile, "config-file", "", "YAML/JSON body for the Rule (required)")
	_ = artifactsRulesCreateCmd.MarkFlagRequired("config-file")
	artifactsRulesUpdateCmd.Flags().StringVar(&flagArtRuleConfigFile, "config-file", "", "YAML/JSON body for the Rule update (required)")
	artifactsRulesUpdateCmd.Flags().StringVar(&flagArtRuleMask, "update-mask", "", "Fields to update")
	_ = artifactsRulesUpdateCmd.MarkFlagRequired("config-file")
	artifactsRulesListCmd.Flags().StringVar(&flagArtRuleFilter, "filter", "", "Filter expression")

	artifactsRulesCmd.AddCommand(
		artifactsRulesCreateCmd, artifactsRulesDeleteCmd,
		artifactsRulesDescribeCmd, artifactsRulesListCmd, artifactsRulesUpdateCmd,
	)
	artifactsCmd.AddCommand(artifactsRulesCmd)
}

func artRuleName(project, location, repo, raw string) string {
	return artFullName(artRepoParent(project, location, repo)+"/rules", raw)
}

func runArtifactsRulesCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &artifactregistry.GoogleDevtoolsArtifactregistryV1Rule{}
	if err := loadYAMLOrJSONInto(flagArtRuleConfigFile, body); err != nil {
		return err
	}
	parent := artRepoParent(project, flagArtRuleLocation, flagArtRuleRepository)
	out, err := svc.Projects.Locations.Repositories.Rules.Create(parent, body).
		RuleId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating rule: %w", err)
	}
	return emitFormatted(out, flagArtRuleFormat)
}

func runArtifactsRulesDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artRuleName(project, flagArtRuleLocation, flagArtRuleRepository, args[0])
	out, err := svc.Projects.Locations.Repositories.Rules.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting rule: %w", err)
	}
	return emitFormatted(out, flagArtRuleFormat)
}

func runArtifactsRulesDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	name := artRuleName(project, flagArtRuleLocation, flagArtRuleRepository, args[0])
	r, err := svc.Projects.Locations.Repositories.Rules.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing rule: %w", err)
	}
	return emitFormatted(r, flagArtRuleFormat)
}

func runArtifactsRulesList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	parent := artRepoParent(project, flagArtRuleLocation, flagArtRuleRepository)
	var all []*artifactregistry.GoogleDevtoolsArtifactregistryV1Rule
	token := ""
	for {
		call := svc.Projects.Locations.Repositories.Rules.List(parent).Context(ctx)
		if token != "" {
			call = call.PageToken(token)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing rules: %w", err)
		}
		all = append(all, resp.Rules...)
		if resp.NextPageToken == "" {
			break
		}
		token = resp.NextPageToken
	}
	return emitFormatted(all, flagArtRuleFormat)
}

func runArtifactsRulesUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.ArtifactRegistryService(ctx, flagAccount)
	if err != nil {
		return err
	}
	body := &artifactregistry.GoogleDevtoolsArtifactregistryV1Rule{}
	if err := loadYAMLOrJSONInto(flagArtRuleConfigFile, body); err != nil {
		return err
	}
	mask := flagArtRuleMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := artRuleName(project, flagArtRuleLocation, flagArtRuleRepository, args[0])
	call := svc.Projects.Locations.Repositories.Rules.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating rule: %w", err)
	}
	return emitFormatted(out, flagArtRuleFormat)
}
