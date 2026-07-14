package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	transcoder "google.golang.org/api/transcoder/v1"
)

// --- gcloud transcoder (#392) ---

var transcoderCmd = &cobra.Command{Use: "transcoder", Short: "Manage Transcoder"}

// --- jobs ---

var transcoderJobsCmd = &cobra.Command{
	Use:   "jobs",
	Short: "Manage Transcoder jobs",
}

var (
	tcJobCreateCmd = &cobra.Command{
		Use: "create JOB", Short: "Create a Transcoder job from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runTCJobCreate,
	}
	tcJobDeleteCmd = &cobra.Command{
		Use: "delete JOB", Short: "Delete a Transcoder job",
		Args: cobra.ExactArgs(1), RunE: runTCJobDelete,
	}
	tcJobDescribeCmd = &cobra.Command{
		Use: "describe JOB", Short: "Describe a Transcoder job",
		Args: cobra.ExactArgs(1), RunE: runTCJobDescribe,
	}
	tcJobListCmd = &cobra.Command{
		Use: "list", Short: "List Transcoder jobs in a location",
		Args: cobra.NoArgs, RunE: runTCJobList,
	}
)

var (
	flagTCJobLocation   string
	flagTCJobConfigFile string
	flagTCJobFormat     string
)

// --- templates ---

var transcoderTemplatesCmd = &cobra.Command{
	Use:   "templates",
	Short: "Manage Transcoder job templates",
}

var (
	tcTplCreateCmd = &cobra.Command{
		Use: "create TEMPLATE", Short: "Create a Transcoder template from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runTCTplCreate,
	}
	tcTplDeleteCmd = &cobra.Command{
		Use: "delete TEMPLATE", Short: "Delete a Transcoder template",
		Args: cobra.ExactArgs(1), RunE: runTCTplDelete,
	}
	tcTplDescribeCmd = &cobra.Command{
		Use: "describe TEMPLATE", Short: "Describe a Transcoder template",
		Args: cobra.ExactArgs(1), RunE: runTCTplDescribe,
	}
	tcTplListCmd = &cobra.Command{
		Use: "list", Short: "List Transcoder templates in a location",
		Args: cobra.NoArgs, RunE: runTCTplList,
	}
)

var (
	flagTCTplLocation   string
	flagTCTplConfigFile string
	flagTCTplFormat     string
)

func init() {
	for _, c := range []*cobra.Command{tcJobCreateCmd, tcJobDeleteCmd, tcJobDescribeCmd, tcJobListCmd} {
		c.Flags().StringVar(&flagTCJobLocation, "location", "", "Location containing the job (required)")
		_ = c.MarkFlagRequired("location")
	}
	tcJobCreateCmd.Flags().StringVar(&flagTCJobConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the Job message body (required)")
	_ = tcJobCreateCmd.MarkFlagRequired("config-file")
	tcJobDescribeCmd.Flags().StringVar(&flagTCJobFormat, "format", "", "Output format")
	tcJobListCmd.Flags().StringVar(&flagTCJobFormat, "format", "", "Output format")

	for _, c := range []*cobra.Command{tcTplCreateCmd, tcTplDeleteCmd, tcTplDescribeCmd, tcTplListCmd} {
		c.Flags().StringVar(&flagTCTplLocation, "location", "", "Location containing the template (required)")
		_ = c.MarkFlagRequired("location")
	}
	tcTplCreateCmd.Flags().StringVar(&flagTCTplConfigFile, "config-file", "",
		"Path to a JSON/YAML file with the JobTemplate message body (required)")
	_ = tcTplCreateCmd.MarkFlagRequired("config-file")
	tcTplDescribeCmd.Flags().StringVar(&flagTCTplFormat, "format", "", "Output format")
	tcTplListCmd.Flags().StringVar(&flagTCTplFormat, "format", "", "Output format")

	transcoderJobsCmd.AddCommand(tcJobCreateCmd, tcJobDeleteCmd, tcJobDescribeCmd, tcJobListCmd)
	transcoderTemplatesCmd.AddCommand(tcTplCreateCmd, tcTplDeleteCmd, tcTplDescribeCmd, tcTplListCmd)
	transcoderCmd.AddCommand(transcoderJobsCmd, transcoderTemplatesCmd)
	rootCmd.AddCommand(transcoderCmd)
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

func runTCJobCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	job := &transcoder.Job{}
	if err := loadYAMLOrJSONInto(flagTCJobConfigFile, job); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.TranscoderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.Create(fmt.Sprintf("projects/%s/locations/%s", project, flagTCJobLocation), job).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating job: %w", err)
	}
	return emitFormatted(got, "")
}

func runTCJobDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.TranscoderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Jobs.Delete(tcJobName(args[0], project, flagTCJobLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting job: %w", err)
	}
	fmt.Printf("Deleted job [%s].\n", args[0])
	return nil
}

func runTCJobDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.TranscoderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.Get(tcJobName(args[0], project, flagTCJobLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing job: %w", err)
	}
	return emitFormatted(got, flagTCJobFormat)
}

func runTCJobList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.TranscoderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Jobs.List(fmt.Sprintf("projects/%s/locations/%s", project, flagTCJobLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing jobs: %w", err)
	}
	if flagTCJobFormat != "" {
		return emitFormatted(resp.Jobs, flagTCJobFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, j := range resp.Jobs {
		fmt.Printf("%-40s %s\n", path.Base(j.Name), j.State)
	}
	return nil
}

func runTCTplCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	tpl := &transcoder.JobTemplate{}
	if err := loadYAMLOrJSONInto(flagTCTplConfigFile, tpl); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.TranscoderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.JobTemplates.Create(fmt.Sprintf("projects/%s/locations/%s", project, flagTCTplLocation), tpl).
		JobTemplateId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating template: %w", err)
	}
	return emitFormatted(got, "")
}

func runTCTplDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.TranscoderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.JobTemplates.Delete(tcTplName(args[0], project, flagTCTplLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting template: %w", err)
	}
	fmt.Printf("Deleted template [%s].\n", args[0])
	return nil
}

func runTCTplDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.TranscoderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.JobTemplates.Get(tcTplName(args[0], project, flagTCTplLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing template: %w", err)
	}
	return emitFormatted(got, flagTCTplFormat)
}

func runTCTplList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.TranscoderService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.JobTemplates.List(fmt.Sprintf("projects/%s/locations/%s", project, flagTCTplLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing templates: %w", err)
	}
	if flagTCTplFormat != "" {
		return emitFormatted(resp.JobTemplates, flagTCTplFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, t := range resp.JobTemplates {
		fmt.Println(path.Base(t.Name))
	}
	return nil
}
