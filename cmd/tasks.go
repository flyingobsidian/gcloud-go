package cmd

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	cloudtasks "google.golang.org/api/cloudtasks/v2"
)

// --- gcloud tasks (#389) ---

func tasksLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

// --- locations ---

// --- cmek-config ---

// --- queues ---

func (s *tasksCmds) initTasks() {
	// locations
	s.tasksLocDescribeCmd.Flags().StringVar(&s.flagTasksFormat, "format", "", "Output format")
	s.tasksLocListCmd.Flags().StringVar(&s.flagTasksFormat, "format", "", "Output format")
	s.tasksLocationsCmd.AddCommand(s.tasksLocDescribeCmd, s.tasksLocListCmd)
	s.tasksCmd.AddCommand(s.tasksLocationsCmd)

	// cmek-config
	for _, c := range []*cobra.Command{s.tasksCmekDescribeCmd, s.tasksCmekUpdateCmd} {
		c.Flags().StringVar(&s.flagTasksCmekLocation, "location", "", "Location whose CMEK config to inspect/update (required)")
		_ = c.MarkFlagRequired("location")
	}
	s.tasksCmekDescribeCmd.Flags().StringVar(&s.flagTasksFormat, "format", "", "Output format")
	s.tasksCmekUpdateCmd.Flags().StringVar(&s.flagTasksCmekKey, "kms-key", "",
		"Fully qualified KMS CryptoKey resource name (mutually exclusive with --clear)")
	s.tasksCmekUpdateCmd.Flags().BoolVar(&s.flagTasksCmekClear, "clear", false, "Clear the CMEK config")
	s.tasksCmekConfigCmd.AddCommand(s.tasksCmekDescribeCmd, s.tasksCmekUpdateCmd)
	s.tasksCmd.AddCommand(s.tasksCmekConfigCmd)

	// queues
	tqAll := []*cobra.Command{s.tqCreateCmd, s.tqDeleteCmd, s.tqDescribeCmd, s.tqListCmd, s.tqUpdateCmd,
		s.tqPauseCmd, s.tqResumeCmd, s.tqPurgeCmd, s.tqGetIamCmd, s.tqSetIamCmd, s.tqAddIamCmd, s.tqRemoveIamCmd}
	for _, c := range tqAll {
		c.Flags().StringVar(&s.flagTQLocation, "location", "", "Location containing the queue (required)")
		_ = c.MarkFlagRequired("location")
	}
	for _, c := range []*cobra.Command{s.tqCreateCmd, s.tqUpdateCmd} {
		c.Flags().StringVar(&s.flagTQConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the Queue message body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.tqUpdateCmd.Flags().StringVar(&s.flagTQUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.tqDescribeCmd.Flags().StringVar(&s.flagTasksFormat, "format", "", "Output format")
	s.tqListCmd.Flags().StringVar(&s.flagTasksFormat, "format", "", "Output format")
	s.tqGetIamCmd.Flags().StringVar(&s.flagTasksFormat, "format", "", "Output format")
	for _, c := range []*cobra.Command{s.tqAddIamCmd, s.tqRemoveIamCmd} {
		c.Flags().StringVar(&s.flagTQIamMember, "member", "", "IAM member (required)")
		c.Flags().StringVar(&s.flagTQIamRole, "role", "", "IAM role (required)")
		_ = c.MarkFlagRequired("member")
		_ = c.MarkFlagRequired("role")
	}
	s.tasksQueuesCmd.AddCommand(tqAll...)
	s.tasksCmd.AddCommand(s.tasksQueuesCmd)

}

func (s *tasksCmds) runTasksLocDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudTasks(ctx)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(tasksLocationParent(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, loc, s.flagTasksFormat)
}

func (s *tasksCmds) runTasksLocList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudTasks(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing locations: %w", err)
	}
	if s.flagTasksFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Locations, s.flagTasksFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Locations), "%-20s %s\n", "LOCATION", "DISPLAY_NAME")
	for _, l := range resp.Locations {
		fmt.Fprintf(s.inv.stdout, "%-20s %s\n", l.LocationId, l.DisplayName)
	}
	return nil
}

func tasksCmekConfigName(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s/cmekConfig", project, location)
}

func (s *tasksCmds) runTasksCmekDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudTasks(ctx)
	if err != nil {
		return err
	}
	cfg, err := svc.Projects.Locations.GetCmekConfig(tasksCmekConfigName(project, s.flagTasksCmekLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing CMEK config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, cfg, s.flagTasksFormat)
}

func (s *tasksCmds) runTasksCmekUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	if s.flagTasksCmekKey == "" && !s.flagTasksCmekClear {
		return fmt.Errorf("either --kms-key or --clear is required")
	}
	if s.flagTasksCmekKey != "" && s.flagTasksCmekClear {
		return fmt.Errorf("--kms-key and --clear are mutually exclusive")
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudTasks(ctx)
	if err != nil {
		return err
	}
	cfg := &cloudtasks.CmekConfig{KmsKey: s.flagTasksCmekKey}
	got, err := svc.Projects.Locations.UpdateCmekConfig(tasksCmekConfigName(project, s.flagTasksCmekLocation), cfg).
		UpdateMask("kmsKey").Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating CMEK config: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func tqName(id, project, location string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/queues/%s", tasksLocationParent(project, location), id)
}

func (s *tasksCmds) runTQCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	q := &cloudtasks.Queue{}
	if err := loadYAMLOrJSONInto(s.flagTQConfigFile, q); err != nil {
		return err
	}
	q.Name = tqName(args[0], project, s.flagTQLocation)
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudTasks(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Queues.Create(tasksLocationParent(project, s.flagTQLocation), q).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating queue: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *tasksCmds) runTQDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudTasks(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Queues.Delete(tqName(args[0], project, s.flagTQLocation)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting queue: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted queue [%s].\n", args[0])
	return nil
}

func (s *tasksCmds) runTQDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudTasks(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Queues.Get(tqName(args[0], project, s.flagTQLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing queue: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagTasksFormat)
}

func (s *tasksCmds) runTQList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudTasks(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.Queues.List(tasksLocationParent(project, s.flagTQLocation)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing queues: %w", err)
	}
	if s.flagTasksFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.Queues, s.flagTasksFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.Queues), "%-40s %s\n", "NAME", "STATE")
	for _, q := range resp.Queues {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(q.Name), q.State)
	}
	return nil
}

func (s *tasksCmds) runTQUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	q := &cloudtasks.Queue{}
	if err := loadYAMLOrJSONInto(s.flagTQConfigFile, q); err != nil {
		return err
	}
	mask := s.flagTQUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(q))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudTasks(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Queues.Patch(tqName(args[0], project, s.flagTQLocation), q).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating queue: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *tasksCmds) runTQPause(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudTasks(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Queues.Pause(tqName(args[0], project, s.flagTQLocation), &cloudtasks.PauseQueueRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("pausing queue: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Paused queue [%s].\n", args[0])
	return nil
}

func (s *tasksCmds) runTQResume(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudTasks(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Queues.Resume(tqName(args[0], project, s.flagTQLocation), &cloudtasks.ResumeQueueRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("resuming queue: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Resumed queue [%s].\n", args[0])
	return nil
}

func (s *tasksCmds) runTQPurge(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudTasks(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Queues.Purge(tqName(args[0], project, s.flagTQLocation), &cloudtasks.PurgeQueueRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("purging queue: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Purged queue [%s].\n", args[0])
	return nil
}

func (s *tasksCmds) runTQGetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudTasks(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Queues.GetIamPolicy(tqName(args[0], project, s.flagTQLocation), &cloudtasks.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagTasksFormat)
}

func (s *tasksCmds) runTQSetIam(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	policy := &cloudtasks.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().CloudTasks(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Queues.SetIamPolicy(tqName(args[0], project, s.flagTQLocation), &cloudtasks.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *tasksCmds) runTQAddIam(cmd *cobra.Command, args []string) error {
	return s.tqModifyIam(args[0], func(p *cloudtasks.Policy) {
		for _, b := range p.Bindings {
			if b.Role == s.flagTQIamRole {
				for _, m := range b.Members {
					if m == s.flagTQIamMember {
						return
					}
				}
				b.Members = append(b.Members, s.flagTQIamMember)
				return
			}
		}
		p.Bindings = append(p.Bindings, &cloudtasks.Binding{Role: s.flagTQIamRole, Members: []string{s.flagTQIamMember}})
	})
}

func (s *tasksCmds) runTQRemoveIam(cmd *cobra.Command, args []string) error {
	return s.tqModifyIam(args[0], func(p *cloudtasks.Policy) {
		for _, b := range p.Bindings {
			if b.Role != s.flagTQIamRole {
				continue
			}
			out := b.Members[:0]
			for _, m := range b.Members {
				if m != s.flagTQIamMember {
					out = append(out, m)
				}
			}
			b.Members = out
		}
	})
}

func (s *tasksCmds) tqModifyIam(name string, mutate func(*cloudtasks.Policy)) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().CloudTasks(ctx)
	if err != nil {
		return err
	}
	resource := tqName(name, project, s.flagTQLocation)
	policy, err := svc.Projects.Locations.Queues.GetIamPolicy(resource, &cloudtasks.GetIamPolicyRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Locations.Queues.SetIamPolicy(resource, &cloudtasks.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// tasksCmds holds the commands and flag values of one invocation's "tasks"
// command group, so invocations don't share state (#2070).
type tasksCmds struct {
	inv *invocation

	tasksCmd              *cobra.Command
	tasksLocationsCmd     *cobra.Command
	tasksLocDescribeCmd   *cobra.Command
	tasksLocListCmd       *cobra.Command
	flagTasksFormat       string
	tasksCmekConfigCmd    *cobra.Command
	tasksCmekDescribeCmd  *cobra.Command
	tasksCmekUpdateCmd    *cobra.Command
	flagTasksCmekLocation string
	flagTasksCmekKey      string
	flagTasksCmekClear    bool
	tasksQueuesCmd        *cobra.Command
	tqCreateCmd           *cobra.Command
	tqDeleteCmd           *cobra.Command
	tqDescribeCmd         *cobra.Command
	tqListCmd             *cobra.Command
	tqUpdateCmd           *cobra.Command
	tqPauseCmd            *cobra.Command
	tqResumeCmd           *cobra.Command
	tqPurgeCmd            *cobra.Command
	tqGetIamCmd           *cobra.Command
	tqSetIamCmd           *cobra.Command
	tqAddIamCmd           *cobra.Command
	tqRemoveIamCmd        *cobra.Command
	flagTQLocation        string
	flagTQConfigFile      string
	flagTQUpdateMask      string
	flagTQIamMember       string
	flagTQIamRole         string
}

// newTasksCmd returns the "tasks" command group for inv.
func newTasksCmd(inv *invocation) *cobra.Command {
	return newTasksCmds(inv).tasksCmd
}

// newTasksCmds builds the "tasks" command group for inv and returns its state.
func newTasksCmds(inv *invocation) *tasksCmds {
	s := &tasksCmds{inv: inv}
	s.tasksCmd = &cobra.Command{Use: "tasks", Short: "Manage Cloud Tasks"}
	s.tasksLocationsCmd = &cobra.Command{Use: "locations", Short: "Explore Cloud Tasks locations"}
	s.tasksLocDescribeCmd = &cobra.Command{
		Use: "describe LOCATION", Short: "Describe a Cloud Tasks location",
		Args: cobra.ExactArgs(1), RunE: s.runTasksLocDescribe,
	}
	s.tasksLocListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Tasks locations",
		Args: cobra.NoArgs, RunE: s.runTasksLocList,
	}
	s.tasksCmekConfigCmd = &cobra.Command{Use: "cmek-config", Short: "Manage CMEK config for Cloud Tasks"}
	s.tasksCmekDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the CMEK config for a location",
		Args: cobra.NoArgs, RunE: s.runTasksCmekDescribe,
	}
	s.tasksCmekUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update or clear the CMEK config (--kms-key or --clear)",
		Args: cobra.NoArgs, RunE: s.runTasksCmekUpdate,
	}
	s.tasksQueuesCmd = &cobra.Command{Use: "queues", Short: "Manage Cloud Tasks queues"}
	s.tqCreateCmd = &cobra.Command{
		Use: "create QUEUE", Short: "Create a queue from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runTQCreate,
	}
	s.tqDeleteCmd = &cobra.Command{
		Use: "delete QUEUE", Short: "Delete a queue",
		Args: cobra.ExactArgs(1), RunE: s.runTQDelete,
	}
	s.tqDescribeCmd = &cobra.Command{
		Use: "describe QUEUE", Short: "Describe a queue",
		Args: cobra.ExactArgs(1), RunE: s.runTQDescribe,
	}
	s.tqListCmd = &cobra.Command{
		Use: "list", Short: "List queues in a location",
		Args: cobra.NoArgs, RunE: s.runTQList,
	}
	s.tqUpdateCmd = &cobra.Command{
		Use: "update QUEUE", Short: "Update a queue from a --config-file",
		Args: cobra.ExactArgs(1), RunE: s.runTQUpdate,
	}
	s.tqPauseCmd = &cobra.Command{
		Use: "pause QUEUE", Short: "Pause a queue",
		Args: cobra.ExactArgs(1), RunE: s.runTQPause,
	}
	s.tqResumeCmd = &cobra.Command{
		Use: "resume QUEUE", Short: "Resume a queue",
		Args: cobra.ExactArgs(1), RunE: s.runTQResume,
	}
	s.tqPurgeCmd = &cobra.Command{
		Use: "purge QUEUE", Short: "Purge all tasks from a queue",
		Args: cobra.ExactArgs(1), RunE: s.runTQPurge,
	}
	s.tqGetIamCmd = &cobra.Command{
		Use: "get-iam-policy QUEUE", Short: "Get the IAM policy for a queue",
		Args: cobra.ExactArgs(1), RunE: s.runTQGetIam,
	}
	s.tqSetIamCmd = &cobra.Command{
		Use: "set-iam-policy QUEUE POLICY_FILE", Short: "Replace the IAM policy for a queue",
		Args: cobra.ExactArgs(2), RunE: s.runTQSetIam,
	}
	s.tqAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding QUEUE", Short: "Add an IAM binding to a queue",
		Args: cobra.ExactArgs(1), RunE: s.runTQAddIam,
	}
	s.tqRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding QUEUE", Short: "Remove an IAM binding from a queue",
		Args: cobra.ExactArgs(1), RunE: s.runTQRemoveIam,
	}
	s.initTasks()
	return s
}
