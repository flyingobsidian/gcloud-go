package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (s *configCmds) initConfigConfigurations() {
	s.configConfigurationsCmd.AddCommand(s.configConfigurationsCreateCmd)
	s.configConfigurationsCmd.AddCommand(s.configConfigurationsListCmd)
	s.configConfigurationsCmd.AddCommand(s.configConfigurationsActivateCmd)
	s.configConfigurationsCmd.AddCommand(s.configConfigurationsDeleteCmd)
	s.configConfigurationsCmd.AddCommand(s.configConfigurationsDescribeCmd)
	s.configCmd.AddCommand(s.configConfigurationsCmd)
}

func (s *configCmds) runConfigConfigurationsCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	if err := s.inv.env.CreateConfiguration(name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Created configuration [%s].\n", name)
	return nil
}

func (s *configCmds) runConfigConfigurationsList(cmd *cobra.Command, args []string) error {
	names, err := s.inv.env.ListConfigurations()
	if err != nil {
		return fmt.Errorf("listing configurations: %w", err)
	}

	active := s.inv.env.ActiveConfigName()

	printTableHeadingTo(s.inv.stdout, len(names), "NAME      IS_ACTIVE\n")
	for _, name := range names {
		marker := "False"
		if name == active {
			marker = "True"
		}
		fmt.Fprintf(s.inv.stdout, "%-10s%s\n", name, marker)
	}
	return nil
}

func (s *configCmds) runConfigConfigurationsActivate(cmd *cobra.Command, args []string) error {
	name := args[0]
	if err := s.inv.env.ActivateConfiguration(name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Activated configuration [%s].\n", name)
	return nil
}

func (s *configCmds) runConfigConfigurationsDelete(cmd *cobra.Command, args []string) error {
	name := args[0]
	if err := s.inv.env.DeleteConfiguration(name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Deleted configuration [%s].\n", name)
	return nil
}

func (s *configCmds) runConfigConfigurationsDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]
	props, err := s.inv.env.LoadNamed(name)
	if err != nil {
		return fmt.Errorf("loading configuration [%s]: %w", name, err)
	}

	active := s.inv.env.ActiveConfigName()
	fmt.Fprintf(s.inv.stdout, "name: %s\n", name)
	fmt.Fprintf(s.inv.stdout, "is_active: %v\n", name == active)

	fmt.Fprintln(s.inv.stdout, "properties:")
	fmt.Fprintln(s.inv.stdout, "  core:")
	fmt.Fprintf(s.inv.stdout, "    account: %s\n", props.Core.Account)
	fmt.Fprintf(s.inv.stdout, "    project: %s\n", props.Core.Project)
	fmt.Fprintln(s.inv.stdout, "  compute:")
	fmt.Fprintf(s.inv.stdout, "    region: %s\n", props.Compute.Region)
	fmt.Fprintf(s.inv.stdout, "    zone: %s\n", props.Compute.Zone)
	fmt.Fprintln(s.inv.stdout, "  dataflow:")
	fmt.Fprintf(s.inv.stdout, "    region: %s\n", props.Dataflow.Region)
	fmt.Fprintln(s.inv.stdout, "  run:")
	fmt.Fprintf(s.inv.stdout, "    region: %s\n", props.Run.Region)
	fmt.Fprintln(s.inv.stdout, "  redis:")
	fmt.Fprintf(s.inv.stdout, "    region: %s\n", props.Redis.Region)
	fmt.Fprintln(s.inv.stdout, "  functions:")
	fmt.Fprintf(s.inv.stdout, "    region: %s\n", props.Functions.Region)

	return nil
}
