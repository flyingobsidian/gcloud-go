package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	notebooksv1 "google.golang.org/api/notebooks/v1"
)

// --- gcloud notebooks instances (#1062) ---

func (s *notebooksCmds) initNotebooksInstances() {
	all := []*cobra.Command{
		s.notebooksInstCreateCmd, s.notebooksInstDeleteCmd, s.notebooksInstDescribeCmd,
		s.notebooksInstListCmd, s.notebooksInstUpdateCmd,
		s.notebooksInstAddIamCmd, s.notebooksInstRemoveIamCmd,
		s.notebooksInstGetIamCmd, s.notebooksInstSetIamCmd,
		s.notebooksInstDiagnoseCmd, s.notebooksInstGetHealthCmd, s.notebooksInstIsUpgradeableCmd,
		s.notebooksInstMigrateCmd, s.notebooksInstRegisterCmd, s.notebooksInstResetCmd,
		s.notebooksInstRollbackCmd, s.notebooksInstStartCmd, s.notebooksInstStopCmd,
		s.notebooksInstUpgradeCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagNotebooksInstLocation, "location", "", "Notebook location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagNotebooksInstFormat, "format", "", "Output format")
	}
	// Endpoints that consume a --config-file (required):
	for _, c := range []*cobra.Command{
		s.notebooksInstCreateCmd, s.notebooksInstUpdateCmd, s.notebooksInstDiagnoseCmd,
		s.notebooksInstRegisterCmd, s.notebooksInstRollbackCmd,
	} {
		c.Flags().StringVar(&s.flagNotebooksInstConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	// Migrate takes an optional --config-file (the request body may be empty).
	s.notebooksInstMigrateCmd.Flags().StringVar(&s.flagNotebooksInstConfigFile, "config-file", "",
		"Path to a YAML/JSON file with the MigrateInstanceRequest body (optional)")
	s.notebooksInstUpdateCmd.Flags().StringVar(&s.flagNotebooksInstUpdateMask, "update-mask", "",
		"Comma-separated list of InstanceConfig fields to update (defaults to every populated field)")
	s.notebooksInstListCmd.Flags().Int64Var(&s.flagNotebooksInstPageSize, "page-size", 0, "Maximum results per page")
	for _, c := range []*cobra.Command{s.notebooksInstAddIamCmd, s.notebooksInstRemoveIamCmd} {
		notebooksIamMemberFlags(c, &s.flagNotebooksInstIamMember, &s.flagNotebooksInstIamRole,
			&s.flagNotebooksInstIamCondExpr, &s.flagNotebooksInstIamCondT, &s.flagNotebooksInstIamCondD)
	}
	s.notebooksInstRemoveIamCmd.Flags().BoolVar(&s.flagNotebooksInstIamAllCond, "all", false,
		"Remove the member from all bindings for the role, regardless of condition")

	s.notebooksInstCmd.AddCommand(all...)
	s.notebooksCmd.AddCommand(s.notebooksInstCmd)
}

func (s *notebooksCmds) notebooksInstParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagNotebooksInstLocation), nil
}

func (s *notebooksCmds) notebooksInstName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.notebooksInstParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/instances/%s", parent, id), nil
}

func (s *notebooksCmds) runNotebooksInstCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.notebooksInstParent()
	if err != nil {
		return err
	}
	body := &notebooksv1.Instance{}
	if err := loadYAMLOrJSONInto(s.flagNotebooksInstConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Create(parent, body).InstanceId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating notebook instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksInstFormat)
}

func (s *notebooksCmds) runNotebooksInstDelete(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting notebook instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksInstFormat)
}

func (s *notebooksCmds) runNotebooksInstDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Instances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing notebook instance: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNotebooksInstFormat)
}

func (s *notebooksCmds) runNotebooksInstList(cmd *cobra.Command, args []string) error {
	parent, err := s.notebooksInstParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	var all []*notebooksv1.Instance
	pageToken := ""
	for {
		call := svc.Projects.Locations.Instances.List(parent).Context(ctx)
		if s.flagNotebooksInstPageSize > 0 {
			call = call.PageSize(s.flagNotebooksInstPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing notebook instances: %w", err)
		}
		all = append(all, resp.Instances...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagNotebooksInstFormat)
}

// runNotebooksInstUpdate calls the UpdateConfig endpoint (a PATCH against
// v1/{name}:updateConfig) with an InstanceConfig loaded from --config-file.
// The v1 API does not expose a top-level Instance PATCH; UpdateConfig is the
// generalized configuration-update endpoint.
func (s *notebooksCmds) runNotebooksInstUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksInstName(args[0])
	if err != nil {
		return err
	}
	// Allow the config file to be either a bare InstanceConfig or a full
	// UpdateInstanceConfigRequest ({"config": {...}}).
	req := &notebooksv1.UpdateInstanceConfigRequest{}
	if err := loadYAMLOrJSONInto(s.flagNotebooksInstConfigFile, req); err != nil {
		return err
	}
	if req.Config == nil {
		ic := &notebooksv1.InstanceConfig{}
		if err := loadYAMLOrJSONInto(s.flagNotebooksInstConfigFile, ic); err != nil {
			return err
		}
		req.Config = ic
	}
	_ = s.flagNotebooksInstUpdateMask // accepted but the UpdateConfig endpoint infers the mask from the request body
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.UpdateConfig(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating notebook instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksInstFormat)
}

func (s *notebooksCmds) runNotebooksInstGetIam(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Instances.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, policy, s.flagNotebooksInstFormat)
}

func (s *notebooksCmds) runNotebooksInstSetIam(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksInstName(args[0])
	if err != nil {
		return err
	}
	policy := &notebooksv1.Policy{}
	if err := loadYAMLOrJSONInto(args[1], policy); err != nil {
		return err
	}
	policy.Version = 3
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	updated, err := svc.Projects.Locations.Instances.SetIamPolicy(name,
		&notebooksv1.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.notebooksUpdatedIam(fmt.Sprintf("notebook instance [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagNotebooksInstFormat)
}

func (s *notebooksCmds) runNotebooksInstAddIam(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Instances.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	notebooksIamAddBinding(policy, s.flagNotebooksInstIamRole, s.flagNotebooksInstIamMember,
		notebooksIamBuildCondition(s.flagNotebooksInstIamCondExpr, s.flagNotebooksInstIamCondT, s.flagNotebooksInstIamCondD))
	policy.Version = 3
	updated, err := svc.Projects.Locations.Instances.SetIamPolicy(name,
		&notebooksv1.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.notebooksUpdatedIam(fmt.Sprintf("notebook instance [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagNotebooksInstFormat)
}

func (s *notebooksCmds) runNotebooksInstRemoveIam(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	policy, err := svc.Projects.Locations.Instances.GetIamPolicy(name).OptionsRequestedPolicyVersion(3).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting IAM policy: %w", err)
	}
	if !notebooksIamRemoveBinding(policy, s.flagNotebooksInstIamRole, s.flagNotebooksInstIamMember,
		notebooksIamBuildCondition(s.flagNotebooksInstIamCondExpr, s.flagNotebooksInstIamCondT, s.flagNotebooksInstIamCondD),
		s.flagNotebooksInstIamAllCond) {
		return fmt.Errorf("policy binding not found for role [%s] and member [%s]",
			s.flagNotebooksInstIamRole, s.flagNotebooksInstIamMember)
	}
	updated, err := svc.Projects.Locations.Instances.SetIamPolicy(name,
		&notebooksv1.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting IAM policy: %w", err)
	}
	s.notebooksUpdatedIam(fmt.Sprintf("notebook instance [%s]", args[0]))
	return emitFormattedTo(s.inv.stdout, updated, s.flagNotebooksInstFormat)
}

func (s *notebooksCmds) runNotebooksInstDiagnose(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksInstName(args[0])
	if err != nil {
		return err
	}
	req := &notebooksv1.DiagnoseInstanceRequest{}
	if err := loadYAMLOrJSONInto(s.flagNotebooksInstConfigFile, req); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Diagnose(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("diagnosing notebook instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Diagnose request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksInstFormat)
}

func (s *notebooksCmds) runNotebooksInstGetHealth(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Instances.GetInstanceHealth(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting notebook instance health: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNotebooksInstFormat)
}

func (s *notebooksCmds) runNotebooksInstIsUpgradeable(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Instances.IsUpgradeable(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("checking notebook instance upgradeability: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagNotebooksInstFormat)
}

func (s *notebooksCmds) runNotebooksInstMigrate(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksInstName(args[0])
	if err != nil {
		return err
	}
	req := &notebooksv1.MigrateInstanceRequest{}
	if s.flagNotebooksInstConfigFile != "" {
		if _, err := os.Stat(s.flagNotebooksInstConfigFile); err == nil {
			if err := loadYAMLOrJSONInto(s.flagNotebooksInstConfigFile, req); err != nil {
				return err
			}
		}
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Migrate(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("migrating notebook instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Migrate request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksInstFormat)
}

func (s *notebooksCmds) runNotebooksInstRegister(cmd *cobra.Command, args []string) error {
	parent, err := s.notebooksInstParent()
	if err != nil {
		return err
	}
	req := &notebooksv1.RegisterInstanceRequest{}
	if err := loadYAMLOrJSONInto(s.flagNotebooksInstConfigFile, req); err != nil {
		return err
	}
	if req.InstanceId == "" {
		req.InstanceId = args[0]
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Register(parent, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("registering notebook instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Register request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksInstFormat)
}

func (s *notebooksCmds) runNotebooksInstReset(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Reset(name, &notebooksv1.ResetInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting notebook instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Reset request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksInstFormat)
}

func (s *notebooksCmds) runNotebooksInstRollback(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksInstName(args[0])
	if err != nil {
		return err
	}
	req := &notebooksv1.RollbackInstanceRequest{}
	if err := loadYAMLOrJSONInto(s.flagNotebooksInstConfigFile, req); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Rollback(name, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("rolling back notebook instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Rollback request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksInstFormat)
}

func (s *notebooksCmds) runNotebooksInstStart(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Start(name, &notebooksv1.StartInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting notebook instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Start request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksInstFormat)
}

func (s *notebooksCmds) runNotebooksInstStop(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Stop(name, &notebooksv1.StopInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping notebook instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Stop request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksInstFormat)
}

func (s *notebooksCmds) runNotebooksInstUpgrade(cmd *cobra.Command, args []string) error {
	name, err := s.notebooksInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().NotebooksV1(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Upgrade(name, &notebooksv1.UpgradeInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("upgrading notebook instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Upgrade request issued for notebook instance [%s] (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagNotebooksInstFormat)
}
