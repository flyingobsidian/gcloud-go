package cmd

import (
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-golang-cli/internal/config"
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

func init() {
	configCmd.AddCommand(configSetCmd)
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

// parseProperty splits "section/key" or defaults to "core/key".
func parseProperty(prop string) (section, key string) {
	if i := strings.IndexByte(prop, '/'); i >= 0 {
		return prop[:i], prop[i+1:]
	}
	return "core", prop
}
