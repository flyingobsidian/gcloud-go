package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	asset "google.golang.org/api/cloudasset/v1"
	"gopkg.in/yaml.v3"
)

var assetSavedQueriesCmd = &cobra.Command{
	Use:   "saved-queries",
	Short: "Manage Cloud Asset Inventory saved queries",
}

var assetSavedQueryCreateCmd = &cobra.Command{
	Use:   "create QUERY_ID",
	Short: "Create a saved query",
	Args:  cobra.ExactArgs(1),
	RunE:  runAssetSavedQueryCreate,
}

var assetSavedQueryDeleteCmd = &cobra.Command{
	Use:   "delete QUERY_ID",
	Short: "Delete a saved query",
	Args:  cobra.ExactArgs(1),
	RunE:  runAssetSavedQueryDelete,
}

var assetSavedQueryDescribeCmd = &cobra.Command{
	Use:   "describe QUERY_ID",
	Short: "Describe a saved query",
	Args:  cobra.ExactArgs(1),
	RunE:  runAssetSavedQueryDescribe,
}

var assetSavedQueryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved queries",
	Args:  cobra.NoArgs,
	RunE:  runAssetSavedQueryList,
}

var assetSavedQueryUpdateCmd = &cobra.Command{
	Use:   "update QUERY_ID",
	Short: "Update a saved query",
	Args:  cobra.ExactArgs(1),
	RunE:  runAssetSavedQueryUpdate,
}

var (
	flagAssetSQProject     string
	flagAssetSQFolder      string
	flagAssetSQOrg         string
	flagAssetSQDescription string
	flagAssetSQQueryFile   string
	flagAssetSQLabels      map[string]string
	flagAssetSQListFilter  string
	flagAssetSQListFormat  string
)

func init() {
	for _, c := range []*cobra.Command{
		assetSavedQueryCreateCmd, assetSavedQueryDeleteCmd, assetSavedQueryDescribeCmd,
		assetSavedQueryListCmd, assetSavedQueryUpdateCmd,
	} {
		c.Flags().StringVar(&flagAssetSQProject, "project", "", "Project ID (mutually exclusive with --folder and --organization)")
		c.Flags().StringVar(&flagAssetSQFolder, "folder", "", "Folder ID (mutually exclusive with --project and --organization)")
		c.Flags().StringVar(&flagAssetSQOrg, "organization", "", "Organization ID (mutually exclusive with --project and --folder)")
	}

	for _, c := range []*cobra.Command{assetSavedQueryCreateCmd, assetSavedQueryUpdateCmd} {
		c.Flags().StringVar(&flagAssetSQDescription, "description", "", "Description of the saved query")
		c.Flags().StringVar(&flagAssetSQQueryFile, "query-file-path", "", "Path to a JSON or YAML file containing the saved query content")
		c.Flags().StringToStringVar(&flagAssetSQLabels, "labels", nil, "Labels to attach to the saved query")
	}
	assetSavedQueryCreateCmd.MarkFlagRequired("query-file-path")

	assetSavedQueryListCmd.Flags().StringVar(&flagAssetSQListFilter, "filter", "", "Filter on saved queries")
	assetSavedQueryListCmd.Flags().StringVar(&flagAssetSQListFormat, "format", "", "Output format (json, yaml, or table)")

	assetSavedQueriesCmd.AddCommand(
		assetSavedQueryCreateCmd, assetSavedQueryDeleteCmd, assetSavedQueryDescribeCmd,
		assetSavedQueryListCmd, assetSavedQueryUpdateCmd,
	)
	assetCmd.AddCommand(assetSavedQueriesCmd)
}

func savedQueryParent() (string, error) {
	return resolveAssetScope(flagAssetSQProject, flagAssetSQFolder, flagAssetSQOrg)
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

func runAssetSavedQueryCreate(cmd *cobra.Command, args []string) error {
	parent, err := savedQueryParent()
	if err != nil {
		return err
	}
	content, err := loadSavedQueryContent(flagAssetSQQueryFile)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	q, err := svc.SavedQueries.Create(parent, &asset.SavedQuery{
		Description: flagAssetSQDescription,
		Content:     content,
		Labels:      flagAssetSQLabels,
	}).SavedQueryId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating saved query: %w", err)
	}
	return yamlEncode(q)
}

func runAssetSavedQueryDelete(cmd *cobra.Command, args []string) error {
	parent, err := savedQueryParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.SavedQueries.Delete(savedQueryName(parent, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting saved query: %w", err)
	}
	fmt.Printf("Deleted saved query [%s].\n", args[0])
	return nil
}

func runAssetSavedQueryDescribe(cmd *cobra.Command, args []string) error {
	parent, err := savedQueryParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	q, err := svc.SavedQueries.Get(savedQueryName(parent, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing saved query: %w", err)
	}
	return yamlEncode(q)
}

func runAssetSavedQueryList(cmd *cobra.Command, args []string) error {
	parent, err := savedQueryParent()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}

	var all []*asset.SavedQuery
	pageToken := ""
	for {
		call := svc.SavedQueries.List(parent).Context(ctx)
		if flagAssetSQListFilter != "" {
			call = call.Filter(flagAssetSQListFilter)
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

	return printListResults(all, flagAssetSQListFormat, func() {
		fmt.Printf("%-60s %s\n", "NAME", "DESCRIPTION")
		for _, q := range all {
			fmt.Printf("%-60s %s\n", q.Name, q.Description)
		}
	})
}

func runAssetSavedQueryUpdate(cmd *cobra.Command, args []string) error {
	parent, err := savedQueryParent()
	if err != nil {
		return err
	}
	content, err := loadSavedQueryContent(flagAssetSQQueryFile)
	if err != nil {
		return err
	}
	name := savedQueryName(parent, args[0])
	q := &asset.SavedQuery{
		Name:        name,
		Description: flagAssetSQDescription,
		Labels:      flagAssetSQLabels,
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

	ctx := context.Background()
	svc, err := gcp.CloudAssetService(ctx, flagAccount)
	if err != nil {
		return err
	}
	updated, err := svc.SavedQueries.Patch(name, q).UpdateMask(strings.Join(masks, ",")).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating saved query: %w", err)
	}
	return yamlEncode(updated)
}
