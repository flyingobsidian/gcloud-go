package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	baremetalsolution "google.golang.org/api/baremetalsolution/v2"
)

// --- gcloud bms instances (#1226) ---

func (s *bmsCmds) initBmsInstances() {
	all := []*cobra.Command{
		s.bmsInstDescribeCmd, s.bmsInstListCmd, s.bmsInstUpdateCmd,
		s.bmsInstResetCmd, s.bmsInstStartCmd, s.bmsInstStopCmd,
		s.bmsInstEnableSerialCmd, s.bmsInstDisableSerialCmd, s.bmsInstRenameCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagBmsInstLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagBmsInstFormat, "format", "", "Output format")
	}
	s.bmsInstListCmd.Flags().Int64Var(&s.flagBmsInstPageSize, "page-size", 0, "Maximum results per page")
	s.bmsInstUpdateCmd.Flags().StringVar(&s.flagBmsInstConfigFile, "config-file", "", "YAML/JSON file with fields to update (required)")
	_ = s.bmsInstUpdateCmd.MarkFlagRequired("config-file")
	s.bmsInstUpdateCmd.Flags().StringVar(&s.flagBmsInstUpdateMask, "update-mask", "", "Field mask (defaults to populated fields)")
	s.bmsInstRenameCmd.Flags().StringVar(&s.flagBmsInstNewName, "new-name", "", "New instance id (required)")
	_ = s.bmsInstRenameCmd.MarkFlagRequired("new-name")

	s.bmsInstancesCmd.AddCommand(all...)
	s.bmsCmd.AddCommand(s.bmsInstancesCmd)
}

func (s *bmsCmds) bmsInstName(id string) (string, error) {
	return s.bmsResource(s.flagBmsInstLocation, "instances", id)
}

func (s *bmsCmds) runBmsInstDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.bmsInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Instances.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBmsInstFormat)
}

func (s *bmsCmds) runBmsInstList(cmd *cobra.Command, args []string) error {
	parent, err := s.bmsLocationParent(s.flagBmsInstLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	var all []*baremetalsolution.Instance
	pageToken := ""
	for {
		call := svc.Projects.Locations.Instances.List(parent).Context(ctx)
		if s.flagBmsInstPageSize > 0 {
			call = call.PageSize(s.flagBmsInstPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing instances: %w", err)
		}
		all = append(all, resp.Instances...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagBmsInstFormat)
}

func (s *bmsCmds) runBmsInstUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.bmsInstName(args[0])
	if err != nil {
		return err
	}
	body := &baremetalsolution.Instance{}
	if err := loadYAMLOrJSONInto(s.flagBmsInstConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	mask := s.flagBmsInstUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(body))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	call := svc.Projects.Locations.Instances.Patch(name, body).Context(ctx)
	if mask != "" {
		call = call.UpdateMask(mask)
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update instance [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBmsInstFormat)
}

func (s *bmsCmds) runBmsInstReset(cmd *cobra.Command, args []string) error {
	name, err := s.bmsInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Reset(name, &baremetalsolution.ResetInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resetting instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Reset instance [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBmsInstFormat)
}

func (s *bmsCmds) runBmsInstStart(cmd *cobra.Command, args []string) error {
	name, err := s.bmsInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Start(name, &baremetalsolution.StartInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("starting instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Start instance [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBmsInstFormat)
}

func (s *bmsCmds) runBmsInstStop(cmd *cobra.Command, args []string) error {
	name, err := s.bmsInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.Stop(name, &baremetalsolution.StopInstanceRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("stopping instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Stop instance [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBmsInstFormat)
}

func (s *bmsCmds) runBmsInstEnableSerial(cmd *cobra.Command, args []string) error {
	name, err := s.bmsInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.EnableInteractiveSerialConsole(name, &baremetalsolution.EnableInteractiveSerialConsoleRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("enabling interactive serial console: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Enable interactive serial console on instance [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBmsInstFormat)
}

func (s *bmsCmds) runBmsInstDisableSerial(cmd *cobra.Command, args []string) error {
	name, err := s.bmsInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	op, err := svc.Projects.Locations.Instances.DisableInteractiveSerialConsole(name, &baremetalsolution.DisableInteractiveSerialConsoleRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("disabling interactive serial console: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Disable interactive serial console on instance [%s] initiated (operation: %s).\n", args[0], op.Name)
	return emitFormattedTo(s.inv.stdout, op, s.flagBmsInstFormat)
}

func (s *bmsCmds) runBmsInstRename(cmd *cobra.Command, args []string) error {
	name, err := s.bmsInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.Instances.Rename(name, &baremetalsolution.RenameInstanceRequest{NewInstanceId: s.flagBmsInstNewName}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("renaming instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Renamed instance [%s] to [%s].\n", args[0], s.flagBmsInstNewName)
	return emitFormattedTo(s.inv.stdout, got, s.flagBmsInstFormat)
}
