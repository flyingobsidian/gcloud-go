package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	storagebatchoperations "google.golang.org/api/storagebatchoperations/v1"
)

// --- gcloud storage batch-operations (#1236) ---

func (s *storageCmds) initStorageBatchOperations() {
	jobs := []*cobra.Command{
		s.storageBatchOpsJobsCancelCmd, s.storageBatchOpsJobsCreateCmd,
		s.storageBatchOpsJobsDeleteCmd, s.storageBatchOpsJobsDescribeCmd,
		s.storageBatchOpsJobsListCmd,
	}
	bucketOps := []*cobra.Command{
		s.storageBatchOpsBucketOpsDescribeCmd, s.storageBatchOpsBucketOpsListCmd,
	}
	all := append(append([]*cobra.Command{}, jobs...), bucketOps...)
	for _, c := range all {
		c.Flags().StringVar(&s.flagStBoLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagStBoFormat, "format", "", "Output format")
	}
	s.storageBatchOpsJobsCreateCmd.Flags().StringVar(&s.flagStBoConfigFile, "config-file", "", "YAML/JSON file with the Job body (required)")
	_ = s.storageBatchOpsJobsCreateCmd.MarkFlagRequired("config-file")
	s.storageBatchOpsJobsListCmd.Flags().Int64Var(&s.flagStBoPageSize, "page-size", 0, "Maximum results per page")
	for _, c := range bucketOps {
		c.Flags().StringVar(&s.flagStBoJob, "job", "", "Parent batch job id (required)")
		_ = c.MarkFlagRequired("job")
	}
	s.storageBatchOpsBucketOpsListCmd.Flags().Int64Var(&s.flagStBoPageSize, "page-size", 0, "Maximum results per page")

	s.storageBatchOpsJobsCmd.AddCommand(jobs...)
	s.storageBatchOpsBucketOpsCmd.AddCommand(bucketOps...)
	s.storageBatchOperationsCmd.AddCommand(s.storageBatchOpsJobsCmd, s.storageBatchOpsBucketOpsCmd)
	s.storageCmd.AddCommand(s.storageBatchOperationsCmd)
}

func (s *storageCmds) stBoLocationParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	if s.flagStBoLocation == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagStBoLocation), nil
}

func (s *storageCmds) stBoJobName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.stBoLocationParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/jobs/%s", parent, id), nil
}

func (s *storageCmds) stBoBucketOpName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	job, err := s.stBoJobName(s.flagStBoJob)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/bucketOperations/%s", job, id), nil
}

func (s *storageCmds) runStBoJobCancel(cmd *cobra.Command, args []string) error {
	name, err := s.stBoJobName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().StorageBatchOperations(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Jobs.Cancel(name, &storagebatchoperations.CancelJobRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled batch job [%s].\n", args[0])
	return nil
}

func (s *storageCmds) runStBoJobCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.stBoLocationParent()
	if err != nil {
		return err
	}
	body := &storagebatchoperations.Job{}
	if err := loadYAMLOrJSONInto(s.flagStBoConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().StorageBatchOperations(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Jobs.Create(parent, body).JobId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create batch job [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagStBoFormat)
}

func (s *storageCmds) runStBoJobDelete(cmd *cobra.Command, args []string) error {
	name, err := s.stBoJobName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().StorageBatchOperations(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Jobs.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting job: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted batch job [%s].\n", args[0])
	return nil
}

func (s *storageCmds) runStBoJobDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.stBoJobName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().StorageBatchOperations(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing job: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagStBoFormat)
}

func (s *storageCmds) runStBoJobList(cmd *cobra.Command, args []string) error {
	parent, err := s.stBoLocationParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().StorageBatchOperations(ctx)
	if err != nil {
		return err
	}
	var all []*storagebatchoperations.Job
	pageToken := ""
	for {
		call := svc.Projects.Locations.Jobs.List(parent).Context(ctx)
		if s.flagStBoPageSize > 0 {
			call = call.PageSize(s.flagStBoPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing jobs: %w", err)
		}
		all = append(all, resp.Jobs...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagStBoFormat)
}

func (s *storageCmds) runStBoBucketOpDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.stBoBucketOpName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().StorageBatchOperations(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Jobs.BucketOperations.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing bucket-operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagStBoFormat)
}

func (s *storageCmds) runStBoBucketOpList(cmd *cobra.Command, args []string) error {
	parent, err := s.stBoJobName(s.flagStBoJob)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().StorageBatchOperations(ctx)
	if err != nil {
		return err
	}
	var all []*storagebatchoperations.BucketOperation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Jobs.BucketOperations.List(parent).Context(ctx)
		if s.flagStBoPageSize > 0 {
			call = call.PageSize(s.flagStBoPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing bucket-operations: %w", err)
		}
		all = append(all, resp.BucketOperations...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagStBoFormat)
}
