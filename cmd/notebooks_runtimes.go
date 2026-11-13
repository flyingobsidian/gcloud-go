package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	notebooksv1 "google.golang.org/api/notebooks/v1"
)

// --- gcloud notebooks runtimes (#1064) ---

func (s *notebooksCmds) initNotebooksRuntimes() {
	all := []*cobra.Command{
		s.notebooksRuntimeCreateCmd, s.notebooksRuntimeDeleteCmd, s.notebooksRuntimeDescribeCmd,
		s.notebooksRuntimeListCmd,
		s.notebooksRuntimeDiagnoseCmd, s.notebooksRuntimeMigrateCmd,
		s.notebooksRuntimeResetCmd, s.notebooksRuntimeStartCmd, s.notebooksRuntimeStopCmd,
		s.notebooksRuntimeSwitchCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNotebooksRuntimeLocation, "location", "", "Notebook location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNotebooksRuntimeFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		s.notebooksRuntimeCreateCmd, s.notebooksRuntimeDiagnoseCmd,
		s.notebooksRuntimeMigrateCmd, s.notebooksRuntimeSwitchCmd,
	} {
		c.Flags().StringVar(&s.flagNotebooksRuntimeConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.notebooksRuntimeListCmd.Flags().Int64Var(&s.flagNotebooksRuntimePageSize, "page-size", 0, "Maximum results per page")

	s.notebooksRuntimeCmd.AddCommand(all...)
	s.notebooksCmd.AddCommand(s.notebooksRuntimeCmd)
}

func (s *notebooksCmds) notebooksRuntimeParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagNotebooksRuntimeLocation), nil
}

func (s *notebooksCmds) notebooksRuntimeName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.notebooksRuntimeParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/runtimes/%s", parent, id), nil
}

func (s *notebooksCmds) runNotebooksRuntimeCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.notebooksRuntimeParent()
	if err != nil {
		return err
	}
	body := &notebooksv1.Runtime{}
	if err := loadYAMLOrJSONInto(s.flagNotebooksRuntimeConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Runtimes.Create(parent, body).RuntimeId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating notebook runtime: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for notebook runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksRuntimeFormat)
}

func (s *notebooksCmds) runNotebooksRuntimeDelete(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksRuntimeName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Runtimes.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting notebook runtime: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for notebook runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksRuntimeFormat)
}

func (s *notebooksCmds) runNotebooksRuntimeDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksRuntimeName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Runtimes.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing notebook runtime: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNotebooksRuntimeFormat)
}

func (s *notebooksCmds) runNotebooksRuntimeList(cmd *cobra.Command, args []string) error {
	parent, err := s.notebooksRuntimeParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	var all []*notebooksv1.Runtime
	pageToken := ""
	for {
		call := svc.Projects.Locations.Runtimes.List(parent).Context(ctx)
		if s.flagNotebooksRuntimePageSize > 0 {
			call = call.PageSize(s.flagNotebooksRuntimePageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing notebook runtimes: %w", err)
		}
		all = append(all, resp.Runtimes...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNotebooksRuntimeFormat)
}

func (s *notebooksCmds) runNotebooksRuntimeDiagnose(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksRuntimeName(args[0])
	if err != nil {
		return err
	}
	req := &notebooksv1.DiagnoseRuntimeRequest{}
	if err := loadYAMLOrJSONInto(s.flagNotebooksRuntimeConfigFile, req); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Runtimes.Diagnose(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("diagnosing notebook runtime: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Diagnose request issued for notebook runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksRuntimeFormat)
}

func (s *notebooksCmds) runNotebooksRuntimeMigrate(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksRuntimeName(args[0])
	if err != nil {
		return err
	}
	req := &notebooksv1.MigrateRuntimeRequest{}
	if err := loadYAMLOrJSONInto(s.flagNotebooksRuntimeConfigFile, req); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Runtimes.Migrate(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("migrating notebook runtime: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Migrate request issued for notebook runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksRuntimeFormat)
}

func (s *notebooksCmds) runNotebooksRuntimeReset(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksRuntimeName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Runtimes.Reset(name, &notebooksv1.ResetRuntimeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting notebook runtime: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Reset request issued for notebook runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksRuntimeFormat)
}

func (s *notebooksCmds) runNotebooksRuntimeStart(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksRuntimeName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Runtimes.Start(name, &notebooksv1.StartRuntimeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting notebook runtime: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Start request issued for notebook runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksRuntimeFormat)
}

func (s *notebooksCmds) runNotebooksRuntimeStop(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksRuntimeName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Runtimes.Stop(name, &notebooksv1.StopRuntimeRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping notebook runtime: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Stop request issued for notebook runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksRuntimeFormat)
}

func (s *notebooksCmds) runNotebooksRuntimeSwitch(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksRuntimeName(args[0])
	if err != nil {
		return err
	}
	req := &notebooksv1.SwitchRuntimeRequest{}
	if err := loadYAMLOrJSONInto(s.flagNotebooksRuntimeConfigFile, req); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Runtimes.Switch(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("switching notebook runtime: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Switch request issued for notebook runtime [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksRuntimeFormat)
}
