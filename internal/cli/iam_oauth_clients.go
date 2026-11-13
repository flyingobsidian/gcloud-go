package cli

import (
	"fmt"
	"path"
	"strings"

	"github.com/spf13/cobra"
	iam "google.golang.org/api/iam/v1"
)

// --- gcloud iam oauth-clients (#1009) ---

func (s *iamCmds) initIamOauthClients() {
	ocAll := []*cobra.Command{s.iamOCCreateCmd, s.iamOCDeleteCmd, s.iamOCDescribeCmd, s.iamOCListCmd,
		s.iamOCUpdateCmd, s.iamOCUndeleteCmd}
	for _, c := range ocAll {
		c.Flags().StringVar(&s.flagIAMOCLocation, "location", "global", "Location containing the OAuth client")
	}
	for _, c := range []*cobra.Command{s.iamOCCreateCmd, s.iamOCUpdateCmd} {
		c.Flags().StringVar(&s.flagIAMOCConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the OauthClient body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.iamOCUpdateCmd.Flags().StringVar(&s.flagIAMOCUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.iamOCDescribeCmd, s.iamOCListCmd} {
		c.Flags().StringVar(&s.flagIAMOCFormat, "format", "", "Output format")
	}
	s.iamOauthClientsCmd.AddCommand(ocAll...)

	// credentials (nested)
	credAll := []*cobra.Command{s.iamOCCredCreateCmd, s.iamOCCredDeleteCmd, s.iamOCCredDescribeCmd,
		s.iamOCCredListCmd, s.iamOCCredUpdateCmd}
	for _, c := range credAll {
		c.Flags().StringVar(&s.flagIAMOCLocation, "location", "global", "Location containing the OAuth client")
		c.Flags().StringVar(&s.flagIAMOCClient, "oauth-client", "", "OAuth client containing the credential (required)")
		_ = c.MarkFlagRequired("oauth-client")
	}
	for _, c := range []*cobra.Command{s.iamOCCredCreateCmd, s.iamOCCredUpdateCmd} {
		c.Flags().StringVar(&s.flagIAMOCConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the OauthClientCredential body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.iamOCCredUpdateCmd.Flags().StringVar(&s.flagIAMOCUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{s.iamOCCredDescribeCmd, s.iamOCCredListCmd} {
		c.Flags().StringVar(&s.flagIAMOCFormat, "format", "", "Output format")
	}
	s.iamOauthClientCredsCmd.AddCommand(credAll...)
	s.iamOauthClientsCmd.AddCommand(s.iamOauthClientCredsCmd)

	s.iamCmd.AddCommand(s.iamOauthClientsCmd)
}

func (s *iamCmds) iamOCParent(project string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagIAMOCLocation)
}

func (s *iamCmds) iamOCName(project, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/oauthClients/%s", s.iamOCParent(project), id)
}

func (s *iamCmds) iamOCCredParent(project string) string {
	return s.iamOCName(project, s.flagIAMOCClient)
}

func (s *iamCmds) iamOCCredName(project, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/credentials/%s", s.iamOCCredParent(project), id)
}

func (s *iamCmds) runIAMOCCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	c := &iam.OauthClient{}
	if err := loadYAMLOrJSONInto(s.flagIAMOCConfigFile, c); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OauthClients.Create(s.iamOCParent(project), c).
		OauthClientId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating oauth client: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *iamCmds) runIAMOCDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OauthClients.Delete(s.iamOCName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting oauth client: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *iamCmds) runIAMOCDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OauthClients.Get(s.iamOCName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing oauth client: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIAMOCFormat)
}

func (s *iamCmds) runIAMOCList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	var all []*iam.OauthClient
	pageToken := ""
	for {
		call := svc.Projects.Locations.OauthClients.List(s.iamOCParent(project)).Context(ctx)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing oauth clients: %w", err)
		}
		all = append(all, resp.OauthClients...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if s.flagIAMOCFormat != "" {
		return emitFormattedTo(s.inv.stdout, all, s.flagIAMOCFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(all), "%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, c := range all {
		fmt.Fprintf(s.inv.stdout, "%-40s %s\n", path.Base(c.Name), c.DisplayName)
	}
	return nil
}

func (s *iamCmds) runIAMOCUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	c := &iam.OauthClient{}
	if err := loadYAMLOrJSONInto(s.flagIAMOCConfigFile, c); err != nil {
		return err
	}
	mask := s.flagIAMOCUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(c))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OauthClients.Patch(s.iamOCName(project, args[0]), c).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating oauth client: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *iamCmds) runIAMOCUndelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OauthClients.Undelete(s.iamOCName(project, args[0]),
		&iam.UndeleteOauthClientRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting oauth client: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

// --- credentials impl ---

func (s *iamCmds) runIAMOCCredCreate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	c := &iam.OauthClientCredential{}
	if err := loadYAMLOrJSONInto(s.flagIAMOCConfigFile, c); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OauthClients.Credentials.Create(s.iamOCCredParent(project), c).
		OauthClientCredentialId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating oauth client credential: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}

func (s *iamCmds) runIAMOCCredDelete(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.OauthClients.Credentials.Delete(s.iamOCCredName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting oauth client credential: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted oauth client credential [%s].\n", args[0])
	return nil
}

func (s *iamCmds) runIAMOCCredDescribe(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OauthClients.Credentials.Get(s.iamOCCredName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing oauth client credential: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagIAMOCFormat)
}

func (s *iamCmds) runIAMOCCredList(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.OauthClients.Credentials.List(s.iamOCCredParent(project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing oauth client credentials: %w", err)
	}
	if s.flagIAMOCFormat != "" {
		return emitFormattedTo(s.inv.stdout, resp.OauthClientCredentials, s.flagIAMOCFormat)
	}
	printTableHeadingTo(s.inv.stdout, len(resp.OauthClientCredentials), "%-40s\n", "NAME")
	for _, c := range resp.OauthClientCredentials {
		fmt.Fprintln(s.inv.stdout, path.Base(c.Name))
	}
	return nil
}

func (s *iamCmds) runIAMOCCredUpdate(cmd *cobra.Command, args []string) error {
	project, err := s.inv.project()
	if err != nil {
		return err
	}
	c := &iam.OauthClientCredential{}
	if err := loadYAMLOrJSONInto(s.flagIAMOCConfigFile, c); err != nil {
		return err
	}
	mask := s.flagIAMOCUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(c))
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OauthClients.Credentials.Patch(s.iamOCCredName(project, args[0]), c).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating oauth client credential: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, "")
}
