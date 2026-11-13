package cmd

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/spf13/cobra"
)

func (s *configCmds) initConfig() {
	s.configListCmd.Flags().BoolVar(&s.flagConfigListAll, "all", false, "Show all properties including unset ones")
	s.configListCmd.Flags().StringVar(&s.flagConfigListFilter, "filter", "", "Filter output by section name")
	s.configCmd.AddCommand(s.configSetCmd)
	s.configCmd.AddCommand(s.configGetCmd)
	s.configCmd.AddCommand(s.configGetValueCmd)
	s.configCmd.AddCommand(s.configUnsetCmd)
	s.configCmd.AddCommand(s.configListCmd)

}

func (s *configCmds) runConfigSet(cmd *cobra.Command, args []string) error {
	section, key := parseProperty(args[0])
	if err := s.inv.env.SetProperty(section, key, args[1]); err != nil {
		return fmt.Errorf("setting property: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated property [%s/%s].\n", section, key)
	return nil
}

func (s *configCmds) runConfigUnset(cmd *cobra.Command, args []string) error {
	section, key := parseProperty(args[0])
	if err := s.inv.env.UnsetProperty(section, key); err != nil {
		return fmt.Errorf("unsetting property: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Unset property [%s/%s].\n", section, key)
	return nil
}

func (s *configCmds) runConfigGet(cmd *cobra.Command, args []string) error {
	section, key := parseProperty(args[0])
	if err := config.ValidatePropertyName(section, key); err != nil {
		return err
	}

	val, err := s.inv.env.Property(section, key)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// gcloud-python only reports a non-default active configuration here.
	if name := s.inv.env.ActiveConfigName(); name != "default" {
		fmt.Fprintf(cmd.ErrOrStderr(), "Your active configuration is: [%s]\n", name)
	}
	if val == "" {
		fmt.Fprintln(s.inv.stdout, "(unset)")
		return nil
	}
	fmt.Fprintln(s.inv.stdout, val)
	return nil
}

// configListKnown are the properties `config list --all` shows even when
// unset.
func configListKnown() map[string][]string {
	return map[string][]string{
		"core":      {"account", "project"},
		"compute":   {"region", "zone"},
		"billing":   {"quota_project"},
		"dataflow":  {"region"},
		"run":       {"region"},
		"redis":     {"region"},
		"functions": {"region"},
	}
}

func (s *configCmds) runConfigList(cmd *cobra.Command, args []string) error {
	props, err := s.inv.env.FileProperties()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if s.flagConfigListAll {
		for section, keys := range configListKnown() {
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
		if s.flagConfigListFilter != "" && !strings.Contains(section, s.flagConfigListFilter) {
			continue
		}
		keys := props[section]
		if len(keys) == 0 {
			continue
		}
		fmt.Fprintf(s.inv.stdout, "[%s]\n", section)
		for _, key := range slices.Sorted(maps.Keys(keys)) {
			display := keys[key]
			if display == "" {
				display = "(unset)"
			}
			fmt.Fprintf(s.inv.stdout, "%s = %s\n", key, display)
		}
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "\nYour active configuration is: [%s]\n", s.inv.env.ActiveConfigName())
	return nil
}

// parseProperty splits "section/key" or defaults to "core/key".
func parseProperty(prop string) (section, key string) {
	if i := strings.IndexByte(prop, '/'); i >= 0 {
		return prop[:i], prop[i+1:]
	}
	return "core", prop
}

// configCmds holds the commands and flag values of one invocation's "config"
// command group, so invocations don't share state (#2070).
type configCmds struct {
	inv *invocation

	configCmd    *cobra.Command
	configSetCmd *cobra.Command
	configGetCmd *cobra.Command
	// configGetValueCmd remains registered as a deprecated alias for
	// backwards-compat, matching gcloud-python. See #536.
	configGetValueCmd               *cobra.Command
	configUnsetCmd                  *cobra.Command
	configListCmd                   *cobra.Command
	flagConfigListAll               bool
	flagConfigListFilter            string
	configConfigurationsCmd         *cobra.Command
	configConfigurationsCreateCmd   *cobra.Command
	configConfigurationsListCmd     *cobra.Command
	configConfigurationsActivateCmd *cobra.Command
	configConfigurationsDeleteCmd   *cobra.Command
	configConfigurationsDescribeCmd *cobra.Command
}

// newConfigCmd returns the "config" command group for inv.
func newConfigCmd(inv *invocation) *cobra.Command {
	return newConfigCmds(inv).configCmd
}

// newConfigCmds builds the "config" command group for inv and returns its state.
func newConfigCmds(inv *invocation) *configCmds {
	s := &configCmds{inv: inv}
	s.configCmd = &cobra.Command{
		Use:   "config",
		Short: "Manage gcloud CLI configuration",
	}
	s.configSetCmd = &cobra.Command{
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
		RunE: s.runConfigSet,
	}
	s.configGetCmd = &cobra.Command{
		Use:   "get PROPERTY",
		Short: "Print the value of a configuration property",
		Long: `Print the value of a configuration property. Property can be specified as SECTION/PROPERTY or just PROPERTY for core section.

Examples:
  gcloud config get project
  gcloud config get compute/zone`,
		Args: cobra.ExactArgs(1),
		RunE: s.runConfigGet,
	}
	s.configGetValueCmd = &cobra.Command{
		Use:        "get-value PROPERTY",
		Short:      "(DEPRECATED) alias for gcloud config get",
		Deprecated: "use `gcloud config get` instead.",
		Hidden:     false,
		Args:       cobra.ExactArgs(1),
		RunE:       s.runConfigGet,
	}
	s.configUnsetCmd = &cobra.Command{
		Use:   "unset PROPERTY",
		Short: "Unset a configuration property",
		Long: `Unset a configuration property. Property can be specified as SECTION/PROPERTY or just PROPERTY for core section.

Examples:
  gcloud config unset project
  gcloud config unset compute/zone`,
		Args: cobra.ExactArgs(1),
		RunE: s.runConfigUnset,
	}
	s.configListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all properties in the active configuration",
		Args:  cobra.NoArgs,
		RunE:  s.runConfigList,
	}
	s.configConfigurationsCmd = &cobra.Command{
		Use:   "configurations",
		Short: "Manage named configurations",
	}
	s.configConfigurationsCreateCmd = &cobra.Command{
		Use:   "create NAME",
		Short: "Create a new named configuration",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runConfigConfigurationsCreate,
	}
	s.configConfigurationsListCmd = &cobra.Command{
		Use:   "list",
		Short: "List all available configurations",
		Args:  cobra.NoArgs,
		RunE:  s.runConfigConfigurationsList,
	}
	s.configConfigurationsActivateCmd = &cobra.Command{
		Use:   "activate NAME",
		Short: "Activate a named configuration",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runConfigConfigurationsActivate,
	}
	s.configConfigurationsDeleteCmd = &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a named configuration",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runConfigConfigurationsDelete,
	}
	s.configConfigurationsDescribeCmd = &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe a named configuration",
		Args:  cobra.ExactArgs(1),
		RunE:  s.runConfigConfigurationsDescribe,
	}
	s.initConfig()
	s.initConfigConfigurations()
	return s
}
