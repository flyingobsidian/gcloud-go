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

func init() {
	configConfigurationsCmd.AddCommand(configConfigurationsCreateCmd)
	configConfigurationsCmd.AddCommand(configConfigurationsListCmd)
	configConfigurationsCmd.AddCommand(configConfigurationsActivateCmd)
	configConfigurationsCmd.AddCommand(configConfigurationsDeleteCmd)
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
