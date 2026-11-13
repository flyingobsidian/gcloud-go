package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	iam "google.golang.org/api/iam/v1"
)

// --- iam roles list (#2008) ---

func (s *iamCmds) initIamRolesList() {
	s.roleListCmd.Flags().StringVar(&s.flagRolesListOrganization, "organization", "",
		"Organization whose custom roles to list")
	s.roleListCmd.Flags().BoolVar(&s.flagRolesListShowDeleted, "show-deleted", false, "Show deleted roles")
	s.roleListCmd.Flags().Int64Var(&s.flagRolesListLimit, "limit", 0, "Maximum number of roles to list (0 = no limit)")
	s.rolesCmd.AddCommand(s.roleListCmd)
}

// rolesListParent returns the parent whose custom roles `iam roles list`
// lists, or "" for IAM's predefined roles. As in gcloud, only a --project
// given on the command line selects a project; core/project does not.
func rolesListParent(organization, project string, projectSet bool) (string, error) {
	switch {
	case organization != "" && projectSet:
		return "", errors.New("argument --organization: at most one of --organization | --project may be specified")
	case organization != "":
		return "organizations/" + organization, nil
	case projectSet && project != "":
		return "projects/" + project, nil
	}
	return "", nil
}

// rolesPager fetches one page of roles for the given page token.
type rolesPager func(ctx context.Context, pageToken string) (*iam.ListRolesResponse, error)

// listRoles pages through roles until they run out or limit roles have been
// collected (limit <= 0 means no limit). It never returns a nil slice.
func listRoles(ctx context.Context, page rolesPager, limit int64) ([]*iam.Role, error) {
	all := []*iam.Role{}
	pageToken := ""
	for {
		resp, err := page(ctx, pageToken)
		if err != nil {
			return nil, fmt.Errorf("listing roles: %w", err)
		}
		all = append(all, resp.Roles...)
		if limit > 0 && int64(len(all)) >= limit {
			return all[:limit], nil
		}
		if resp.NextPageToken == "" {
			return all, nil
		}
		pageToken = resp.NextPageToken
	}
}

func (s *iamCmds) runRolesList(cmd *cobra.Command, args []string) error {
	projectSet := cmd.Flags().Changed("project")
	parent, err := rolesListParent(s.flagRolesListOrganization, s.inv.flags.project, projectSet)
	if err != nil {
		return err
	}
	if parent != "" && cmd.Flags().Changed("limit") && s.flagRolesListLimit < 1 {
		return errors.New("Limit size must be >=1")
	}
	ctx := cmd.Context()
	svc, err := s.inv.clients().IAM(ctx)
	if err != nil {
		return err
	}
	page := func(ctx context.Context, pageToken string) (*iam.ListRolesResponse, error) {
		switch {
		case parent == "":
			call := svc.Roles.List().ShowDeleted(s.flagRolesListShowDeleted).Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			return call.Do()
		case s.flagRolesListOrganization != "":
			call := svc.Organizations.Roles.List(parent).ShowDeleted(s.flagRolesListShowDeleted).Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			return call.Do()
		default:
			call := svc.Projects.Roles.List(parent).ShowDeleted(s.flagRolesListShowDeleted).Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			return call.Do()
		}
	}
	roles, err := listRoles(ctx, page, s.flagRolesListLimit)
	if err != nil {
		return err
	}
	return emitFormattedTo(s.inv.stdout, roles, s.inv.flags.format)
}
