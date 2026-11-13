package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud memorystore acl-policies (#976) ---

func (s *memorystoreCmds) initMemorystoreAclPolicies() {
	all := []*cobra.Command{s.memstoreACLCreateCmd, s.memstoreACLDeleteCmd, s.memstoreACLDescribeCmd, s.memstoreACLListCmd, s.memstoreACLUpdateCmd}
	for _, c := range all {
		c.Flags().StringVar(&s.flagMemstoreACLLocation, "location", "", "Memorystore location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagMemstoreACLFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.memstoreACLCreateCmd, s.memstoreACLUpdateCmd} {
		c.Flags().StringVar(&s.flagMemstoreACLConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the ACL policy body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	s.memstoreACLUpdateCmd.Flags().StringVar(&s.flagMemstoreACLUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.memstoreACLListCmd.Flags().Int64Var(&s.flagMemstoreACLPageSize, "page-size", 0, "Maximum results per page")

	s.memstoreACLCmd.AddCommand(all...)
	s.memorystoreCmd.AddCommand(s.memstoreACLCmd)
}

func (s *memorystoreCmds) memstoreACLParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagMemstoreACLLocation), nil
}

func (s *memorystoreCmds) memstoreACLName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.memstoreACLParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/aclPolicies/%s", parent, id), nil
}

func (s *memorystoreCmds) runMemstoreACLCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.memstoreACLParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagMemstoreACLConfigFile, &body); err != nil {
		return err
	}
	q := url.Values{}
	q.Set("aclPolicyId", args[0])
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(memorystoreRest).do(ctx, http.MethodPost, "/"+parent+"/aclPolicies", q, body, &op); err != nil {
		return fmt.Errorf("creating ACL policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for ACL policy [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagMemstoreACLFormat)
}

func (s *memorystoreCmds) runMemstoreACLDelete(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreACLName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(memorystoreRest).do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting ACL policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for ACL policy [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagMemstoreACLFormat)
}

func (s *memorystoreCmds) runMemstoreACLDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreACLName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(memorystoreRest).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing ACL policy: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMemstoreACLFormat)
}

func (s *memorystoreCmds) runMemstoreACLList(cmd *cobra.Command, args []string) error {
	parent, err := s.memstoreACLParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(memorystoreRest).paginate(ctx, "/"+parent+"/aclPolicies", nil, "aclPolicies", s.flagMemstoreACLPageSize)
	if err != nil {
		return fmt.Errorf("listing ACL policies: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagMemstoreACLFormat)
}

func (s *memorystoreCmds) runMemstoreACLUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreACLName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagMemstoreACLConfigFile, &body); err != nil {
		return err
	}
	mask := s.flagMemstoreACLUpdateMask
	if mask == "" {
		mask = joinMask(dcTopLevelKeys(body))
	}
	q := url.Values{}
	if mask != "" {
		q.Set("updateMask", mask)
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(memorystoreRest).do(ctx, http.MethodPatch, "/"+name, q, body, &op); err != nil {
		return fmt.Errorf("updating ACL policy: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for ACL policy [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagMemstoreACLFormat)
}
