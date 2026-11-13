package cmd

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

func (s *infoCmds) initInfo() {
	s.infoCmd.Flags().StringVar(&s.flagInfoFormat, "format", "", "Output format (e.g. json, value(FIELD))")
	s.infoCmd.Flags().BoolVar(&s.flagInfoAnonymize, "anonymize", false, "Anonymize sensitive information in output")

}

type infoBasic struct {
	Version         string `json:"version"`
	OperatingSystem string `json:"operating_system"`
	Architecture    string `json:"architecture"`
	GoVersion       string `json:"go_version"`
}

type infoConfig struct {
	Account string `json:"account"`
	Project string `json:"project"`
	Zone    string `json:"zone"`
	Region  string `json:"region"`
}

type infoInstallation struct {
	ConfigDir string `json:"config_dir"`
}

type infoResult struct {
	Basic        infoBasic        `json:"basic"`
	Installation infoInstallation `json:"installation"`
	Config       infoConfig       `json:"config"`
}

func (s *infoCmds) runInfo(cmd *cobra.Command, args []string) error {
	configDir := s.inv.env.Getenv("CLOUDSDK_CONFIG")
	if configDir == "" {
		home, _ := s.inv.env.HomeDir()
		configDir = filepath.Join(home, ".config", "gcloud")
	}

	props, _ := s.inv.properties()

	info := infoResult{
		Basic: infoBasic{
			Version:         "gcloud-go " + Version,
			OperatingSystem: runtime.GOOS,
			Architecture:    runtime.GOARCH,
			GoVersion:       runtime.Version(),
		},
		Installation: infoInstallation{
			ConfigDir: configDir,
		},
	}
	if props != nil {
		info.Config = infoConfig{
			Account: props.Core.Account,
			Project: props.Core.Project,
			Zone:    props.Compute.Zone,
			Region:  props.Compute.Region,
		}
	}

	if s.flagInfoAnonymize {
		info.Config.Account = "REDACTED"
		info.Config.Project = "REDACTED"
		info.Installation.ConfigDir = "REDACTED"
	}

	// Handle --format
	if s.flagInfoFormat == "json" {
		return s.inv.printJSON(info)
	}

	if strings.HasPrefix(s.flagInfoFormat, "value(") && strings.HasSuffix(s.flagInfoFormat, ")") {
		field := s.flagInfoFormat[6 : len(s.flagInfoFormat)-1]
		val := getInfoField(info, field)
		fmt.Fprintln(s.inv.stdout, val)
		return nil
	}

	// Default text output
	fmt.Fprintf(s.inv.stdout, "gcloud-go [%s]\n\n", info.Basic.Version)
	fmt.Fprintf(s.inv.stdout, "Platform: [%s, %s]\n", info.Basic.OperatingSystem, info.Basic.Architecture)
	fmt.Fprintf(s.inv.stdout, "Go Version: [%s]\n", info.Basic.GoVersion)
	fmt.Fprintf(s.inv.stdout, "Config Directory: [%s]\n", info.Installation.ConfigDir)
	fmt.Fprintf(s.inv.stdout, "\nAccount: [%s]\n", info.Config.Account)
	fmt.Fprintf(s.inv.stdout, "Project: [%s]\n", info.Config.Project)
	fmt.Fprintf(s.inv.stdout, "Zone: [%s]\n", info.Config.Zone)
	fmt.Fprintf(s.inv.stdout, "Region: [%s]\n", info.Config.Region)
	return nil
}

func getInfoField(info infoResult, field string) string {
	switch field {
	case "basic.version":
		return info.Basic.Version
	case "basic.operating_system":
		return info.Basic.OperatingSystem
	case "basic.architecture":
		return info.Basic.Architecture
	case "basic.go_version":
		return info.Basic.GoVersion
	case "installation.config_dir":
		return info.Installation.ConfigDir
	case "config.account":
		return info.Config.Account
	case "config.project":
		return info.Config.Project
	case "config.zone":
		return info.Config.Zone
	case "config.region":
		return info.Config.Region
	default:
		return ""
	}
}

// infoCmds holds the commands and flag values of one invocation's "info"
// command group, so invocations don't share state (#2070).
type infoCmds struct {
	inv *invocation

	infoCmd           *cobra.Command
	flagInfoFormat    string
	flagInfoAnonymize bool
}

// newInfoCmd returns the "info" command group for inv.
func newInfoCmd(inv *invocation) *cobra.Command {
	return newInfoCmds(inv).infoCmd
}

// newInfoCmds builds the "info" command group for inv and returns its state.
func newInfoCmds(inv *invocation) *infoCmds {
	s := &infoCmds{inv: inv}
	s.infoCmd = &cobra.Command{
		Use:   "info",
		Short: "Display information about the current gcloud-go environment",
		Args:  cobra.NoArgs,
		RunE:  s.runInfo,
	}
	s.initInfo()
	return s
}
