package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	artifactregistry "google.golang.org/api/artifactregistry/v1"
)

// --- gcloud artifacts rules (#1080) ---

func (s *artifactsCmds) initArtifactsRules() {
	for _, c := range []*cobra.Command{
		s.artifactsRulesCreateCmd, s.artifactsRulesDeleteCmd,
		s.artifactsRulesDescribeCmd, s.artifactsRulesListCmd, s.artifactsRulesUpdateCmd,
	} {
		c.Flags().StringVar(&s.flagArtRuleLocation, "location", "", "Repository location (required)")
		c.Flags().StringVar(&s.flagArtRuleRepository, "repository", "", "Repository name (required)")
		c.Flags().StringVar(&s.flagArtRuleFormat, "format", "", "Output format")
		_ = c.MarkFlagRequired("location")
		_ = c.MarkFlagRequired("repository")
	}
	s.artifactsRulesCreateCmd.Flags().StringVar(&s.flagArtRuleConfigFile, "config-file", "", "YAML/JSON body for the Rule (required)")
	_ = s.artifactsRulesCreateCmd.MarkFlagRequired("config-file")
	s.artifactsRulesUpdateCmd.Flags().StringVar(&s.flagArtRuleConfigFile, "config-file", "", "YAML/JSON body for the Rule update (required)")
	s.artifactsRulesUpdateCmd.Flags().StringVar(&s.flagArtRuleMask, "update-mask", "", "Fields to update")
	_ = s.artifactsRulesUpdateCmd.MarkFlagRequired("config-file")
	s.artifactsRulesListCmd.Flags().StringVar(&s.flagArtRuleFilter, "filter", "", "Filter expression")

	s.artifactsRulesCmd.AddCommand(
		s.artifactsRulesCreateCmd, s.artifactsRulesDeleteCmd,
		s.artifactsRulesDescribeCmd, s.artifactsRulesListCmd, s.artifactsRulesUpdateCmd,
	)
	s.artifactsCmd.AddCommand(s.artifactsRulesCmd)
}

func artRuleName(project, location, repo, raw string) string {
	return artFullName(artRepoParent(project, location, repo)+"/rules", raw)
}

func (s *artifactsCmds) runArtifactsRulesCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	body := &artifactregistry.GoogleDevtoolsArtifactregistryV1Rule{}
	if err := loadYAMLOrJSONInto(s.flagArtRuleConfigFile, body); err != nil {
		return err
	}
	parent := artRepoParent(project, s.flagArtRuleLocation, s.flagArtRuleRepository)
	out, err := svc.Projects.Locations.Repositories.Rules.Create(parent, body).
		RuleId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating rule: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtRuleFormat)
}

func (s *artifactsCmds) runArtifactsRulesDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artRuleName(project, s.flagArtRuleLocation, s.flagArtRuleRepository, args[0])
	out, err := svc.Projects.Locations.Repositories.Rules.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting rule: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtRuleFormat)
}

func (s *artifactsCmds) runArtifactsRulesDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	name := artRuleName(project, s.flagArtRuleLocation, s.flagArtRuleRepository, args[0])
	r, err := svc.Projects.Locations.Repositories.Rules.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing rule: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, r, s.flagArtRuleFormat)
}

func (s *artifactsCmds) runArtifactsRulesList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	parent := artRepoParent(project, s.flagArtRuleLocation, s.flagArtRuleRepository)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagArtRuleFormat)
}

func (s *artifactsCmds) runArtifactsRulesUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().ArtifactRegistry(ctx)
	if err != nil {
		return err
	}
	body := &artifactregistry.GoogleDevtoolsArtifactregistryV1Rule{}
	if err := loadYAMLOrJSONInto(s.flagArtRuleConfigFile, body); err != nil {
		return err
	}
	mask := s.flagArtRuleMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	name := artRuleName(project, s.flagArtRuleLocation, s.flagArtRuleRepository, args[0])
	call := svc.Projects.Locations.Repositories.Rules.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	out, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating rule: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, out, s.flagArtRuleFormat)
}
