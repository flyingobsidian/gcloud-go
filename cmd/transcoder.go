package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	transcoder "google.golang.org/api/transcoder/v1"
)

// --- gcloud transcoder (#392) ---

// --- jobs ---

// --- templates ---

func (s *transcoderCmds) initTranscoder() {
	for _, c := range []*cobra.Command{s.tcJobCreateCmd, s.tcJobDeleteCmd, s.tcJobDescribeCmd, s.tcJobListCmd} {
		c.Flags().StringVar(&s.flagTCJobLocation, "location", "", "Location containing the job (required)")
		_ = c.MarkFlagRequired("location")
	}
	s.tcJobCreateCmd.Flags().StringVar(&s.flagTCJobConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Job message body (required)")
	_ = s.tcJobCreateCmd.MarkFlagRequired("config-file")
	s.tcJobDescribeCmd.Flags().StringVar(&s.flagTCJobFormat, "format", "", "Output format")
	s.tcJobListCmd.Flags().StringVar(&s.flagTCJobFormat, "format", "", "Output format")

	for _, c := range []*cobra.Command{s.tcTplCreateCmd, s.tcTplDeleteCmd, s.tcTplDescribeCmd, s.tcTplListCmd} {
		c.Flags().StringVar(&s.flagTCTplLocation, "location", "", "Location containing the template (required)")
		_ = c.MarkFlagRequired("location")
	}
	s.tcTplCreateCmd.Flags().StringVar(&s.flagTCTplConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the JobTemplate message body (required)")
	_ = s.tcTplCreateCmd.MarkFlagRequired("config-file")
	s.tcTplDescribeCmd.Flags().StringVar(&s.flagTCTplFormat, "format", "", "Output format")
	s.tcTplListCmd.Flags().StringVar(&s.flagTCTplFormat, "format", "", "Output format")

	s.transcoderJobsCmd.AddCommand(s.tcJobCreateCmd, s.tcJobDeleteCmd, s.tcJobDescribeCmd, s.tcJobListCmd)
	s.transcoderTemplatesCmd.AddCommand(s.tcTplCreateCmd, s.tcTplDeleteCmd, s.tcTplDescribeCmd, s.tcTplListCmd)
	s.transcoderCmd.AddCommand(s.transcoderJobsCmd, s.transcoderTemplatesCmd)

}

func tcJobName(id, project, location string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/locations/%s/jobs/%s", project, location, id)
}

func tcTplName(id, project, location string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("projects/%s/locations/%s/jobTemplates/%s", project, location, id)
}

func (s *transcoderCmds) runTCJobCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	job := &transcoder.Job{}
	if err := loadYAMLOrJSONInto(s.flagTCJobConfigFile, job); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Transcoder(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.Create(fmt.Sprintf("projects/%s/locations/%s", project, s.flagTCJobLocation), job).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *transcoderCmds) runTCJobDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Transcoder(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Jobs.Delete(tcJobName(args[0], project, s.flagTCJobLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted job [%s].\n", args[0])
	return nil
}

func (s *transcoderCmds) runTCJobDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Transcoder(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.Get(tcJobName(args[0], project, s.flagTCJobLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagTCJobFormat)
}

func (s *transcoderCmds) runTCJobList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Transcoder(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Jobs.List(fmt.Sprintf("projects/%s/locations/%s", project, s.flagTCJobLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing jobs: %w", err)
	}
	if s.flagTCJobFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Jobs, s.flagTCJobFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Jobs), "%-40s %s\n", "NAME", "STATE")
	for _, j := range resp.Jobs {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(j.Name), j.State)
	}
	return nil
}

func (s *transcoderCmds) runTCTplCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	tpl := &transcoder.JobTemplate{}
	if err := loadYAMLOrJSONInto(s.flagTCTplConfigFile, tpl); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Transcoder(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.JobTemplates.Create(fmt.Sprintf("projects/%s/locations/%s", project, s.flagTCTplLocation), tpl).
		JobTemplateId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating template: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *transcoderCmds) runTCTplDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Transcoder(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.JobTemplates.Delete(tcTplName(args[0], project, s.flagTCTplLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting template: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted template [%s].\n", args[0])
	return nil
}

func (s *transcoderCmds) runTCTplDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Transcoder(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.JobTemplates.Get(tcTplName(args[0], project, s.flagTCTplLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing template: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagTCTplFormat)
}

func (s *transcoderCmds) runTCTplList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Transcoder(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.JobTemplates.List(fmt.Sprintf("projects/%s/locations/%s", project, s.flagTCTplLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing templates: %w", err)
	}
	if s.flagTCTplFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.JobTemplates, s.flagTCTplFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.JobTemplates), "%-40s\n", "NAME")
	for _, t := range resp.JobTemplates {
		fmt.Fprintln(s.inv.stdout, path.Base(t.Name))
	}
	return nil
}

// transcoderCmds holds the commands and flag values of one invocation's "transcoder"
// command group, so invocations don't share state (#2070).
type transcoderCmds struct {
	inv *invocation

	transcoderCmd          *cobra.Command
	transcoderJobsCmd      *cobra.Command
	tcJobCreateCmd         *cobra.Command
	tcJobDeleteCmd         *cobra.Command
	tcJobDescribeCmd       *cobra.Command
	tcJobListCmd           *cobra.Command
	flagTCJobLocation      string
	flagTCJobConfigFile    string
	flagTCJobFormat        string
	transcoderTemplatesCmd *cobra.Command
	tcTplCreateCmd         *cobra.Command
	tcTplDeleteCmd         *cobra.Command
	tcTplDescribeCmd       *cobra.Command
	tcTplListCmd           *cobra.Command
	flagTCTplLocation      string
	flagTCTplConfigFile    string
	flagTCTplFormat        string
}

// newTranscoderCmd returns the "transcoder" command group for inv.
func newTranscoderCmd(inv *invocation) *cobra.Command {
	return newTranscoderCmds(inv).transcoderCmd
}

// newTranscoderCmds builds the "transcoder" command group for inv and returns its state.
func newTranscoderCmds(inv *invocation) *transcoderCmds {
	s := &transcoderCmds{inv: inv}
	s.transcoderCmd = &cobra.Command{Use: "transcoder", Short: "Manage Transcoder"}
	s.transcoderJobsCmd = &cobra.Command{
		Use:   "jobs",
		Short: "Manage Transcoder jobs",
	}
	s.tcJobCreateCmd = &cobra.Command{
		Use: "create JOB", Short: "Create a Transcoder job from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runTCJobCreate,
	}
	s.tcJobDeleteCmd = &cobra.Command{
		Use: "delete JOB", Short: "Delete a Transcoder job",
		Args: cobra.ExactArgs(1), RunE: s.runTCJobDelete,
	}
	s.tcJobDescribeCmd = &cobra.Command{
		Use: "describe JOB", Short: "Describe a Transcoder job",
		Args: cobra.ExactArgs(1), RunE: s.runTCJobDescribe,
	}
	s.tcJobListCmd = &cobra.Command{
		Use: "list", Short: "List Transcoder jobs in a location",
		Args: cobra.NoArgs, RunE: s.runTCJobList,
	}
	s.transcoderTemplatesCmd = &cobra.Command{
		Use:   "templates",
		Short: "Manage Transcoder job templates",
	}
	s.tcTplCreateCmd = &cobra.Command{
		Use: "create TEMPLATE", Short: "Create a Transcoder template from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runTCTplCreate,
	}
	s.tcTplDeleteCmd = &cobra.Command{
		Use: "delete TEMPLATE", Short: "Delete a Transcoder template",
		Args: cobra.ExactArgs(1), RunE: s.runTCTplDelete,
	}
	s.tcTplDescribeCmd = &cobra.Command{
		Use: "describe TEMPLATE", Short: "Describe a Transcoder template",
		Args: cobra.ExactArgs(1), RunE: s.runTCTplDescribe,
	}
	s.tcTplListCmd = &cobra.Command{
		Use: "list", Short: "List Transcoder templates in a location",
		Args: cobra.NoArgs, RunE: s.runTCTplList,
	}
	s.initTranscoder()
	return s
}
