package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	baremetalsolution "google.golang.org/api/baremetalsolution/v2"
)

// --- gcloud bms ssh-keys (#1231) ---

func (s *bmsCmds) initBmsSshKeys() {
	all := []*cobra.Command{s.bmsSshCreateCmd, s.bmsSshDeleteCmd, s.bmsSshListCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagBmsSshLocation, "location", "", "Location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagBmsSshFormat, "format", "", "Output format")
	}
	s.bmsSshCreateCmd.Flags().StringVar(&s.flagBmsSshConfigFile, "config-file", "", "YAML/JSON file with the SSHKey body (required)")
	_ = s.bmsSshCreateCmd.MarkFlagRequired("config-file")
	s.bmsSshListCmd.Flags().Int64Var(&s.flagBmsSshPageSize, "page-size", 0, "Maximum results per page")

	s.bmsSshKeysCmd.AddCommand(all...)
	s.bmsCmd.AddCommand(s.bmsSshKeysCmd)
}

func (s *bmsCmds) bmsSshName(id string) (string, error) {
	return s.bmsResource(s.flagBmsSshLocation, "sshKeys", id)
}

func (s *bmsCmds) runBmsSshCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.bmsLocationParent(s.flagBmsSshLocation)
	if err != nil {
		return err
	}
	body := &baremetalsolution.SSHKey{}
	if err := loadYAMLOrJSONInto(s.flagBmsSshConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.SshKeys.Create(parent, body).SshKeyId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating ssh key: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created ssh key [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagBmsSshFormat)
}

func (s *bmsCmds) runBmsSshDelete(cmd *cobra.Command, args []string) error {
	name, err := s.bmsSshName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.SshKeys.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting ssh key: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted ssh key [%s].\n", args[0])
	return nil
}

func (s *bmsCmds) runBmsSshList(cmd *cobra.Command, args []string) error {
	parent, err := s.bmsLocationParent(s.flagBmsSshLocation)
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BareMetalSolution(ctx)
	if err != nil {
		return err
	}
	var all []*baremetalsolution.SSHKey
	pageToken := ""
	for {
		call := svc.Projects.Locations.SshKeys.List(parent).Context(ctx)
		if s.flagBmsSshPageSize > 0 {
			call = call.PageSize(s.flagBmsSshPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagBmsSshFormat)
}
