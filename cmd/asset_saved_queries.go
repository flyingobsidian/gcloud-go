package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	asset "google.golang.org/api/cloudasset/v1"
	"gopkg.in/yaml.v3"
)

func (s *assetCmds) initAssetSavedQueries() {
	for _, c := range []*cobra.Command{
		s.assetSavedQueryCreateCmd, s.assetSavedQueryDeleteCmd, s.assetSavedQueryDescribeCmd,
		s.assetSavedQueryListCmd, s.assetSavedQueryUpdateCmd,
	} {
		c.Flags().StringVar(&s.flagAssetSQProject, "project", "", "Project ID (mutually exclusive with --folder and --organization)")
		c.Flags().StringVar(&s.flagAssetSQFolder, "folder", "", "Folder ID (mutually exclusive with --project and --organization)")
		c.Flags().StringVar(&s.flagAssetSQOrg, "organization", "", "Organization ID (mutually exclusive with --project and --folder)")
	}

	for _, c := range []*cobra.Command{s.assetSavedQueryCreateCmd, s.assetSavedQueryUpdateCmd} {
		c.Flags().StringVar(&s.flagAssetSQDescription, "description", "", "Description of the saved query")
		c.Flags().StringVar(&s.flagAssetSQQueryFile, "query-file-path", "", "Path to a JSON or YAML file containing the saved query content")
		c.Flags().StringToStringVar(&s.flagAssetSQLabels, "labels", nil, "Labels to attach to the saved query")
	}
	s.assetSavedQueryCreateCmd.MarkFlagRequired("query-file-path")

	s.assetSavedQueryListCmd.Flags().StringVar(&s.flagAssetSQListFilter, "filter", "", "Filter on saved queries")
	s.assetSavedQueryListCmd.Flags().StringVar(&s.flagAssetSQListFormat, "format", "", "Output format (json, yaml, or table)")

	s.assetSavedQueriesCmd.AddCommand(
		s.assetSavedQueryCreateCmd, s.assetSavedQueryDeleteCmd, s.assetSavedQueryDescribeCmd,
		s.assetSavedQueryListCmd, s.assetSavedQueryUpdateCmd,
	)
	s.assetCmd.AddCommand(s.assetSavedQueriesCmd)
}

func (s *assetCmds) savedQueryParent() (string, error) {
	return resolveAssetScope(s.flagAssetSQProject, s.flagAssetSQFolder, s.flagAssetSQOrg)
}

func savedQueryName(parent, id string) string {
	if strings.Contains(id, "/savedQueries/") {
		return id
	}
	return parent + "/savedQueries/" + strings.TrimPrefix(id, "savedQueries/")
}

// loadSavedQueryContent reads a query content file (JSON or YAML) into a
// QueryContent struct.
func loadSavedQueryContent(path string) (*asset.QueryContent, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading query file: %w", err)
	}
	content := &asset.QueryContent{}
	if err := yamlUnmarshalOrJSON(data, content); err != nil {
		return nil, fmt.Errorf("parsing query file: %w", err)
	}
	return content, nil
}

// yamlUnmarshalOrJSON decodes the data as JSON first (strictest for known
// keys) and falls back to YAML for files written with YAML syntax.
func yamlUnmarshalOrJSON(data []byte, out any) error {
	if err := json.Unmarshal(data, out); err == nil {
		return nil
	}
	return yaml.Unmarshal(data, out)
}

func (s *assetCmds) runAssetSavedQueryCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.savedQueryParent()
	if err != nil {
		return err
	}
	content, err := loadSavedQueryContent(s.flagAssetSQQueryFile)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	q, err := svc.SavedQueries.Create(parent, &asset.SavedQuery{
		Description: s.flagAssetSQDescription,
		Content:     content,
		Labels:      s.flagAssetSQLabels,
	}).SavedQueryId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating saved query: %w", err)
	}
	return s.inv.yaml(q)
}

func (s *assetCmds) runAssetSavedQueryDelete(cmd *cobra.Command, args []string) error {
	parent, err := s.savedQueryParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.SavedQueries.Delete(savedQueryName(parent, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting saved query: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted saved query [%s].\n", args[0])
	return nil
}

func (s *assetCmds) runAssetSavedQueryDescribe(cmd *cobra.Command, args []string) error {
	parent, err := s.savedQueryParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	q, err := svc.SavedQueries.Get(savedQueryName(parent, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing saved query: %w", err)
	}
	return s.inv.yaml(q)
}

func (s *assetCmds) runAssetSavedQueryList(cmd *cobra.Command, args []string) error {
	parent, err := s.savedQueryParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}

	var all []*asset.SavedQuery
	pageToken := ""
	for {
		call := svc.SavedQueries.List(parent).Context(ctx)
		if s.flagAssetSQListFilter != "" {
			call = call.Filter(s.flagAssetSQListFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing saved queries: %w", err)
		}
		all = append(all, resp.SavedQueries...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return s.inv.printListResults(all, s.flagAssetSQListFormat, func() {
		printTableHeadingTo(s.inv.stdout, len(all), "%-60s %s\n", "NAME", "DESCRIPTION")
		for _, q := range all {
			fmt.Fprintf(s.inv.stdout, "%-60s %s\n", q.Name, q.Description)
		}
	})
}

func (s *assetCmds) runAssetSavedQueryUpdate(cmd *cobra.Command, args []string) error {
	parent, err := s.savedQueryParent()
	if err != nil {
		return err
	}
	content, err := loadSavedQueryContent(s.flagAssetSQQueryFile)
	if err != nil {
		return err
	}
	name := savedQueryName(parent, args[0])
	q := &asset.SavedQuery{
		Name:        name,
		Description: s.flagAssetSQDescription,
		Labels:      s.flagAssetSQLabels,
		Content:     content,
	}

	var masks []string
	if cmd.Flags().Changed("description") {
		masks = append(masks, "description")
	}
	if cmd.Flags().Changed("labels") {
		masks = append(masks, "labels")
	}
	if cmd.Flags().Changed("query-file-path") {
		masks = append(masks, "content")
	}
	if len(masks) == 0 {
		return fmt.Errorf("at least one of --description, --labels, or --query-file-path must be provided")
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudAsset(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.SavedQueries.Patch(name, q).UpdateMask(strings.Join(masks, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating saved query: %w", err)
	}
	return s.inv.yaml(updated)
}
