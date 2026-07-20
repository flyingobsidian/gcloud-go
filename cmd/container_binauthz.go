package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	binaryauthorization "google.golang.org/api/binaryauthorization/v1"
)

// --- gcloud container binauthz (#1135) ---
//
// Wires the core Binary Authorization surface: the project-scoped policy
// singleton, and the attestors subgroup.

var containerBinauthzCmd = &cobra.Command{Use: "binauthz", Short: "Manage Binary Authorization"}
var containerBinauthzPolicyCmd = &cobra.Command{Use: "policy", Short: "Manage the Binary Authorization policy singleton for the project"}
var containerBinauthzAttestorsCmd = &cobra.Command{Use: "attestors", Short: "Manage Binary Authorization attestors"}

var (
	flagBaFormat     string
	flagBaConfigFile string
	flagBaPageSize   int64
)

var (
	binauthzPolicyDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the project's Binary Authorization policy",
		Args: cobra.NoArgs, RunE: runBaPolicyDescribe,
	}
	binauthzPolicyUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update the project's Binary Authorization policy (loads body from --config-file)",
		Args: cobra.NoArgs, RunE: runBaPolicyUpdate,
	}

	binauthzAttestorsCreateCmd = &cobra.Command{
		Use: "create ATTESTOR", Short: "Create an attestor (loads body from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runBaAttestorsCreate,
	}
	binauthzAttestorsDeleteCmd = &cobra.Command{
		Use: "delete ATTESTOR", Short: "Delete an attestor",
		Args: cobra.ExactArgs(1), RunE: runBaAttestorsDelete,
	}
	binauthzAttestorsDescribeCmd = &cobra.Command{
		Use: "describe ATTESTOR", Short: "Describe an attestor",
		Args: cobra.ExactArgs(1), RunE: runBaAttestorsDescribe,
	}
	binauthzAttestorsListCmd = &cobra.Command{
		Use: "list", Short: "List attestors in the current project",
		Args: cobra.NoArgs, RunE: runBaAttestorsList,
	}
	binauthzAttestorsUpdateCmd = &cobra.Command{
		Use: "update ATTESTOR", Short: "Update an attestor (loads body from --config-file)",
		Args: cobra.ExactArgs(1), RunE: runBaAttestorsUpdate,
	}
)

func init() {
	for _, c := range []*cobra.Command{
		binauthzPolicyDescribeCmd, binauthzPolicyUpdateCmd,
		binauthzAttestorsCreateCmd, binauthzAttestorsDeleteCmd,
		binauthzAttestorsDescribeCmd, binauthzAttestorsListCmd, binauthzAttestorsUpdateCmd,
	} {
		c.Flags().StringVar(&flagBaFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		binauthzPolicyUpdateCmd, binauthzAttestorsCreateCmd, binauthzAttestorsUpdateCmd,
	} {
		c.Flags().StringVar(&flagBaConfigFile, "config-file", "", "YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	binauthzAttestorsListCmd.Flags().Int64Var(&flagBaPageSize, "page-size", 0, "Maximum results per page")

	containerBinauthzPolicyCmd.AddCommand(binauthzPolicyDescribeCmd, binauthzPolicyUpdateCmd)
	containerBinauthzAttestorsCmd.AddCommand(
		binauthzAttestorsCreateCmd, binauthzAttestorsDeleteCmd,
		binauthzAttestorsDescribeCmd, binauthzAttestorsListCmd, binauthzAttestorsUpdateCmd,
	)
	containerBinauthzCmd.AddCommand(containerBinauthzPolicyCmd, containerBinauthzAttestorsCmd)
	containerCmd.AddCommand(containerBinauthzCmd)
}

func baProjectName() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return "projects/" + project, nil
}

func baPolicyName() (string, error) {
	p, err := baProjectName()
	if err != nil {
		return "", err
	}
	return p + "/policy", nil
}

func baAttestorName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	p, err := baProjectName()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/attestors/%s", p, id), nil
}

// policy

func runBaPolicyDescribe(cmd *cobra.Command, args []string) error {
	name, err := baPolicyName()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BinaryAuthorizationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.GetPolicy(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing policy: %w", err)
	}
	return emitFormatted(got, flagBaFormat)
}

func runBaPolicyUpdate(cmd *cobra.Command, args []string) error {
	name, err := baPolicyName()
	if err != nil {
		return err
	}
	body := &binaryauthorization.Policy{}
	if err := loadYAMLOrJSONInto(flagBaConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.BinaryAuthorizationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.UpdatePolicy(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating policy: %w", err)
	}
	fmt.Println("Updated Binary Authorization policy.")
	return emitFormatted(got, flagBaFormat)
}

// attestors

func runBaAttestorsCreate(cmd *cobra.Command, args []string) error {
	parent, err := baProjectName()
	if err != nil {
		return err
	}
	body := &binaryauthorization.Attestor{}
	if err := loadYAMLOrJSONInto(flagBaConfigFile, body); err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BinaryAuthorizationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Attestors.Create(parent, body).AttestorId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating attestor: %w", err)
	}
	fmt.Printf("Created attestor [%s].\n", args[0])
	return emitFormatted(got, flagBaFormat)
}

func runBaAttestorsDelete(cmd *cobra.Command, args []string) error {
	name, err := baAttestorName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BinaryAuthorizationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Attestors.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting attestor: %w", err)
	}
	fmt.Printf("Deleted attestor [%s].\n", args[0])
	return nil
}

func runBaAttestorsDescribe(cmd *cobra.Command, args []string) error {
	name, err := baAttestorName(args[0])
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BinaryAuthorizationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Attestors.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing attestor: %w", err)
	}
	return emitFormatted(got, flagBaFormat)
}

func runBaAttestorsList(cmd *cobra.Command, args []string) error {
	parent, err := baProjectName()
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := gcp.BinaryAuthorizationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	var all []*binaryauthorization.Attestor
	pageToken := ""
	for {
		call := svc.Projects.Attestors.List(parent).Context(ctx)
		if flagBaPageSize > 0 {
			call = call.PageSize(flagBaPageSize)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing attestors: %w", err)
		}
		all = append(all, resp.Attestors...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	return emitFormatted(all, flagBaFormat)
}

func runBaAttestorsUpdate(cmd *cobra.Command, args []string) error {
	name, err := baAttestorName(args[0])
	if err != nil {
		return err
	}
	body := &binaryauthorization.Attestor{}
	if err := loadYAMLOrJSONInto(flagBaConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	ctx := context.Background()
	svc, err := gcp.BinaryAuthorizationService(ctx, flagAccount)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Attestors.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating attestor: %w", err)
	}
	fmt.Printf("Updated attestor [%s].\n", args[0])
	return emitFormatted(got, flagBaFormat)
}
