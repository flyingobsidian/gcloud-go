package cmd

import (
	"context"
	"errors"
	"fmt"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	iam "google.golang.org/api/iam/v1"
)

// --- iam roles list (#2008) ---

var (
	flagRolesListOrganization string
	flagRolesListShowDeleted  bool
	flagRolesListLimit        int64
)

var roleListCmd = &cobra.Command{
	Use:   "list",
	Short: "List predefined roles, or the custom roles for an organization or project",
	Args:  cobra.NoArgs,
	RunE:  runRolesList,
}

func init() {
	roleListCmd.Flags().StringVar(&flagRolesListOrganization, "organization", "",
		"Organization whose custom roles to list")
	roleListCmd.Flags().BoolVar(&flagRolesListShowDeleted, "show-deleted", false, "Show deleted roles")
	roleListCmd.Flags().Int64Var(&flagRolesListLimit, "limit", 0, "Maximum number of roles to list (0 = no limit)")
	rolesCmd.AddCommand(roleListCmd)
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

func runRolesList(cmd *cobra.Command, args []string) error {
	projectSet := cmd.Flags().Changed("project")
	parent, err := rolesListParent(flagRolesListOrganization, flagProject, projectSet)
	if err != nil {
		return err
	}
	if parent != "" && cmd.Flags().Changed("limit") && flagRolesListLimit < 1 {
		return errors.New("Limit size must be >=1")
	}
	ctx := context.Background()
	svc, err := gcp.IAMService(ctx, flagAccount)
	if err != nil {
		return err
	}
	page := func(ctx context.Context, pageToken string) (*iam.ListRolesResponse, error) {
		switch {
		case parent == "":
			call := svc.Roles.List().ShowDeleted(flagRolesListShowDeleted).Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			return call.Do()
		case flagRolesListOrganization != "":
			call := svc.Organizations.Roles.List(parent).ShowDeleted(flagRolesListShowDeleted).Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			return call.Do()
		default:
			call := svc.Projects.Roles.List(parent).ShowDeleted(flagRolesListShowDeleted).Context(ctx)
			if pageToken != "" {
				call = call.PageToken(pageToken)
			}
			return call.Do()
		}
	}
	roles, err := listRoles(ctx, page, flagRolesListLimit)
	if err != nil {
		return err
	}
	return emitFormatted(roles, flagFormat)
}
