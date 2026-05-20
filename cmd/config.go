package cmd

import (
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gcloud CLI configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set PROPERTY VALUE",
	Short: "Set a configuration property",
	Long: `Set a configuration property. Property can be specified as SECTION/PROPERTY or just PROPERTY for core section.

Examples:
  gcloud config set project my-project
  gcloud config set compute/zone us-central1-a
  gcloud config set compute/region europe-west1
  gcloud config set core/account user@example.com`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

var configGetValueCmd = &cobra.Command{
	Use:   "get-value PROPERTY",
	Short: "Print the value of a configuration property",
	Long: `Print the value of a configuration property. Property can be specified as SECTION/PROPERTY or just PROPERTY for core section.

Examples:
  gcloud config get-value project
  gcloud config get-value compute/zone`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGetValue,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all properties in the active configuration",
	Args:  cobra.NoArgs,
	RunE:  runConfigList,
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetValueCmd)
	configCmd.AddCommand(configListCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	property := args[0]
	value := args[1]

	section, key := parseProperty(property)

	props, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	switch section {
	case "core":
		switch key {
		case "account":
			props.Core.Account = value
		case "project":
			props.Core.Project = value
		default:
			return fmt.Errorf("unrecognized property: core/%s", key)
		}
	case "compute":
		switch key {
		case "zone":
			props.Compute.Zone = value
		case "region":
			props.Compute.Region = value
		default:
			return fmt.Errorf("unrecognized property: compute/%s", key)
		}
	default:
		return fmt.Errorf("unrecognized section: %s", section)
	}

	if err := props.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Updated property [%s/%s].\n", section, key)
	return nil
}

func runConfigGetValue(cmd *cobra.Command, args []string) error {
	section, key := parseProperty(args[0])

	props, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	val, err := getPropertyValue(props, section, key)
	if err != nil {
		return err
	}

	if val == "" {
		fmt.Fprintln(cmd.ErrOrStderr(), "Your active configuration is: [default]")
		fmt.Println("(unset)")
		return nil
	}

	fmt.Fprintln(cmd.ErrOrStderr(), "Your active configuration is: [default]")
	fmt.Println(val)
	return nil
}

func runConfigList(cmd *cobra.Command, args []string) error {
	props, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	fmt.Println("[core]")
	if props.Core.Account != "" {
		fmt.Printf("account = %s\n", props.Core.Account)
	}
	if props.Core.Project != "" {
		fmt.Printf("project = %s\n", props.Core.Project)
	}

	if props.Compute.Zone != "" || props.Compute.Region != "" {
		fmt.Println("[compute]")
		if props.Compute.Region != "" {
			fmt.Printf("region = %s\n", props.Compute.Region)
		}
		if props.Compute.Zone != "" {
			fmt.Printf("zone = %s\n", props.Compute.Zone)
		}
	}

	fmt.Fprintln(cmd.ErrOrStderr(), "Your active configuration is: [default]")
	return nil
}

func getPropertyValue(props *config.Properties, section, key string) (string, error) {
	switch section {
	case "core":
		switch key {
		case "account":
			return props.Core.Account, nil
		case "project":
			return props.Core.Project, nil
		default:
			return "", fmt.Errorf("unrecognized property: core/%s", key)
		}
	case "compute":
		switch key {
		case "zone":
			return props.Compute.Zone, nil
		case "region":
			return props.Compute.Region, nil
		default:
			return "", fmt.Errorf("unrecognized property: compute/%s", key)
		}
	default:
		return "", fmt.Errorf("unrecognized section: %s", section)
	}
}

// parseProperty splits "section/key" or defaults to "core/key".
func parseProperty(prop string) (section, key string) {
	if i := strings.IndexByte(prop, '/'); i >= 0 {
		return prop[:i], prop[i+1:]
	}
	return "core", prop
}
