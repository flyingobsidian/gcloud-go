package cmd

import (
	"github.com/spf13/cobra"
)

var computeCmd = &cobra.Command{
	Use:   "compute",
	Short: "Compute Engine commands",
}

var instancesCmd = &cobra.Command{
	Use:   "instances",
	Short: "Manage Compute Engine instances",
}

var instanceGroupsCmd = &cobra.Command{
	Use:   "instance-groups",
	Short: "Manage instance groups",
}

var unmanagedCmd = &cobra.Command{
	Use:   "unmanaged",
	Short: "Manage unmanaged instance groups",
}

func init() {
	instanceGroupsCmd.AddCommand(unmanagedCmd)
	computeCmd.AddCommand(instancesCmd)
	computeCmd.AddCommand(instanceGroupsCmd)
	rootCmd.AddCommand(computeCmd)
}
