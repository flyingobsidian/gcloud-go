package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	binaryauthorization "google.golang.org/api/binaryauthorization/v1"
)

// --- gcloud container binauthz (#1135) ---
//
// Wires the core Binary Authorization surface: the project-scoped policy
// singleton, and the attestors subgroup.

func (s *containerCmds) initContainerBinauthz() {
	for _, c := range []*cobra.Command{
		s.binauthzPolicyDescribeCmd, s.binauthzPolicyUpdateCmd,
		s.binauthzAttestorsCreateCmd, s.binauthzAttestorsDeleteCmd,
		s.binauthzAttestorsDescribeCmd, s.binauthzAttestorsListCmd, s.binauthzAttestorsUpdateCmd,
	} {
		c.Flags().StringVar(&s.flagBaFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{
		s.binauthzPolicyUpdateCmd, s.binauthzAttestorsCreateCmd, s.binauthzAttestorsUpdateCmd,
	} {
		c.Flags().StringVar(&s.flagBaConfigFile, "config-file", "", "YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.binauthzAttestorsListCmd.Flags().Int64Var(&s.flagBaPageSize, "page-size", 0, "Maximum results per page")

	s.containerBinauthzPolicyCmd.AddCommand(s.binauthzPolicyDescribeCmd, s.binauthzPolicyUpdateCmd)
	s.containerBinauthzAttestorsCmd.AddCommand(
		s.binauthzAttestorsCreateCmd, s.binauthzAttestorsDeleteCmd,
		s.binauthzAttestorsDescribeCmd, s.binauthzAttestorsListCmd, s.binauthzAttestorsUpdateCmd,
	)
	s.containerBinauthzCmd.AddCommand(s.containerBinauthzPolicyCmd, s.containerBinauthzAttestorsCmd)
	s.containerCmd.AddCommand(s.containerBinauthzCmd)
}

func (s *containerCmds) baProjectName() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return "projects/" + project, nil
}

func (s *containerCmds) baPolicyName() (string, error) {
	p, err := s.baProjectName()
	if err != nil {
		return "", err
	}
	return p + "/policy", nil
}

func (s *containerCmds) baAttestorName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	p, err := s.baProjectName()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/attestors/%s", p, id), nil
}

// policy

func (s *containerCmds) runBaPolicyDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.baPolicyName()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BinaryAuthorization(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.GetPolicy(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBaFormat)
}

func (s *containerCmds) runBaPolicyUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.baPolicyName()
	if err != nil {
		return err
	}
	body := &binaryauthorization.Policy{}
	if err := loadYAMLOrJSONInto(s.flagBaConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().BinaryAuthorization(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.UpdatePolicy(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating policy: %w", err)
	}
	fmt.Fprintln(s.inv.stdout, "Updated Binary Authorization policy.")
	return emitFormattedTo(s.inv.stdout, got, s.flagBaFormat)
}

// attestors

func (s *containerCmds) runBaAttestorsCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.baProjectName()
	if err != nil {
		return err
	}
	body := &binaryauthorization.Attestor{}
	if err := loadYAMLOrJSONInto(s.flagBaConfigFile, body); err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BinaryAuthorization(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Attestors.Create(parent, body).AttestorId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating attestor: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Created attestor [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagBaFormat)
}

func (s *containerCmds) runBaAttestorsDelete(cmd *cobra.Command, args []string) error {
	name, err := s.baAttestorName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BinaryAuthorization(ctx)
	if err != nil {
		return err
	}
	if _, err := svc.Projects.Attestors.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("deleting attestor: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Deleted attestor [%s].\n", args[0])
	return nil
}

func (s *containerCmds) runBaAttestorsDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.baAttestorName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BinaryAuthorization(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Attestors.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing attestor: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagBaFormat)
}

func (s *containerCmds) runBaAttestorsList(cmd *cobra.Command, args []string) error {
	parent, err := s.baProjectName()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().BinaryAuthorization(ctx)
	if err != nil {
		return err
	}
	var all []*binaryauthorization.Attestor
	pageToken := ""
	for {
		call := svc.Projects.Attestors.List(parent).Context(ctx)
		if s.flagBaPageSize > 0 {
			call = call.PageSize(s.flagBaPageSize)
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
	return emitFormattedTo(s.inv.stdout, all, s.flagBaFormat)
}

func (s *containerCmds) runBaAttestorsUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.baAttestorName(args[0])
	if err != nil {
		return err
	}
	body := &binaryauthorization.Attestor{}
	if err := loadYAMLOrJSONInto(s.flagBaConfigFile, body); err != nil {
		return err
	}
	body.Name = name
	ctx := cmd.Context()
	svc, err := s.inv.clients().BinaryAuthorization(ctx)
	if err != nil {
		return err
	}
	got, err := svc.Projects.Attestors.Update(name, body).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating attestor: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Updated attestor [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, got, s.flagBaFormat)
}
