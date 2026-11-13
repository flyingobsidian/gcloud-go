package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	notebooksv1 "google.golang.org/api/notebooks/v1"
)

// --- gcloud notebooks environments (#1061) ---

func (s *notebooksCmds) initNotebooksEnvironments() {
	all := []*cobra.Command{
		s.notebooksEnvCreateCmd, s.notebooksEnvDeleteCmd, s.notebooksEnvDescribeCmd, s.notebooksEnvListCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNotebooksEnvLocation, "location", "", "Notebook location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNotebooksEnvFormat, "format", "", "Output format")
	}
	s.notebooksEnvCreateCmd.Flags().StringVar(&s.flagNotebooksEnvConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the Environment body (required)")
	_ = s.notebooksEnvCreateCmd.MarkFlagRequired("config-file")
	s.notebooksEnvListCmd.Flags().Int64Var(&s.flagNotebooksEnvPageSize, "page-size", 0, "Maximum results per page")

	s.notebooksEnvCmd.AddCommand(all...)
	s.notebooksCmd.AddCommand(s.notebooksEnvCmd)
}

func (s *notebooksCmds) notebooksEnvParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagNotebooksEnvLocation), nil
}

func (s *notebooksCmds) notebooksEnvName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.notebooksEnvParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/environments/%s", parent, id), nil
}

func (s *notebooksCmds) runNotebooksEnvCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.notebooksEnvParent()
	if err != nil {
		return err
	}
	body := &notebooksv1.Environment{}
	if err := loadYAMLOrJSONInto(s.flagNotebooksEnvConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.Create(parent, body).EnvironmentId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating notebook environment: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for notebook environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksEnvFormat)
}

func (s *notebooksCmds) runNotebooksEnvDelete(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksEnvName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Environments.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting notebook environment: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for notebook environment [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksEnvFormat)
}

func (s *notebooksCmds) runNotebooksEnvDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksEnvName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Environments.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing notebook environment: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNotebooksEnvFormat)
}

func (s *notebooksCmds) runNotebooksEnvList(cmd *cobra.Command, args []string) error {
	parent, err := s.notebooksEnvParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	var all []*notebooksv1.Environment
	pageToken := ""
	for {
		call := svc.Projects.Locations.Environments.List(parent).Context(ctx)
		if s.flagNotebooksEnvPageSize > 0 {
			call = call.PageSize(s.flagNotebooksEnvPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing notebook environments: %w", err)
		}
		all = append(all, resp.Environments...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNotebooksEnvFormat)
}
