package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/config"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Display information about the current gcloud-go environment",
	Args:  cobra.NoArgs,
	RunE:  runInfo,
}

var flagInfoFormat string

func init() {
	infoCmd.Flags().StringVar(&flagInfoFormat, "format", "", "Output format (e.g. json, value(FIELD))")
	rootCmd.AddCommand(infoCmd)
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

func runInfo(cmd *cobra.Command, args []string) error {
	configDir := os.Getenv("CLOUDSDK_CONFIG")
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config", "gcloud")
	}

	props, _ := config.Load()

	info := infoResult{
		Basic: infoBasic{
			Version:         "gcloud-go 0.1.0",
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

	// Handle --format
	if flagInfoFormat == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(info)
	}

	if strings.HasPrefix(flagInfoFormat, "value(") && strings.HasSuffix(flagInfoFormat, ")") {
		field := flagInfoFormat[6 : len(flagInfoFormat)-1]
		val := getInfoField(info, field)
		fmt.Println(val)
		return nil
	}

	// Default text output
	fmt.Printf("gcloud-go [%s]\n\n", info.Basic.Version)
	fmt.Printf("Platform: [%s, %s]\n", info.Basic.OperatingSystem, info.Basic.Architecture)
	fmt.Printf("Go Version: [%s]\n", info.Basic.GoVersion)
	fmt.Printf("Config Directory: [%s]\n", info.Installation.ConfigDir)
	fmt.Printf("\nAccount: [%s]\n", info.Config.Account)
	fmt.Printf("Project: [%s]\n", info.Config.Project)
	fmt.Printf("Zone: [%s]\n", info.Config.Zone)
	fmt.Printf("Region: [%s]\n", info.Config.Region)
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
