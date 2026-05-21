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

var configUnsetCmd = &cobra.Command{
	Use:   "unset PROPERTY",
	Short: "Unset a configuration property",
	Long: `Unset a configuration property. Property can be specified as SECTION/PROPERTY or just PROPERTY for core section.

Examples:
  gcloud config unset project
  gcloud config unset compute/zone`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigUnset,
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all properties in the active configuration",
	Args:  cobra.NoArgs,
	RunE:  runConfigList,
}

var (
	flagConfigListAll    bool
	flagConfigListFilter string
)

func init() {
	configListCmd.Flags().BoolVar(&flagConfigListAll, "all", false, "Show all properties including unset ones")
	configListCmd.Flags().StringVar(&flagConfigListFilter, "filter", "", "Filter output by section name")
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetValueCmd)
	configCmd.AddCommand(configUnsetCmd)
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

	if err := setPropertyValue(props, section, key, value); err != nil {
		return err
	}

	if err := props.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Updated property [%s/%s].\n", section, key)
	return nil
}

func runConfigUnset(cmd *cobra.Command, args []string) error {
	property := args[0]
	section, key := parseProperty(property)

	props, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := setPropertyValue(props, section, key, ""); err != nil {
		return err
	}

	if err := props.Save(); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}

	fmt.Printf("Unset property [%s/%s].\n", section, key)
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

	showAll := flagConfigListAll
	filter := flagConfigListFilter

	type section struct {
		name  string
		props []struct{ key, val string }
	}

	sections := []section{
		{"core", []struct{ key, val string }{
			{"account", props.Core.Account},
			{"project", props.Core.Project},
		}},
		{"compute", []struct{ key, val string }{
			{"region", props.Compute.Region},
			{"zone", props.Compute.Zone},
		}},
		{"dataflow", []struct{ key, val string }{{"region", props.Dataflow.Region}}},
		{"run", []struct{ key, val string }{{"region", props.Run.Region}}},
		{"redis", []struct{ key, val string }{{"region", props.Redis.Region}}},
		{"functions", []struct{ key, val string }{{"region", props.Functions.Region}}},
	}

	for _, s := range sections {
		if filter != "" && !strings.Contains(s.name, filter) {
			continue
		}
		hasValues := false
		for _, p := range s.props {
			if p.val != "" {
				hasValues = true
				break
			}
		}
		if !hasValues && !showAll {
			continue
		}
		fmt.Printf("[%s]\n", s.name)
		for _, p := range s.props {
			if p.val != "" || showAll {
				display := p.val
				if display == "" {
					display = "(unset)"
				}
				fmt.Printf("%s = %s\n", p.key, display)
			}
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
	case "dataflow":
		if key == "region" {
			return props.Dataflow.Region, nil
		}
		return "", fmt.Errorf("unrecognized property: dataflow/%s", key)
	case "run":
		if key == "region" {
			return props.Run.Region, nil
		}
		return "", fmt.Errorf("unrecognized property: run/%s", key)
	case "redis":
		if key == "region" {
			return props.Redis.Region, nil
		}
		return "", fmt.Errorf("unrecognized property: redis/%s", key)
	case "functions":
		if key == "region" {
			return props.Functions.Region, nil
		}
		return "", fmt.Errorf("unrecognized property: functions/%s", key)
	default:
		return "", fmt.Errorf("unrecognized section: %s", section)
	}
}

func setPropertyValue(props *config.Properties, section, key, value string) error {
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
	case "dataflow":
		if key == "region" {
			props.Dataflow.Region = value
			return nil
		}
		return fmt.Errorf("unrecognized property: dataflow/%s", key)
	case "run":
		if key == "region" {
			props.Run.Region = value
			return nil
		}
		return fmt.Errorf("unrecognized property: run/%s", key)
	case "redis":
		if key == "region" {
			props.Redis.Region = value
			return nil
		}
		return fmt.Errorf("unrecognized property: redis/%s", key)
	case "functions":
		if key == "region" {
			props.Functions.Region = value
			return nil
		}
		return fmt.Errorf("unrecognized property: functions/%s", key)
	default:
		return fmt.Errorf("unrecognized section: %s", section)
	}
	return nil
}

// parseProperty splits "section/key" or defaults to "core/key".
func parseProperty(prop string) (section, key string) {
	if i := strings.IndexByte(prop, '/'); i >= 0 {
		return prop[:i], prop[i+1:]
	}
	return "core", prop
}
