package cmd

import (
	"context"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	baremetalsolution "google.golang.org/api/baremetalsolution/v2"
)

// --- gcloud bms ssh-keys (#1231) ---

var bmsSshKeysCmd = &cobra.Command{Use: "ssh-keys", Short: "Manage BMS SSH keys"}

var (
	flagBmsSshLocation   string
	flagBmsSshFormat     string
	flagBmsSshConfigFile string
	flagBmsSshPageSize   int64
)

var (
	bmsSshCreateCmd = &cobra.Command{
		Use: "create SSH_KEY", Short: "Create an SSH key",
		Args: cobra.ExactArgs(1), RunE: runBmsSshCreate,
	}
	bmsSshDeleteCmd = &cobra.Command{
		Use: "delete SSH_KEY", Short: "Delete an SSH key",
		Args: cobra.ExactArgs(1), RunE: runBmsSshDelete,
	}
	bmsSshListCmd = &cobra.Command{
		Use: "list", Short: "List SSH keys in a location",
		Args: cobra.NoArgs, RunE: runBmsSshList,
	}
)

func init() {
	all := []*cobra.Command{bmsSshCreateCmd, bmsSshDeleteCmd, bmsSshListCmd}
	for _, c := range all {
		c.Flags().StringVar(&flagBmsSshLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&flagBmsSshFormat, "format", "", "Output format")
	}
	bmsSshCreateCmd.Flags().StringVar(&flagBmsSshConfigFile, "config-file", "", "YAML/JSON file with the SSHKey body (required)")
	_ = bmsSshCreateCmd.MarkFlagRequired("config-file")
	bmsSshListCmd.Flags().Int64Var(&flagBmsSshPageSize, "page-size", 0, "Maximum results per page")

	bmsSshKeysCmd.AddCommand(all...)
	bmsCmd.AddCommand(bmsSshKeysCmd)
}

func bmsSshName(id string) (string, error) {
	return bmsResource(flagBmsSshLocation, "sshKeys", id)
}

func runBmsSshCreate(cmd *cobra.Command, args []string) error {
	parent, err := bmsLocationParent(flagBmsSshLocation)
	if err != nil {
		return err
	}
	body := &baremetalsolution.SSHKey{}
	if err := loadYAMLOrJSONInto(flagBmsSshConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.SshKeys.Create(parent, body).SshKeyId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating ssh key: %w", err)
	}
	fmt.Printf("Created ssh key [%s].\n", args[0])
	return emitFormatted(got, flagBmsSshFormat)
}

func runBmsSshDelete(cmd *cobra.Command, args []string) error {
	name, err := bmsSshName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.SshKeys.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting ssh key: %w", err)
	}
	fmt.Printf("Deleted ssh key [%s].\n", args[0])
	return nil
}

func runBmsSshList(cmd *cobra.Command, args []string) error {
	parent, err := bmsLocationParent(flagBmsSshLocation)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BareMetalSolutionService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*baremetalsolution.SSHKey
	pageToken := ""
	for {
		call := svc.Projects.Locations.SshKeys.List(parent).Context(ctx)
		if flagBmsSshPageSize > 0 {
			call = call.PageSize(flagBmsSshPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing ssh keys: %w", err)
		}
		all = append(all, resp.SshKeys...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagBmsSshFormat)
}
