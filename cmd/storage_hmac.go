package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	storage "google.golang.org/api/storage/v1"
)

// --- gcloud storage hmac (#1238) ---

func (s *storageCmds) initStorageHmac() {
	all := []*cobra.Command{
		s.storageHmacCreateCmd, s.storageHmacDeleteCmd, s.storageHmacDescribeCmd,
		s.storageHmacListCmd, s.storageHmacUpdateCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagStHmacFormat, "format", "", "Output format")
	}
	s.storageHmacCreateCmd.Flags().StringVar(&s.flagStHmacServiceAccount, "service-account", "", "Service account email to create the key for (required)")
	_ = s.storageHmacCreateCmd.MarkFlagRequired("service-account")
	s.storageHmacListCmd.Flags().StringVar(&s.flagStHmacServiceAccount, "service-account", "", "Restrict list to a service account")
	s.storageHmacUpdateCmd.Flags().StringVar(&s.flagStHmacState, "state", "", "New state: ACTIVE or INACTIVE (required)")
	_ = s.storageHmacUpdateCmd.MarkFlagRequired("state")

	s.storageHmacCmd.AddCommand(all...)
	s.storageCmd.AddCommand(s.storageHmacCmd)
}

func (s *storageCmds) runStHmacCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.HmacKeys.Create(project, s.flagStHmacServiceAccount).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating HMAC key: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagStHmacFormat)
}

func (s *storageCmds) runStHmacDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	if err := svc.Projects.HmacKeys.Delete(project, args[0]).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting HMAC key: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted HMAC key [%s].\n", args[0])
	return nil
}

func (s *storageCmds) runStHmacDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.HmacKeys.Get(project, args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing HMAC key: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagStHmacFormat)
}

func (s *storageCmds) runStHmacList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	var all []*storage.HmacKeyMetadata
	pageToken := ""
	for {
		call := svc.Projects.HmacKeys.List(project).Context(ctx)
		if s.flagStHmacServiceAccount != "" {
			call = call.ServiceAccountEmail(s.flagStHmacServiceAccount)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing HMAC keys: %w", err)
		}
		all = append(all, resp.Items...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormattedTo(s.inv.stdout, all, s.flagStHmacFormat)
}

func (s *storageCmds) runStHmacUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().Storage(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.HmacKeys.Update(project, args[0], &storage.HmacKeyMetadata{State: s.flagStHmacState}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating HMAC key: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated HMAC key [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagStHmacFormat)
}
