package cmd

import (
	"fmt"
	"maps"
	"slices"
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
  gcloud config set core/account user@example.com
  gcloud config set workflows/location europe-west2`,
	Args: cobra.ExactArgs(2),
	RunE: runConfigSet,
}

var configGetCmd = &cobra.Command{
	Use:   "get PROPERTY",
	Short: "Print the value of a configuration property",
	Long: `Print the value of a configuration property. Property can be specified as SECTION/PROPERTY or just PROPERTY for core section.

Examples:
  gcloud config get project
  gcloud config get compute/zone`,
	Args: cobra.ExactArgs(1),
	RunE: runConfigGet,
}

// configGetValueCmd remains registered as a deprecated alias for
// backwards-compat, matching gcloud-python. See #536.
var configGetValueCmd = &cobra.Command{
	Use:        "get-value PROPERTY",
	Short:      "(DEPRECATED) alias for gcloud config get",
	Deprecated: "use `gcloud config get` instead.",
	Hidden:     false,
	Args:       cobra.ExactArgs(1),
	RunE:       runConfigGet,
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
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configGetValueCmd)
	configCmd.AddCommand(configUnsetCmd)
	configCmd.AddCommand(configListCmd)
	rootCmd.AddCommand(configCmd)
}

func runConfigSet(cmd *cobra.Command, args []string) error {
	section, key := parseProperty(args[0])
	if err := config.SetProperty(section, key, args[1]); err != nil {
		return fmt.Errorf("setting property: %w", err)
	}
	fmt.Printf("Updated property [%s/%s].\n", section, key)
	return nil
}

func runConfigUnset(cmd *cobra.Command, args []string) error {
	section, key := parseProperty(args[0])
	if err := config.UnsetProperty(section, key); err != nil {
		return fmt.Errorf("unsetting property: %w", err)
	}
	fmt.Printf("Unset property [%s/%s].\n", section, key)
	return nil
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	section, key := parseProperty(args[0])
	if err := config.ValidatePropertyName(section, key); err != nil {
		return err
	}

	val, err := config.Property(section, key)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	fmt.Fprintln(cmd.ErrOrStderr(), "Your active configuration is: [default]")
	if val == "" {
		fmt.Println("(unset)")
		return nil
	}
	fmt.Println(val)
	return nil
}

// configListKnown are the properties `config list --all` shows even when
// unset.
var configListKnown = map[string][]string{
	"core":      {"account", "project"},
	"compute":   {"region", "zone"},
	"billing":   {"quota_project"},
	"dataflow":  {"region"},
	"run":       {"region"},
	"redis":     {"region"},
	"functions": {"region"},
}

func runConfigList(cmd *cobra.Command, args []string) error {
	props, err := config.FileProperties()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if flagConfigListAll {
		for section, keys := range configListKnown {
			if props[section] == nil {
				props[section] = map[string]string{}
			}
			for _, key := range keys {
				if _, ok := props[section][key]; !ok {
					props[section][key] = ""
				}
			}
		}
	}

	// Sections and keys are sorted, as gcloud-python's config printer does.
	for _, section := range slices.Sorted(maps.Keys(props)) {
		if flagConfigListFilter != "" && !strings.Contains(section, flagConfigListFilter) {
			continue
		}
		keys := props[section]
		if len(keys) == 0 {
			continue
		}
		fmt.Printf("[%s]\n", section)
		for _, key := range slices.Sorted(maps.Keys(keys)) {
			display := keys[key]
			if display == "" {
				display = "(unset)"
			}
			fmt.Printf("%s = %s\n", key, display)
		}
	}

	fmt.Fprintln(cmd.ErrOrStderr(), "Your active configuration is: [default]")
	return nil
}

// parseProperty splits "section/key" or defaults to "core/key".
func parseProperty(prop string) (section, key string) {
	if i := strings.IndexByte(prop, '/'); i >= 0 {
		return prop[:i], prop[i+1:]
	}
	return "core", prop
}
