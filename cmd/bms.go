package cmd

import "github.com/spf13/cobra"

// --- gcloud bms (#310) ---

var bmsCmd = &cobra.Command{Use: "bms", Short: "Manage Bare Metal Solution (stubbed)"}

func init() {
	crud := []string{"describe", "list", "update"}
	registerStubGroup(bmsCmd, "instances", "Manage bare metal instances", append(crud, "reset", "start", "stop", "enable-interactive-serial-console", "disable-interactive-serial-console", "rename")...)
	registerStubGroup(bmsCmd, "networks", "Manage networks", crud...)
	registerStubGroup(bmsCmd, "nfs-shares", "Manage NFS shares", append(crud, "rename")...)
	registerStubGroup(bmsCmd, "operations", "Manage operations", "describe", "list")
	registerStubGroup(bmsCmd, "os-images", "Manage OS images", "list")
	registerStubGroup(bmsCmd, "ssh-keys", "Manage SSH keys", "create", "delete", "list")
	registerStubGroup(bmsCmd, "volumes", "Manage volumes", append(crud, "rename", "resize", "restore")...)
	rootCmd.AddCommand(bmsCmd)
}
