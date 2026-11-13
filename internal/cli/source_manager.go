package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	securesourcemanager "google.golang.org/api/securesourcemanager/v1"
)

// --- gcloud source-manager (#925-#928) ---

// --- Instances ---

// --- Locations ---

// --- Operations ---

// --- Repos ---

func (s *sourceManagerCmds) initSourceManager() {
	// Instances flags.
	for _, c := range []*cobra.Command{
		s.sourceManagerInstancesCreateCmd, s.sourceManagerInstancesDeleteCmd,
		s.sourceManagerInstancesDescribeCmd, s.sourceManagerInstancesUpdateCmd,
	} {
		c.Flags().StringVar(&s.flagSMRegion, "region", "", "Region containing the instance")
	}
	s.sourceManagerInstancesCreateCmd.Flags().StringVar(&s.flagSMKmsKey, "kms-key", "", "Customer-managed encryption key")
	s.sourceManagerInstancesCreateCmd.Flags().StringVar(&s.flagSMFile, "config-from-file", "", "YAML/JSON file with the full instance spec")

	s.sourceManagerInstancesListCmd.Flags().StringVar(&s.flagSMRegion, "region", "", "Region to list instances in")
	s.sourceManagerInstancesListCmd.Flags().Int64Var(&s.flagSMPageSize, "page-size", 0, "Number of results per page")
	s.sourceManagerInstancesListCmd.Flags().Int64Var(&s.flagSMLimit, "limit", 0, "Maximum number of results to return")
	s.sourceManagerInstancesListCmd.Flags().StringVar(&s.flagSMFilter, "filter", "", "Server-side filter expression")
	s.sourceManagerInstancesListCmd.Flags().StringVar(&s.flagSMOrderBy, "order-by", "", "Server-side ordering expression")

	s.sourceManagerInstancesUpdateCmd.Flags().StringVar(&s.flagSMFile, "config-from-file", "", "YAML/JSON file with the instance patch")

	// Locations flags.
	s.sourceManagerLocationsListCmd.Flags().Int64Var(&s.flagSMPageSize, "page-size", 0, "Number of results per page")
	s.sourceManagerLocationsListCmd.Flags().Int64Var(&s.flagSMLimit, "limit", 0, "Maximum number of results to return")
	s.sourceManagerLocationsListCmd.Flags().StringVar(&s.flagSMFilter, "filter", "", "Server-side filter expression")

	// Operations flags. Operations may be identified by short id + --region or fully-qualified name.
	for _, c := range []*cobra.Command{
		s.sourceManagerOperationsCancelCmd, s.sourceManagerOperationsDeleteCmd,
		s.sourceManagerOperationsDescribeCmd, s.sourceManagerOperationsWaitCmd,
	} {
		c.Flags().StringVar(&s.flagSMRegion, "region", "", "Region containing the operation")
	}
	s.sourceManagerOperationsListCmd.Flags().StringVar(&s.flagSMRegion, "region", "", "Region to list operations in")
	s.sourceManagerOperationsListCmd.Flags().Int64Var(&s.flagSMPageSize, "page-size", 0, "Number of results per page")
	s.sourceManagerOperationsListCmd.Flags().Int64Var(&s.flagSMLimit, "limit", 0, "Maximum number of results to return")
	s.sourceManagerOperationsListCmd.Flags().StringVar(&s.flagSMFilter, "filter", "", "Server-side filter expression")

	// Repos flags.
	for _, c := range []*cobra.Command{
		s.sourceManagerReposCreateCmd, s.sourceManagerReposDeleteCmd,
		s.sourceManagerReposDescribeCmd, s.sourceManagerReposUpdateCmd,
		s.sourceManagerReposGetIamCmd, s.sourceManagerReposSetIamCmd,
		s.sourceManagerReposAddIamCmd, s.sourceManagerReposRemoveIamCmd,
	} {
		c.Flags().StringVar(&s.flagSMRegion, "region", "", "Region containing the repository")
	}
	s.sourceManagerReposCreateCmd.Flags().StringVar(&s.flagSMInstance, "instance", "", "Instance short name that hosts the repository (required)")
	s.sourceManagerReposCreateCmd.Flags().StringVar(&s.flagSMDescription, "description", "", "Repository description")
	s.sourceManagerReposCreateCmd.Flags().StringVar(&s.flagSMDefaultBranch, "default-branch", "", "Default branch for the initial repository configuration")
	s.sourceManagerReposCreateCmd.MarkFlagRequired("instance")

	s.sourceManagerReposUpdateCmd.Flags().StringVar(&s.flagSMDescription, "description", "", "New description for the repository")

	s.sourceManagerReposListCmd.Flags().StringVar(&s.flagSMRegion, "region", "", "Region to list repositories in")
	s.sourceManagerReposListCmd.Flags().StringVar(&s.flagSMInstance, "instance", "", "Only list repositories on this instance")
	s.sourceManagerReposListCmd.Flags().Int64Var(&s.flagSMPageSize, "page-size", 0, "Number of results per page")
	s.sourceManagerReposListCmd.Flags().Int64Var(&s.flagSMLimit, "limit", 0, "Maximum number of results to return")
	s.sourceManagerReposListCmd.Flags().StringVar(&s.flagSMFilter, "filter", "", "Server-side filter expression")
	s.sourceManagerReposListCmd.Flags().StringVar(&s.flagSMOrderBy, "order-by", "", "Server-side ordering expression")

	s.sourceManagerReposAddIamCmd.Flags().StringVar(&s.flagSMIamRole, "role", "", "IAM role to grant (required)")
	s.sourceManagerReposAddIamCmd.Flags().StringVar(&s.flagSMIamMember, "member", "", "IAM member to add (required)")
	s.sourceManagerReposAddIamCmd.MarkFlagRequired("role")
	s.sourceManagerReposAddIamCmd.MarkFlagRequired("member")
	s.sourceManagerReposRemoveIamCmd.Flags().StringVar(&s.flagSMIamRole, "role", "", "IAM role to revoke (required)")
	s.sourceManagerReposRemoveIamCmd.Flags().StringVar(&s.flagSMIamMember, "member", "", "IAM member to remove (required)")
	s.sourceManagerReposRemoveIamCmd.MarkFlagRequired("role")
	s.sourceManagerReposRemoveIamCmd.MarkFlagRequired("member")

	s.sourceManagerInstancesCmd.AddCommand(
		s.sourceManagerInstancesCreateCmd, s.sourceManagerInstancesDeleteCmd,
		s.sourceManagerInstancesDescribeCmd, s.sourceManagerInstancesListCmd,
		s.sourceManagerInstancesUpdateCmd,
	)
	s.sourceManagerLocationsCmd.AddCommand(s.sourceManagerLocationsDescribeCmd, s.sourceManagerLocationsListCmd)
	s.sourceManagerOperationsCmd.AddCommand(
		s.sourceManagerOperationsCancelCmd, s.sourceManagerOperationsDeleteCmd,
		s.sourceManagerOperationsDescribeCmd, s.sourceManagerOperationsListCmd,
		s.sourceManagerOperationsWaitCmd,
	)
	s.sourceManagerReposCmd.AddCommand(
		s.sourceManagerReposCreateCmd, s.sourceManagerReposDeleteCmd,
		s.sourceManagerReposDescribeCmd, s.sourceManagerReposListCmd,
		s.sourceManagerReposUpdateCmd, s.sourceManagerReposGetIamCmd,
		s.sourceManagerReposSetIamCmd, s.sourceManagerReposAddIamCmd,
		s.sourceManagerReposRemoveIamCmd,
	)
	s.sourceManagerCmd.AddCommand(
		s.sourceManagerInstancesCmd, s.sourceManagerLocationsCmd,
		s.sourceManagerOperationsCmd, s.sourceManagerReposCmd,
	)

}

// --- Helpers ---

func (s *sourceManagerCmds) smRequireRegion() (string, string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", "", err
	}
	if s.flagSMRegion == "" {
		return "", "", fmt.Errorf("--region is required")
	}
	return project, s.flagSMRegion, nil
}

func smParent(project, region string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, region)
}

func smInstanceName(id, project, region string) string {
	return fmt.Sprintf("%s/instances/%s", smParent(project, region), id)
}

func smRepoName(id, project, region string) string {
	return fmt.Sprintf("%s/repositories/%s", smParent(project, region), id)
}

func smOperationName(id, project, region string) string {
	if id != "" && id[0] == '/' {
		return id[1:]
	}
	// Accept either a bare operation id or an already-qualified path.
	if len(id) > len("projects/") && id[:len("projects/")] == "projects/" {
		return id
	}
	return fmt.Sprintf("%s/operations/%s", smParent(project, region), id)
}

// --- Instances impl ---

func (s *sourceManagerCmds) runSMInstancesCreate(cmd *cobra.Command, args []string) error {
	project, region, err := s.smRequireRegion()
	if err != nil {
		return err
	}
	inst := &securesourcemanager.Instance{}
	if s.flagSMFile != "" {
		if err := loadYAMLOrJSONInto(s.flagSMFile, inst); err != nil {
			return err
		}
	}
	if s.flagSMKmsKey != "" {
		inst.KmsKey = s.flagSMKmsKey
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Create(smParent(project, region), inst).InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for instance [%s] (operation: %s)\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, "")
}

func (s *sourceManagerCmds) runSMInstancesDelete(cmd *cobra.Command, args []string) error {
	project, region, err := s.smRequireRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Delete(smInstanceName(args[0], project, region)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for instance [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *sourceManagerCmds) runSMInstancesDescribe(cmd *cobra.Command, args []string) error {
	project, region, err := s.smRequireRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	inst, err := svc.Projects.Locations.Instances.Get(smInstanceName(args[0], project, region)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, inst, "")
}

func (s *sourceManagerCmds) runSMInstancesList(cmd *cobra.Command, args []string) error {
	project, region, err := s.smRequireRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	var all []*securesourcemanager.Instance
	pageToken := ""
	for {
		call := svc.Projects.Locations.Instances.List(smParent(project, region)).Context(ctx)
		if s.flagSMPageSize > 0 {
			call = call.PageSize(s.flagSMPageSize)
		}
		if s.flagSMFilter != "" {
			call = call.Filter(s.flagSMFilter)
		}
		if s.flagSMOrderBy != "" {
			call = call.OrderBy(s.flagSMOrderBy)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing instances: %w", err)
		}
		all = append(all, resp.Instances...)
		if s.flagSMLimit > 0 && int64(len(all)) >= s.flagSMLimit {
			all = all[:s.flagSMLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, "")
}

func runSMInstancesUpdate(cmd *cobra.Command, args []string) error {
	// The Secure Source Manager v1 Instances resource has no server-side Patch;
	// the only mutable path is to read, delete, and re-create with the full
	// spec. Rather than silently doing that here we return a clear error so
	// operators know the operation is unsupported by the underlying API.
	return fmt.Errorf("source-manager instances update: not supported by the Secure Source Manager v1 API")
}

// --- Locations impl ---

func (s *sourceManagerCmds) runSMLocationsDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	loc, err := svc.Projects.Locations.Get(fmt.Sprintf("projects/%s/locations/%s", project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing location: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, loc, "")
}

func (s *sourceManagerCmds) runSMLocationsList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	var all []*securesourcemanager.Location
	pageToken := ""
	for {
		call := svc.Projects.Locations.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
		if s.flagSMPageSize > 0 {
			call = call.PageSize(int64(s.flagSMPageSize))
		}
		if s.flagSMFilter != "" {
			call = call.Filter(s.flagSMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing locations: %w", err)
		}
		all = append(all, resp.Locations...)
		if s.flagSMLimit > 0 && int64(len(all)) >= s.flagSMLimit {
			all = all[:s.flagSMLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, "")
}

// --- Operations impl ---

func (s *sourceManagerCmds) runSMOperationsCancel(cmd *cobra.Command, args []string) error {
	project, region, err := s.smRequireRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Cancel(smOperationName(args[0], project, region), &securesourcemanager.CancelOperationRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancelling operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Cancelled operation [%s].\n", args[0])
	return nil
}

func (s *sourceManagerCmds) runSMOperationsDelete(cmd *cobra.Command, args []string) error {
	project, region, err := s.smRequireRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.Operations.Delete(smOperationName(args[0], project, region)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting operation: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted operation [%s].\n", args[0])
	return nil
}

func (s *sourceManagerCmds) runSMOperationsDescribe(cmd *cobra.Command, args []string) error {
	project, region, err := s.smRequireRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Operations.Get(smOperationName(args[0], project, region)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing operation: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, op, "")
}

func (s *sourceManagerCmds) runSMOperationsList(cmd *cobra.Command, args []string) error {
	project, region, err := s.smRequireRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	var all []*securesourcemanager.Operation
	pageToken := ""
	for {
		call := svc.Projects.Locations.Operations.List(smParent(project, region)).Context(ctx)
		if s.flagSMPageSize > 0 {
			call = call.PageSize(int64(s.flagSMPageSize))
		}
		if s.flagSMFilter != "" {
			call = call.Filter(s.flagSMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing operations: %w", err)
		}
		all = append(all, resp.Operations...)
		if s.flagSMLimit > 0 && int64(len(all)) >= s.flagSMLimit {
			all = all[:s.flagSMLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, "")
}

func (s *sourceManagerCmds) runSMOperationsWait(cmd *cobra.Command, args []string) error {
	project, region, err := s.smRequireRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	name := smOperationName(args[0], project, region)
	op, err := pollSMOperation(ctx, svc, name)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, op, "")
}

// --- Repos impl ---

func (s *sourceManagerCmds) runSMReposCreate(cmd *cobra.Command, args []string) error {
	project, region, err := s.smRequireRegion()
	if err != nil {
		return err
	}
	repo := &securesourcemanager.Repository{
		Description: s.flagSMDescription,
		Instance:    smInstanceName(s.flagSMInstance, project, region),
	}
	if s.flagSMDefaultBranch != "" {
		repo.InitialConfig = &securesourcemanager.InitialConfig{DefaultBranch: s.flagSMDefaultBranch}
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Repositories.Create(smParent(project, region), repo).RepositoryId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating repository: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for repository [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *sourceManagerCmds) runSMReposDelete(cmd *cobra.Command, args []string) error {
	project, region, err := s.smRequireRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Repositories.Delete(smRepoName(args[0], project, region)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting repository: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for repository [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *sourceManagerCmds) runSMReposDescribe(cmd *cobra.Command, args []string) error {
	project, region, err := s.smRequireRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	repo, err := svc.Projects.Locations.Repositories.Get(smRepoName(args[0], project, region)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing repository: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, repo, "")
}

func (s *sourceManagerCmds) runSMReposList(cmd *cobra.Command, args []string) error {
	project, region, err := s.smRequireRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	var all []*securesourcemanager.Repository
	pageToken := ""
	for {
		call := svc.Projects.Locations.Repositories.List(smParent(project, region)).Context(ctx)
		if s.flagSMInstance != "" {
			call = call.Instance(smInstanceName(s.flagSMInstance, project, region))
		}
		if s.flagSMPageSize > 0 {
			call = call.PageSize(s.flagSMPageSize)
		}
		if s.flagSMFilter != "" {
			call = call.Filter(s.flagSMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing repositories: %w", err)
		}
		all = append(all, resp.Repositories...)
		if s.flagSMLimit > 0 && int64(len(all)) >= s.flagSMLimit {
			all = all[:s.flagSMLimit]
			break
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, "")
}

func (s *sourceManagerCmds) runSMReposUpdate(cmd *cobra.Command, args []string) error {
	project, region, err := s.smRequireRegion()
	if err != nil {
		return err
	}
	repo := &securesourcemanager.Repository{}
	mask := ""
	if s.flagSMDescription != "" {
		repo.Description = s.flagSMDescription
		mask = "description"
	}
	if mask == "" {
		return fmt.Errorf("nothing to update: pass --description")
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Repositories.Patch(smRepoName(args[0], project, region), repo).UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating repository: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for repository [%s] (operation: %s)\n", args[0], op.Name)
	return nil
}

func (s *sourceManagerCmds) runSMReposGetIam(cmd *cobra.Command, args []string) error {
	project, region, err := s.smRequireRegion()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Repositories.GetIamPolicy(smRepoName(args[0], project, region)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, "")
}

func (s *sourceManagerCmds) runSMReposSetIam(cmd *cobra.Command, args []string) error {
	project, region, err := s.smRequireRegion()
	if err != nil {
		return err
	}
	policy := &securesourcemanager.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Repositories.SetIamPolicy(smRepoName(args[0], project, region), &securesourcemanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *sourceManagerCmds) runSMReposAddIam(cmd *cobra.Command, args []string) error {
	return s.smReposModifyIam(args[0], func(p *securesourcemanager.Policy) {
		for _, b := range p.Bindings {
			if b.Role == s.flagSMIamRole {
				for _, m := range b.Members {
					if m == s.flagSMIamMember {
						return
					}
				}
				b.Members = append(b.Members, s.flagSMIamMember)
				return
			}
		}
		p.Bindings = append(p.Bindings, &securesourcemanager.Binding{Role: s.flagSMIamRole, Members: []string{s.flagSMIamMember}})
	})
}

func (s *sourceManagerCmds) runSMReposRemoveIam(cmd *cobra.Command, args []string) error {
	return s.smReposModifyIam(args[0], func(p *securesourcemanager.Policy) {
		for _, b := range p.Bindings {
			if b.Role != s.flagSMIamRole {
				continue
			}
			out := b.Members[:0]
			for _, m := range b.Members {
				if m != s.flagSMIamMember {
					out = append(out, m)
				}
			}
			b.Members = out
		}
	})
}

func (s *sourceManagerCmds) smReposModifyIam(id string, mutate func(*securesourcemanager.Policy)) error {
	project, region, err := s.smRequireRegion()
	if err != nil {
		return err
	}
	ctx := s.inv.context()
	svc, err := s.inv.clients().SecureSourceManager(ctx)
	if err != nil {
		return err
	}
	resource := smRepoName(id, project, region)
	policy, err := svc.Projects.Locations.Repositories.GetIamPolicy(resource).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	mutate(policy)
	got, err := svc.Projects.Locations.Repositories.SetIamPolicy(resource, &securesourcemanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// pollSMOperation blocks until the given operation is Done or the context is
// cancelled. Callers with a wall-clock deadline should wrap the context.
func pollSMOperation(ctx context.Context, svc *securesourcemanager.Service, name string) (*securesourcemanager.Operation, error) {
	for {
		op, err := svc.Projects.Locations.Operations.Get(name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation: %w", err)
		}
		if op.Done {
			if op.Error != nil {
				return op, fmt.Errorf("operation %s failed: %s", name, op.Error.Message)
			}
			return op, nil
		}
		select {
		case <-ctx.Done():
			return op, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

// sourceManagerCmds holds the commands and flag values of one invocation's "sourceManager"
// command group, so invocations don't share state (#2070).
type sourceManagerCmds struct {
	inv *invocation

	sourceManagerCmd                   *cobra.Command
	flagSMRegion                       string
	flagSMInstance                     string
	flagSMKmsKey                       string
	flagSMDescription                  string
	flagSMDefaultBranch                string
	flagSMFile                         string
	flagSMIamRole                      string
	flagSMIamMember                    string
	flagSMPageSize                     int64
	flagSMLimit                        int64
	flagSMFilter                       string
	flagSMOrderBy                      string
	sourceManagerInstancesCmd          *cobra.Command
	sourceManagerInstancesCreateCmd    *cobra.Command
	sourceManagerInstancesDeleteCmd    *cobra.Command
	sourceManagerInstancesDescribeCmd  *cobra.Command
	sourceManagerInstancesListCmd      *cobra.Command
	sourceManagerInstancesUpdateCmd    *cobra.Command
	sourceManagerLocationsCmd          *cobra.Command
	sourceManagerLocationsDescribeCmd  *cobra.Command
	sourceManagerLocationsListCmd      *cobra.Command
	sourceManagerOperationsCmd         *cobra.Command
	sourceManagerOperationsCancelCmd   *cobra.Command
	sourceManagerOperationsDeleteCmd   *cobra.Command
	sourceManagerOperationsDescribeCmd *cobra.Command
	sourceManagerOperationsListCmd     *cobra.Command
	sourceManagerOperationsWaitCmd     *cobra.Command
	sourceManagerReposCmd              *cobra.Command
	sourceManagerReposCreateCmd        *cobra.Command
	sourceManagerReposDeleteCmd        *cobra.Command
	sourceManagerReposDescribeCmd      *cobra.Command
	sourceManagerReposListCmd          *cobra.Command
	sourceManagerReposUpdateCmd        *cobra.Command
	sourceManagerReposGetIamCmd        *cobra.Command
	sourceManagerReposSetIamCmd        *cobra.Command
	sourceManagerReposAddIamCmd        *cobra.Command
	sourceManagerReposRemoveIamCmd     *cobra.Command
}

// newSourceManagerCmd returns the "sourceManager" command group for inv.
func newSourceManagerCmd(inv *invocation) *cobra.Command {
	return newSourceManagerCmds(inv).sourceManagerCmd
}

// newSourceManagerCmds builds the "sourceManager" command group for inv and returns its state.
func newSourceManagerCmds(inv *invocation) *sourceManagerCmds {
	s := &sourceManagerCmds{inv: inv}
	s.sourceManagerCmd = &cobra.Command{
		Use:   "source-manager",
		Short: "Manage Secure Source Manager",
	}
	s.sourceManagerInstancesCmd = &cobra.Command{
		Use:   "instances",
		Short: "Manage Secure Source Manager instances",
	}
	s.sourceManagerInstancesCreateCmd = &cobra.Command{
		Use:   "create INSTANCE",
		Short: "Create a Secure Source Manager instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSMInstancesCreate,
	}
	s.sourceManagerInstancesDeleteCmd = &cobra.Command{
		Use:   "delete INSTANCE",
		Short: "Delete a Secure Source Manager instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSMInstancesDelete,
	}
	s.sourceManagerInstancesDescribeCmd = &cobra.Command{
		Use:   "describe INSTANCE",
		Short: "Describe a Secure Source Manager instance",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSMInstancesDescribe,
	}
	s.sourceManagerInstancesListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Secure Source Manager instances",
		Args:  cobra.NoArgs,
		RunE:  s.runSMInstancesList,
	}
	s.sourceManagerInstancesUpdateCmd = &cobra.Command{
		Use:   "update INSTANCE",
		Short: "Update a Secure Source Manager instance",
		Args:  cobra.ExactArgs(1),
		RunE:  runSMInstancesUpdate,
	}
	s.sourceManagerLocationsCmd = &cobra.Command{
		Use:   "locations",
		Short: "Manage Secure Source Manager locations",
	}
	s.sourceManagerLocationsDescribeCmd = &cobra.Command{
		Use:   "describe LOCATION",
		Short: "Describe a Secure Source Manager location",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSMLocationsDescribe,
	}
	s.sourceManagerLocationsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Secure Source Manager locations",
		Args:  cobra.NoArgs,
		RunE:  s.runSMLocationsList,
	}
	s.sourceManagerOperationsCmd = &cobra.Command{
		Use:   "operations",
		Short: "Manage Secure Source Manager operations",
	}
	s.sourceManagerOperationsCancelCmd = &cobra.Command{
		Use:   "cancel OPERATION",
		Short: "Cancel a Secure Source Manager operation",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSMOperationsCancel,
	}
	s.sourceManagerOperationsDeleteCmd = &cobra.Command{
		Use:   "delete OPERATION",
		Short: "Delete a Secure Source Manager operation",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSMOperationsDelete,
	}
	s.sourceManagerOperationsDescribeCmd = &cobra.Command{
		Use:   "describe OPERATION",
		Short: "Describe a Secure Source Manager operation",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSMOperationsDescribe,
	}
	s.sourceManagerOperationsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Secure Source Manager operations",
		Args:  cobra.NoArgs,
		RunE:  s.runSMOperationsList,
	}
	s.sourceManagerOperationsWaitCmd = &cobra.Command{
		Use:   "wait OPERATION",
		Short: "Poll a Secure Source Manager operation until it completes",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSMOperationsWait,
	}
	s.sourceManagerReposCmd = &cobra.Command{
		Use:   "repos",
		Short: "Manage Secure Source Manager repositories",
	}
	s.sourceManagerReposCreateCmd = &cobra.Command{
		Use:   "create REPO",
		Short: "Create a Secure Source Manager repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSMReposCreate,
	}
	s.sourceManagerReposDeleteCmd = &cobra.Command{
		Use:   "delete REPO",
		Short: "Delete a Secure Source Manager repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSMReposDelete,
	}
	s.sourceManagerReposDescribeCmd = &cobra.Command{
		Use:   "describe REPO",
		Short: "Describe a Secure Source Manager repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSMReposDescribe,
	}
	s.sourceManagerReposListCmd = &cobra.Command{
		Use:   "list",
		Short: "List Secure Source Manager repositories",
		Args:  cobra.NoArgs,
		RunE:  s.runSMReposList,
	}
	s.sourceManagerReposUpdateCmd = &cobra.Command{
		Use:   "update REPO",
		Short: "Update a Secure Source Manager repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSMReposUpdate,
	}
	s.sourceManagerReposGetIamCmd = &cobra.Command{
		Use:   "get-iam-policy REPO",
		Short: "Print the IAM policy for a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSMReposGetIam,
	}
	s.sourceManagerReposSetIamCmd = &cobra.Command{
		Use:   "set-iam-policy REPO POLICY_FILE",
		Short: "Set the IAM policy for a repository",
		Args:  cobra.ExactArgs(2),
		RunE:  s.runSMReposSetIam,
	}
	s.sourceManagerReposAddIamCmd = &cobra.Command{
		Use:   "add-iam-policy-binding REPO",
		Short: "Add an IAM policy binding to a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSMReposAddIam,
	}
	s.sourceManagerReposRemoveIamCmd = &cobra.Command{
		Use:   "remove-iam-policy-binding REPO",
		Short: "Remove an IAM policy binding from a repository",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runSMReposRemoveIam,
	}
	s.initSourceManager()
	return s
}
