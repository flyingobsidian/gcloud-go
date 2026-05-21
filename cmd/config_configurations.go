package cmd

import (
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/spf13/cobra"
)

var configConfigurationsCmd = &cobra.Command{
	Use:   "configurations",
	Short: "Manage named configurations",
}

var configConfigurationsCreateCmd = &cobra.Command{
	Use:   "create NAME",
	Short: "Create a new named configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigConfigurationsCreate,
}

var configConfigurationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available configurations",
	Args:  cobra.NoArgs,
	RunE:  runConfigConfigurationsList,
}

var configConfigurationsActivateCmd = &cobra.Command{
	Use:   "activate NAME",
	Short: "Activate a named configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigConfigurationsActivate,
}

var configConfigurationsDeleteCmd = &cobra.Command{
	Use:   "delete NAME",
	Short: "Delete a named configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigConfigurationsDelete,
}

var configConfigurationsDescribeCmd = &cobra.Command{
	Use:   "describe NAME",
	Short: "Describe a named configuration",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigConfigurationsDescribe,
}

func init() {
	configConfigurationsCmd.AddCommand(configConfigurationsCreateCmd)
	configConfigurationsCmd.AddCommand(configConfigurationsListCmd)
	configConfigurationsCmd.AddCommand(configConfigurationsActivateCmd)
	configConfigurationsCmd.AddCommand(configConfigurationsDeleteCmd)
	configConfigurationsCmd.AddCommand(configConfigurationsDescribeCmd)
	configCmd.AddCommand(configConfigurationsCmd)
}

func runConfigConfigurationsCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	if err := config.CreateConfiguration(name); err != nil {
		return err
	}
	fmt.Printf("Created configuration [%s].\n", name)
	return nil
}

func runConfigConfigurationsList(cmd *cobra.Command, args []string) error {
	names, err := config.ListConfigurations()
	if err != nil {
		return fmt.Errorf("listing configurations: %w", err)
	}

	active := config.ActiveConfigName()

	fmt.Println("NAME      IS_ACTIVE")
	for _, name := range names {
		marker := "False"
		if name == active {
			marker = "True"
		}
		fmt.Printf("%-10s%s\n", name, marker)
	}
	return nil
}

func runConfigConfigurationsActivate(cmd *cobra.Command, args []string) error {
	name := args[0]
	if err := config.ActivateConfiguration(name); err != nil {
		return err
	}
	fmt.Printf("Activated configuration [%s].\n", name)
	return nil
}

func runConfigConfigurationsDelete(cmd *cobra.Command, args []string) error {
	name := args[0]
	if err := config.DeleteConfiguration(name); err != nil {
		return err
	}
	fmt.Printf("Deleted configuration [%s].\n", name)
	return nil
}

func runConfigConfigurationsDescribe(cmd *cobra.Command, args []string) error {
	name := args[0]
	props, err := config.LoadNamed(name)
	if err != nil {
		return fmt.Errorf("loading configuration [%s]: %w", name, err)
	}

	active := config.ActiveConfigName()
	fmt.Printf("name: %s\n", name)
	fmt.Printf("is_active: %v\n", name == active)

	fmt.Println("properties:")
	fmt.Println("  core:")
	fmt.Printf("    account: %s\n", props.Core.Account)
	fmt.Printf("    project: %s\n", props.Core.Project)
	fmt.Println("  compute:")
	fmt.Printf("    region: %s\n", props.Compute.Region)
	fmt.Printf("    zone: %s\n", props.Compute.Zone)
	fmt.Println("  dataflow:")
	fmt.Printf("    region: %s\n", props.Dataflow.Region)
	fmt.Println("  run:")
	fmt.Printf("    region: %s\n", props.Run.Region)
	fmt.Println("  redis:")
	fmt.Printf("    region: %s\n", props.Redis.Region)
	fmt.Println("  functions:")
	fmt.Printf("    region: %s\n", props.Functions.Region)

	return nil
}
