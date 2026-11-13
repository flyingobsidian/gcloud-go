package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	aiplatform "google.golang.org/api/aiplatform/v1"
)

// --- gcloud ai custom-jobs (#1451) ---

func (s *aiCmds) initAiCustomJobs() {
	all := []*cobra.Command{
		s.aiCJCancelCmd, s.aiCJCreateCmd, s.aiCJDescribeCmd, s.aiCJListCmd,
		s.aiCJStreamLogsCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagAICJRegion, "region", "", "Region where the custom job lives (required)")
		_ = c.MarkFlagRequired("region")
		c.Flags().StringVar(&s.flagAICJFormat, "format", "", "Output format")
	}
	s.aiCJCreateCmd.Flags().StringVar(&s.flagAICJConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the CustomJob body (required)")
	_ = s.aiCJCreateCmd.MarkFlagRequired("config-file")
	s.aiCJListCmd.Flags().StringVar(&s.flagAICJFilter, "filter", "", "Server-side filter expression")
	s.aiCJListCmd.Flags().StringVar(&s.flagAICJOrderBy, "order-by", "", "Order-by expression")
	s.aiCJListCmd.Flags().Int64Var(&s.flagAICJPageSize, "page-size", 0, "Maximum results per page")
	s.aiCJListCmd.Flags().StringVar(&s.flagAICJReadMask, "read-mask", "", "Field mask for reads")
	s.aiCJStreamLogsCmd.Flags().DurationVar(&s.flagAICJTimeout, "timeout", 30*time.Minute,
		"Maximum time to wait for the job to reach a terminal state")

	s.aiCustomJobsCmd.AddCommand(all...)
	s.aiCustomJobsCmd.AddCommand(s.aiCJLocalRunCmd)
	s.aiCmd.AddCommand(s.aiCustomJobsCmd)
}

func (s *aiCmds) aiCJParent() (string, error) { return s.aiParent(s.flagAICJRegion) }

func (s *aiCmds) aiCJName(id string) (string, error) {
	parent, err := s.aiCJParent()
	if err != nil {
		return "", err
	}
	return aiChild("customJobs", id, parent), nil
}

func (s *aiCmds) runAICJCancel(cmd *cobra.Command, args []string) error {
	name, err := s.aiCJName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAICJRegion)
	if err != nil {
		return err
	}
	_, err = svc.Projects.Locations.CustomJobs.Cancel(name, &aiplatform.GoogleCloudAiplatformV1CancelCustomJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancelling custom job: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Cancel request issued for custom job [%s].\n", args[0])
	return nil
}

func (s *aiCmds) runAICJCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.aiCJParent()
	if err != nil {
		return err
	}
	body := &aiplatform.GoogleCloudAiplatformV1CustomJob{}
	if err := loadYAMLOrJSONInto(s.flagAICJConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAICJRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CustomJobs.Create(parent, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating custom job: %w", err)
	}
	fmt.Fprintf(s.inv.stderr, "Created custom job [%s].\n", got.Name)
	return emitFormattedTo(s.inv.stdout, got, s.flagAICJFormat)
}

func (s *aiCmds) runAICJDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.aiCJName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAICJRegion)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.CustomJobs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing custom job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagAICJFormat)
}

func (s *aiCmds) runAICJList(cmd *cobra.Command, args []string) error {
	parent, err := s.aiCJParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAICJRegion)
	if err != nil {
		return err
	}
	var all []*aiplatform.GoogleCloudAiplatformV1CustomJob
	pageToken := ""
	for {
		call := svc.Projects.Locations.CustomJobs.List(parent).Context(ctx)
		if s.flagAICJFilter != "" {
			call = call.Filter(s.flagAICJFilter)
		}
		if s.flagAICJPageSize > 0 {
			call = call.PageSize(s.flagAICJPageSize)
		}
		if s.flagAICJReadMask != "" {
			call = call.ReadMask(s.flagAICJReadMask)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing custom jobs: %w", err)
		}
		all = append(all, resp.CustomJobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagAICJFormat)
}

func runAICJLocalRun(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("local-run requires Docker and is not supported in gcloud-go; use \"create\" to submit to Vertex AI")
}

// runAICJStreamLogs polls the custom job until it reaches a terminal state
// (JOB_STATE_SUCCEEDED / JOB_STATE_FAILED / JOB_STATE_CANCELLED /
// JOB_STATE_EXPIRED / JOB_STATE_PARTIALLY_SUCCEEDED) or --timeout expires.
// The aiplatform REST surface does not expose training-container log entries;
// callers wanting the full log stream should use `gcloud logging tail`.
func (s *aiCmds) runAICJStreamLogs(cmd *cobra.Command, args []string) error {
	name, err := s.aiCJName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().AIPlatform(ctx, s.flagAICJRegion)
	if err != nil {
		return err
	}
	deadline := time.Now().Add(s.flagAICJTimeout)
	for {
		got, err := svc.Projects.Locations.CustomJobs.Get(name).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling custom job: %w", err)
		}
		fmt.Fprintf(s.inv.stderr, "[%s] %s state=%s\n",
			time.Now().Format(time.RFC3339), got.Name, got.State)
		if aiJobStateTerminal(got.State) {
			return emitFormattedTo(s.inv.stdout, got, s.flagAICJFormat)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out after %s waiting for custom job [%s] to reach terminal state (last state=%s)",
				s.flagAICJTimeout, args[0], got.State)
		}
		time.Sleep(5 * time.Second)
	}
}

// aiJobStateTerminal reports whether the given aiplatform JobState value is a
// terminal state. Shared with hp-tuning-jobs stream-logs.
func aiJobStateTerminal(state string) bool {
	switch strings.ToUpper(state) {
	case "JOB_STATE_SUCCEEDED",
		"JOB_STATE_FAILED",
		"JOB_STATE_CANCELLED",
		"JOB_STATE_EXPIRED",
		"JOB_STATE_PARTIALLY_SUCCEEDED":
		return true
	}
	return false
}
