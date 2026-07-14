package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	cloudtasks "google.golang.org/api/cloudtasks/v2"
)

// --- gcloud tasks (#389) ---

var tasksCmd = &cobra.Command{Use: "tasks", Short: "Manage Cloud Tasks"}

func tasksLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

// --- locations ---

var tasksLocationsCmd = &cobra.Command{Use: "locations", Short: "Explore Cloud Tasks locations"}

var (
	tasksLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a Cloud Tasks location",
		Args: cobra.ExactArgs(1), RunE: runTasksLocDescribe,
	}
	tasksLocListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Tasks locations",
		Args: cobra.NoArgs, RunE: runTasksLocList,
	}
)

var flagTasksFormat string

// --- cmek-config ---

var tasksCmekConfigCmd = &cobra.Command{Use: "cmek-config", Short: "Manage CMEK config for Cloud Tasks"}

var (
	tasksCmekDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the CMEK config for a location",
		Args: cobra.NoArgs, RunE: runTasksCmekDescribe,
	}
	tasksCmekUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update or clear the CMEK config (--kms-key or --clear)",
		Args: cobra.NoArgs, RunE: runTasksCmekUpdate,
	}
)

var (
	flagTasksCmekLocation string
	flagTasksCmekKey      string
	flagTasksCmekClear    bool
)

// --- queues ---

var tasksQueuesCmd = &cobra.Command{Use: "queues", Short: "Manage Cloud Tasks queues"}

var (
	tqCreateCmd = &cobra.Command{
		Use: "create QUEUE", Short: "Create a queue from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runTQCreate,
	}
	tqDeleteCmd = &cobra.Command{
		Use: "delete QUEUE", Short: "Delete a queue",
		Args: cobra.ExactArgs(1), RunE: runTQDelete,
	}
	tqDescribeCmd = &cobra.Command{
		Use: "describe QUEUE", Short: "Describe a queue",
		Args: cobra.ExactArgs(1), RunE: runTQDescribe,
	}
	tqListCmd = &cobra.Command{
		Use: "list", Short: "List queues in a location",
		Args: cobra.NoArgs, RunE: runTQList,
	}
	tqUpdateCmd = &cobra.Command{
		Use: "update QUEUE", Short: "Update a queue from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runTQUpdate,
	}
	tqPauseCmd = &cobra.Command{
		Use: "pause QUEUE", Short: "Pause a queue",
		Args: cobra.ExactArgs(1), RunE: runTQPause,
	}
	tqResumeCmd = &cobra.Command{
		Use: "resume QUEUE", Short: "Resume a queue",
		Args: cobra.ExactArgs(1), RunE: runTQResume,
	}
	tqPurgeCmd = &cobra.Command{
		Use: "purge QUEUE", Short: "Purge all tasks from a queue",
		Args: cobra.ExactArgs(1), RunE: runTQPurge,
	}
	tqGetIamCmd = &cobra.Command{
		Use: "get-iam-policy QUEUE", Short: "Get the IAM policy for a queue",
		Args: cobra.ExactArgs(1), RunE: runTQGetIam,
	}
	tqSetIamCmd = &cobra.Command{
		Use: "set-iam-policy QUEUE POLICY_FILE", Short: "Replace the IAM policy for a queue",
		Args: cobra.ExactArgs(2), RunE: runTQSetIam,
	}
	tqAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding QUEUE", Short: "Add an IAM binding to a queue",
		Args: cobra.ExactArgs(1), RunE: runTQAddIam,
	}
	tqRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding QUEUE", Short: "Remove an IAM binding from a queue",
		Args: cobra.ExactArgs(1), RunE: runTQRemoveIam,
	}
)

var (
	flagTQLocation   string
	flagTQConfigFile string
	flagTQUpdateMask string
	flagTQIamMember  string
	flagTQIamRole    string
)

func init() {
	// locations
	tasksLocDescribeCmd.Flags().StringVar(&flagTasksFormat, "format", "", "Output format")
	tasksLocListCmd.Flags().StringVar(&flagTasksFormat, "format", "", "Output format")
	tasksLocationsCmd.AddCommand(tasksLocDescribeCmd, tasksLocListCmd)
	tasksCmd.AddCommand(tasksLocationsCmd)

	// cmek-config
	for _, c := range []*cobra.Command{tasksCmekDescribeCmd, tasksCmekUpdateCmd} {
		c.Flags().StringVar(&flagTasksCmekLocation, "location", "", "Location whose CMEK config to inspect/update (required)")
		_ = c.MarkFlagRequired("location")
	}
	tasksCmekDescribeCmd.Flags().StringVar(&flagTasksFormat, "format", "", "Output format")
	tasksCmekUpdateCmd.Flags().StringVar(&flagTasksCmekKey, "kms-key", "",
		"Fully qualified KMS CryptoKey resource name (mutually exclusive with --clear)")
	tasksCmekUpdateCmd.Flags().BoolVar(&flagTasksCmekClear, "clear", false, "Clear the CMEK config")
	tasksCmekConfigCmd.AddCommand(tasksCmekDescribeCmd, tasksCmekUpdateCmd)
	tasksCmd.AddCommand(tasksCmekConfigCmd)

	// queues
	tqAll := []*cobra.Command{tqCreateCmd, tqDeleteCmd, tqDescribeCmd, tqListCmd, tqUpdateCmd,
		tqPauseCmd, tqResumeCmd, tqPurgeCmd, tqGetIamCmd, tqSetIamCmd, tqAddIamCmd, tqRemoveIamCmd}
	for _, c := range tqAll {
		c.Flags().StringVar(&flagTQLocation, "location", "", "Location containing the queue (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{tqCreateCmd, tqUpdateCmd} {
		c.Flags().StringVar(&flagTQConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Queue message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	tqUpdateCmd.Flags().StringVar(&flagTQUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	tqDescribeCmd.Flags().StringVar(&flagTasksFormat, "format", "", "Output format")
	tqListCmd.Flags().StringVar(&flagTasksFormat, "format", "", "Output format")
	tqGetIamCmd.Flags().StringVar(&flagTasksFormat, "format", "", "Output format")
	for _, c := range []*cobra.Command{tqAddIamCmd, tqRemoveIamCmd} {
		c.Flags().StringVar(&flagTQIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&flagTQIamRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	tasksQueuesCmd.AddCommand(tqAll...)
	tasksCmd.AddCommand(tasksQueuesCmd)

	rootCmd.AddCommand(tasksCmd)
}

func runTasksLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudTasksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(tasksLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormatted(loc, flagTasksFormat)
}

func runTasksLocList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudTasksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	if flagTasksFormat != "" {
		return emitFormatted(resp.Locations, flagTasksFormat)
	}
	fmt.Printf("%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Printf("%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

func tasksCmekConfigName(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s/cmekConfig", project, location)
}

func runTasksCmekDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudTasksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	cfg, err := svc.Projects.Locations.GetCmekConfig(tasksCmekConfigName(project, flagTasksCmekLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing CMEK config: %w", err)
	}
	return emitFormatted(cfg, flagTasksFormat)
}

func runTasksCmekUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	if flagTasksCmekKey == "" && !flagTasksCmekClear {
		return fmt.Errorf("either --kms-key or --clear is required")
	}
	if flagTasksCmekKey != "" && flagTasksCmekClear {
		return fmt.Errorf("--kms-key and --clear are mutually exclusive")
	}
	ctx := context.Background()
	svc, err := gcp.CloudTasksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	cfg := &cloudtasks.CmekConfig{KmsKey: flagTasksCmekKey}
	got, err := svc.Projects.Locations.UpdateCmekConfig(tasksCmekConfigName(project, flagTasksCmekLocation), cfg).
		UpdateMask("kmsKey").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating CMEK config: %w", err)
	}
	return emitFormatted(got, "")
}

func tqName(id, project, location string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/queues/%s", tasksLocationParent(project, location), id)
}

func runTQCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	q := &cloudtasks.Queue{}
	if err := loadYAMLOrJSONInto(flagTQConfigFile, q); err != nil {
		return err
	}
	q.Name = tqName(args[0], project, flagTQLocation)
	ctx := context.Background()
	svc, err := gcp.CloudTasksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Queues.Create(tasksLocationParent(project, flagTQLocation), q).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating queue: %w", err)
	}
	return emitFormatted(got, "")
}

func runTQDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudTasksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Queues.Delete(tqName(args[0], project, flagTQLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting queue: %w", err)
	}
	fmt.Printf("Deleted queue [%s].\n", args[0])
	return nil
}

func runTQDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudTasksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Queues.Get(tqName(args[0], project, flagTQLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing queue: %w", err)
	}
	return emitFormatted(got, flagTasksFormat)
}

func runTQList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudTasksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Queues.List(tasksLocationParent(project, flagTQLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing queues: %w", err)
	}
	if flagTasksFormat != "" {
		return emitFormatted(resp.Queues, flagTasksFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "STATE")
	for _, q := range resp.Queues {
		fmt.Printf("%-40s %s\n", path.Base(q.Name), q.State)
	}
	return nil
}

func runTQUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	q := &cloudtasks.Queue{}
	if err := loadYAMLOrJSONInto(flagTQConfigFile, q); err != nil {
		return err
	}
	mask := flagTQUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(q))
	}
	ctx := context.Background()
	svc, err := gcp.CloudTasksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Queues.Patch(tqName(args[0], project, flagTQLocation), q).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating queue: %w", err)
	}
	return emitFormatted(got, "")
}

func runTQPause(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudTasksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Queues.Pause(tqName(args[0], project, flagTQLocation), &cloudtasks.PauseQueueRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("pausing queue: %w", err)
	}
	fmt.Printf("Paused queue [%s].\n", args[0])
	return nil
}

func runTQResume(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudTasksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Queues.Resume(tqName(args[0], project, flagTQLocation), &cloudtasks.ResumeQueueRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("resuming queue: %w", err)
	}
	fmt.Printf("Resumed queue [%s].\n", args[0])
	return nil
}

func runTQPurge(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudTasksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Queues.Purge(tqName(args[0], project, flagTQLocation), &cloudtasks.PurgeQueueRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("purging queue: %w", err)
	}
	fmt.Printf("Purged queue [%s].\n", args[0])
	return nil
}

func runTQGetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudTasksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Queues.GetIamPolicy(tqName(args[0], project, flagTQLocation), &cloudtasks.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormatted(policy, flagTasksFormat)
}

func runTQSetIam(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	policy := &cloudtasks.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudTasksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Queues.SetIamPolicy(tqName(args[0], project, flagTQLocation), &cloudtasks.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}

func runTQAddIam(cmd *cobra.Command, args []string) error {
	return tqModifyIam(args[0], func(p *cloudtasks.Policy) {
		for _, b := range p.Bindings {
			if b.Role == flagTQIamRole {
				for _, m := range b.Members {
					if m == flagTQIamMember {
						return
					}
				}
				b.Members = append(b.Members, flagTQIamMember)
				return
			}
		}
		p.Bindings = append(p.Bindings, &cloudtasks.Binding{Role: flagTQIamRole, Members: []string{flagTQIamMember}})
	})
}

func runTQRemoveIam(cmd *cobra.Command, args []string) error {
	return tqModifyIam(args[0], func(p *cloudtasks.Policy) {
		for _, b := range p.Bindings {
			if b.Role != flagTQIamRole {
				continue
			}
			out := b.Members[:0]
			for _, m := range b.Members {
				if m != flagTQIamMember {
					out = append(out, m)
				}
			}
			b.Members = out
		}
	})
}

func tqModifyIam(name string, mutate func(*cloudtasks.Policy)) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.CloudTasksService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resource := tqName(name, project, flagTQLocation)
	policy, err := svc.Projects.Locations.Queues.GetIamPolicy(resource, &cloudtasks.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Locations.Queues.SetIamPolicy(resource, &cloudtasks.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormatted(got, "")
}
