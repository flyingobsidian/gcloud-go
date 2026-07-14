package cmd

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	iam "google.golang.org/api/iam/v1"
)

// --- gcloud iam oauth-clients (#1009) ---

var iamOauthClientsCmd = &cobra.Command{Use: "oauth-clients", Short: "Manage IAM OAuth clients"}

var iamOauthClientCredsCmd = &cobra.Command{Use: "credentials", Short: "Manage IAM OAuth client credentials"}

var (
	iamOCCreateCmd = &cobra.Command{
		Use: "create OAUTH_CLIENT_ID", Short: "Create an OAuth client from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runIAMOCCreate,
	}
	iamOCDeleteCmd = &cobra.Command{
		Use: "delete OAUTH_CLIENT_ID", Short: "Delete (soft) an OAuth client",
		Args: cobra.ExactArgs(1), RunE: runIAMOCDelete,
	}
	iamOCDescribeCmd = &cobra.Command{
		Use: "describe OAUTH_CLIENT_ID", Short: "Describe an OAuth client",
		Args: cobra.ExactArgs(1), RunE: runIAMOCDescribe,
	}
	iamOCListCmd = &cobra.Command{
		Use: "list", Short: "List OAuth clients in a location",
		Args: cobra.NoArgs, RunE: runIAMOCList,
	}
	iamOCUpdateCmd = &cobra.Command{
		Use: "update OAUTH_CLIENT_ID", Short: "Update an OAuth client from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runIAMOCUpdate,
	}
	iamOCUndeleteCmd = &cobra.Command{
		Use: "undelete OAUTH_CLIENT_ID", Short: "Undelete an OAuth client",
		Args: cobra.ExactArgs(1), RunE: runIAMOCUndelete,
	}

	iamOCCredCreateCmd = &cobra.Command{
		Use: "create CREDENTIAL_ID", Short: "Create an OAuth client credential from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runIAMOCCredCreate,
	}
	iamOCCredDeleteCmd = &cobra.Command{
		Use: "delete CREDENTIAL_ID", Short: "Delete an OAuth client credential",
		Args: cobra.ExactArgs(1), RunE: runIAMOCCredDelete,
	}
	iamOCCredDescribeCmd = &cobra.Command{
		Use: "describe CREDENTIAL_ID", Short: "Describe an OAuth client credential",
		Args: cobra.ExactArgs(1), RunE: runIAMOCCredDescribe,
	}
	iamOCCredListCmd = &cobra.Command{
		Use: "list", Short: "List OAuth client credentials",
		Args: cobra.NoArgs, RunE: runIAMOCCredList,
	}
	iamOCCredUpdateCmd = &cobra.Command{
		Use: "update CREDENTIAL_ID", Short: "Update an OAuth client credential from a --config-file",
		Args: cobra.ExactArgs(1), RunE: runIAMOCCredUpdate,
	}
)

var (
	flagIAMOCLocation   string
	flagIAMOCConfigFile string
	flagIAMOCUpdateMask string
	flagIAMOCFormat     string
	flagIAMOCClient     string
)

func init() {
	ocAll := []*cobra.Command{iamOCCreateCmd, iamOCDeleteCmd, iamOCDescribeCmd, iamOCListCmd,
		iamOCUpdateCmd, iamOCUndeleteCmd}
	for _, c := range ocAll {
		c.Flags().StringVar(&flagIAMOCLocation, "location", "global", "Location containing the OAuth client")
	}
	for _, c := range []*cobra.Command{iamOCCreateCmd, iamOCUpdateCmd} {
		c.Flags().StringVar(&flagIAMOCConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the OauthClient body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	iamOCUpdateCmd.Flags().StringVar(&flagIAMOCUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{iamOCDescribeCmd, iamOCListCmd} {
		c.Flags().StringVar(&flagIAMOCFormat, "format", "", "Output format")
	}
	iamOauthClientsCmd.AddCommand(ocAll...)

	// credentials (nested)
	credAll := []*cobra.Command{iamOCCredCreateCmd, iamOCCredDeleteCmd, iamOCCredDescribeCmd,
		iamOCCredListCmd, iamOCCredUpdateCmd}
	for _, c := range credAll {
		c.Flags().StringVar(&flagIAMOCLocation, "location", "global", "Location containing the OAuth client")
		c.Flags().StringVar(&flagIAMOCClient, "oauth-client", "", "OAuth client containing the credential (required)")
		_ = c.MarkFlagRequired("oauth-client")
	}
	for _, c := range []*cobra.Command{iamOCCredCreateCmd, iamOCCredUpdateCmd} {
		c.Flags().StringVar(&flagIAMOCConfigFile, "config-file", "",
			"Path to a JSON/YAML file with the OauthClientCredential body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	iamOCCredUpdateCmd.Flags().StringVar(&flagIAMOCUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	for _, c := range []*cobra.Command{iamOCCredDescribeCmd, iamOCCredListCmd} {
		c.Flags().StringVar(&flagIAMOCFormat, "format", "", "Output format")
	}
	iamOauthClientCredsCmd.AddCommand(credAll...)
	iamOauthClientsCmd.AddCommand(iamOauthClientCredsCmd)

	iamCmd.AddCommand(iamOauthClientsCmd)
}

func iamOCParent(project string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, flagIAMOCLocation)
}

func iamOCName(project, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/oauthClients/%s", iamOCParent(project), id)
}

func iamOCCredParent(project string) string {
	return iamOCName(project, flagIAMOCClient)
}

func iamOCCredName(project, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/credentials/%s", iamOCCredParent(project), id)
}

func runIAMOCCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	c := &iam.OauthClient{}
	if err := loadYAMLOrJSONInto(flagIAMOCConfigFile, c); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OauthClients.Create(iamOCParent(project), c).
		OauthClientId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating oauth client: %w", err)
	}
	return emitFormatted(got, "")
}

func runIAMOCDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OauthClients.Delete(iamOCName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting oauth client: %w", err)
	}
	return emitFormatted(got, "")
}

func runIAMOCDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OauthClients.Get(iamOCName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing oauth client: %w", err)
	}
	return emitFormatted(got, flagIAMOCFormat)
}

func runIAMOCList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*iam.OauthClient
	pageToken := ""
	for {
		call := svc.Projects.Locations.OauthClients.List(iamOCParent(project)).Context(ctx)
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
	if flagIAMOCFormat != "" {
		return emitFormatted(all, flagIAMOCFormat)
	}
	fmt.Printf("%-40s %s\n", "NAME", "DISPLAY_NAME")
	for _, c := range all {
		fmt.Printf("%-40s %s\n", path.Base(c.Name), c.DisplayName)
	}
	return nil
}

func runIAMOCUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	c := &iam.OauthClient{}
	if err := loadYAMLOrJSONInto(flagIAMOCConfigFile, c); err != nil {
		return err
	}
	mask := flagIAMOCUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(c))
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OauthClients.Patch(iamOCName(project, args[0]), c).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating oauth client: %w", err)
	}
	return emitFormatted(got, "")
}

func runIAMOCUndelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OauthClients.Undelete(iamOCName(project, args[0]),
		&iam.UndeleteOauthClientRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeleting oauth client: %w", err)
	}
	return emitFormatted(got, "")
}

// --- credentials impl ---

func runIAMOCCredCreate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	c := &iam.OauthClientCredential{}
	if err := loadYAMLOrJSONInto(flagIAMOCConfigFile, c); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OauthClients.Credentials.Create(iamOCCredParent(project), c).
		OauthClientCredentialId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating oauth client credential: %w", err)
	}
	return emitFormatted(got, "")
}

func runIAMOCCredDelete(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Locations.OauthClients.Credentials.Delete(iamOCCredName(project, args[0])).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting oauth client credential: %w", err)
	}
	fmt.Printf("Deleted oauth client credential [%s].\n", args[0])
	return nil
}

func runIAMOCCredDescribe(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OauthClients.Credentials.Get(iamOCCredName(project, args[0])).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing oauth client credential: %w", err)
	}
	return emitFormatted(got, flagIAMOCFormat)
}

func runIAMOCCredList(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	resp, err := svc.Projects.Locations.OauthClients.Credentials.List(iamOCCredParent(project)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("listing oauth client credentials: %w", err)
	}
	if flagIAMOCFormat != "" {
		return emitFormatted(resp.OauthClientCredentials, flagIAMOCFormat)
	}
	fmt.Printf("%-40s\n", "NAME")
	for _, c := range resp.OauthClientCredentials {
		fmt.Println(path.Base(c.Name))
	}
	return nil
}

func runIAMOCCredUpdate(cmd *cobra.Command, args []string) error {
	project, err := resolveProject()
	if err != nil {
		return err
	}
	c := &iam.OauthClientCredential{}
	if err := loadYAMLOrJSONInto(flagIAMOCConfigFile, c); err != nil {
		return err
	}
	mask := flagIAMOCUpdateMask
	if mask == "" {
		mask = joinMask(nonEmptyJSONFields(c))
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Locations.OauthClients.Credentials.Patch(iamOCCredName(project, args[0]), c).
		UpdateMask(mask).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating oauth client credential: %w", err)
	}
	return emitFormatted(got, "")
}
