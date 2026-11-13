package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	apigee "google.golang.org/api/apigee/v1"
)

// --- gcloud apigee (#297) ---

func (s *apigeeCmds) initApigee() {

}

// apigeeOrgName returns "organizations/ORG" for the given --organization
// flag value. Empty is not allowed — apigee is not project-scoped.
func apigeeOrgName(org string) (string, error) {
	if org == "" {
		return "", fmt.Errorf("--organization is required")
	}
	if strings.HasPrefix(org, "organizations/") {
		return org, nil
	}
	return "organizations/" + org, nil
}

// apigeeOrgID returns ORG from an "organizations/ORG[/...]" resource name.
func apigeeOrgID(resource string) string {
	rest := strings.TrimPrefix(resource, "organizations/")
	org, _, _ := strings.Cut(rest, "/")
	return org
}

// apigeeServiceFor returns an Apigee client routed to the regional endpoint
// for the organization that owns resource ("organizations/ORG/..."),
// matching gcloud-python's per-organization location routing.
func (s *apigeeCmds) apigeeServiceFor(ctx context.Context, resource string) (*apigee.Service, error) {
	return gcp.ApigeeServiceForOrg(ctx, s.inv.flags.account, apigeeOrgID(resource))
}

// apigeeResource returns organizations/ORG/COLLECTION/ID.
func apigeeResource(org, collection, id string) (string, error) {
	if strings.HasPrefix(id, "organizations/") {
		return id, nil
	}
	parent, err := apigeeOrgName(org)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id), nil
}

// apigeeCmds holds the commands and flag values of one invocation's "apigee"
// command group, so invocations don't share state (#2070).
type apigeeCmds struct {
	inv *invocation

	apigeeCmd                    *cobra.Command
	apigeeApisCmd                *cobra.Command
	flagApigeeApiOrganization    string
	flagApigeeApiFormat          string
	flagApigeeApiFromBundle      string
	flagApigeeApiFromTemplate    string
	flagApigeeApiSpace           string
	apigeeApisDeleteCmd          *cobra.Command
	apigeeApisDescribeCmd        *cobra.Command
	apigeeApisListCmd            *cobra.Command
	apigeeApisImportCmd          *cobra.Command
	apigeeApplicationsCmd        *cobra.Command
	flagApigeeAppOrganization    string
	flagApigeeAppFormat          string
	flagApigeeAppPageSize        int64
	apigeeAppDescribeCmd         *cobra.Command
	apigeeAppListCmd             *cobra.Command
	apigeeArchivesCmd            *cobra.Command
	flagApigeeArcOrganization    string
	flagApigeeArcEnvironment     string
	flagApigeeArcFormat          string
	flagApigeeArcConfigFile      string
	flagApigeeArcPageSize        int64
	apigeeArcCreateCmd           *cobra.Command
	apigeeArcDeleteCmd           *cobra.Command
	apigeeArcDescribeCmd         *cobra.Command
	apigeeArcListCmd             *cobra.Command
	apigeeArcUpdateCmd           *cobra.Command
	apigeeDeploymentsCmd         *cobra.Command
	flagApigeeDeployOrganization string
	flagApigeeDeployFormat       string
	apigeeDeployListCmd          *cobra.Command
	apigeeDevelopersCmd          *cobra.Command
	flagApigeeDevOrganization    string
	flagApigeeDevFormat          string
	flagApigeeDevConfigFile      string
	apigeeDevCreateCmd           *cobra.Command
	apigeeDevDeleteCmd           *cobra.Command
	apigeeDevDescribeCmd         *cobra.Command
	apigeeDevListCmd             *cobra.Command
	apigeeDevUpdateCmd           *cobra.Command
	apigeeEnvironmentsCmd        *cobra.Command
	flagApigeeEnvOrganization    string
	flagApigeeEnvFormat          string
	flagApigeeEnvConfigFile      string
	apigeeEnvCreateCmd           *cobra.Command
	apigeeEnvDeleteCmd           *cobra.Command
	apigeeEnvDescribeCmd         *cobra.Command
	apigeeEnvUpdateCmd           *cobra.Command
	apigeeOperationsCmd          *cobra.Command
	flagApigeeOpOrganization     string
	flagApigeeOpFormat           string
	flagApigeeOpPageSize         int64
	apigeeOpDescribeCmd          *cobra.Command
	apigeeOpListCmd              *cobra.Command
	apigeeOrganizationsCmd       *cobra.Command
	flagApigeeOrgFormat          string
	flagApigeeOrgConfigFile      string
	apigeeOrgDescribeCmd         *cobra.Command
	apigeeOrgListCmd             *cobra.Command
	apigeeOrgProvisionCmd        *cobra.Command
	apigeeProductsCmd            *cobra.Command
	flagApigeePrdOrganization    string
	flagApigeePrdFormat          string
	flagApigeePrdConfigFile      string
	apigeePrdCreateCmd           *cobra.Command
	apigeePrdDeleteCmd           *cobra.Command
	apigeePrdDescribeCmd         *cobra.Command
	apigeePrdListCmd             *cobra.Command
	apigeePrdUpdateCmd           *cobra.Command
}

// newApigeeCmd returns the "apigee" command group for inv.
func newApigeeCmd(inv *invocation) *cobra.Command {
	return newApigeeCmds(inv).apigeeCmd
}

// newApigeeCmds builds the "apigee" command group for inv and returns its state.
func newApigeeCmds(inv *invocation) *apigeeCmds {
	s := &apigeeCmds{inv: inv}
	s.apigeeCmd = &cobra.Command{Use: "apigee", Short: "Manage Apigee"}
	s.apigeeApisCmd = &cobra.Command{Use: "apis", Short: "Manage Apigee API proxies"}
	s.apigeeApisDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an API proxy",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeApisDelete,
	}
	s.apigeeApisDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an API proxy",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeApisDescribe,
	}
	s.apigeeApisListCmd = &cobra.Command{
		Use: "list", Short: "List API proxies in an organization",
		Args: cobra.NoArgs, RunE: s.runApigeeApisList,
	}
	s.apigeeApisImportCmd = &cobra.Command{
		Use: "import API", Short: "Import an API proxy from a local bundle or feature template",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeApisImport,
	}
	s.apigeeApplicationsCmd = &cobra.Command{Use: "applications", Short: "Manage Apigee applications"}
	s.apigeeAppDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an Apigee application",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeAppDescribe,
	}
	s.apigeeAppListCmd = &cobra.Command{
		Use: "list", Short: "List Apigee applications in an organization",
		Args: cobra.NoArgs, RunE: s.runApigeeAppList,
	}
	s.apigeeArchivesCmd = &cobra.Command{Use: "archives", Short: "Manage Apigee archive deployments"}
	s.apigeeArcCreateCmd = &cobra.Command{
		Use: "create RESOURCE", Short: "Create an archive deployment",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeArcCreate,
	}
	s.apigeeArcDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an archive deployment",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeArcDelete,
	}
	s.apigeeArcDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an archive deployment",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeArcDescribe,
	}
	s.apigeeArcListCmd = &cobra.Command{
		Use: "list", Short: "List archive deployments in an environment",
		Args: cobra.NoArgs, RunE: s.runApigeeArcList,
	}
	s.apigeeArcUpdateCmd = &cobra.Command{
		Use: "update RESOURCE", Short: "Update an archive deployment",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeArcUpdate,
	}
	s.apigeeDeploymentsCmd = &cobra.Command{Use: "deployments", Short: "Manage Apigee deployments"}
	s.apigeeDeployListCmd = &cobra.Command{
		Use: "list", Short: "List deployments in an organization",
		Args: cobra.NoArgs, RunE: s.runApigeeDeployList,
	}
	s.apigeeDevelopersCmd = &cobra.Command{Use: "developers", Short: "Manage Apigee developers"}
	s.apigeeDevCreateCmd = &cobra.Command{
		Use: "create RESOURCE", Short: "Create an Apigee developer",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeDevCreate,
	}
	s.apigeeDevDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an Apigee developer",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeDevDelete,
	}
	s.apigeeDevDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an Apigee developer",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeDevDescribe,
	}
	s.apigeeDevListCmd = &cobra.Command{
		Use: "list", Short: "List Apigee developers in an organization",
		Args: cobra.NoArgs, RunE: s.runApigeeDevList,
	}
	s.apigeeDevUpdateCmd = &cobra.Command{
		Use: "update RESOURCE", Short: "Update an Apigee developer",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeDevUpdate,
	}
	s.apigeeEnvironmentsCmd = &cobra.Command{Use: "environments", Short: "Manage Apigee environments"}
	s.apigeeEnvCreateCmd = &cobra.Command{
		Use: "create RESOURCE", Short: "Create an Apigee environment",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeEnvCreate,
	}
	s.apigeeEnvDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an Apigee environment",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeEnvDelete,
	}
	s.apigeeEnvDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an Apigee environment",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeEnvDescribe,
	}
	s.apigeeEnvUpdateCmd = &cobra.Command{
		Use: "update RESOURCE", Short: "Update an Apigee environment",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeEnvUpdate,
	}
	s.apigeeOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Apigee long-running operations"}
	s.apigeeOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an Apigee operation",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeOpDescribe,
	}
	s.apigeeOpListCmd = &cobra.Command{
		Use: "list", Short: "List Apigee operations in an organization",
		Args: cobra.NoArgs, RunE: s.runApigeeOpList,
	}
	s.apigeeOrganizationsCmd = &cobra.Command{Use: "organizations", Short: "Manage Apigee organizations"}
	s.apigeeOrgDescribeCmd = &cobra.Command{
		Use: "describe ORGANIZATION", Short: "Describe an Apigee organization",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeOrgDescribe,
	}
	s.apigeeOrgListCmd = &cobra.Command{
		Use: "list", Short: "List Apigee organizations you have access to",
		Args: cobra.NoArgs, RunE: s.runApigeeOrgList,
	}
	s.apigeeOrgProvisionCmd = &cobra.Command{
		Use: "provision PROJECT", Short: "Provision an Apigee organization for a GCP project",
		Args: cobra.ExactArgs(1), RunE: s.runApigeeOrgProvision,
	}
	s.apigeeProductsCmd = &cobra.Command{Use: "products", Short: "Manage Apigee API products"}
	s.apigeePrdCreateCmd = &cobra.Command{
		Use: "create RESOURCE", Short: "Create an API product",
		Args: cobra.ExactArgs(1), RunE: s.runApigeePrdCreate,
	}
	s.apigeePrdDeleteCmd = &cobra.Command{
		Use: "delete RESOURCE", Short: "Delete an API product",
		Args: cobra.ExactArgs(1), RunE: s.runApigeePrdDelete,
	}
	s.apigeePrdDescribeCmd = &cobra.Command{
		Use: "describe RESOURCE", Short: "Describe an API product",
		Args: cobra.ExactArgs(1), RunE: s.runApigeePrdDescribe,
	}
	s.apigeePrdListCmd = &cobra.Command{
		Use: "list", Short: "List API products in an organization",
		Args: cobra.NoArgs, RunE: s.runApigeePrdList,
	}
	s.apigeePrdUpdateCmd = &cobra.Command{
		Use: "update RESOURCE", Short: "Update an API product",
		Args: cobra.ExactArgs(1), RunE: s.runApigeePrdUpdate,
	}
	s.initApigee()
	s.initApigeeApis()
	s.initApigeeApplications()
	s.initApigeeArchives()
	s.initApigeeDeployments()
	s.initApigeeDevelopers()
	s.initApigeeEnvironments()
	s.initApigeeOperations()
	s.initApigeeOrganizations()
	s.initApigeeProducts()
	return s
}
