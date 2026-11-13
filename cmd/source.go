package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud source (#385) ---

func (s *sourceCmds) initSource() {

}

// sourceProjectName returns "projects/PROJECT" for the resolved project.
func (s *sourceCmds) sourceProjectName() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return "projects/" + project, nil
}

// sourceRepoName returns projects/PROJECT/repos/NAME. If name is already
// fully qualified it is returned as-is.
func (s *sourceCmds) sourceRepoName(name string) (string, error) {
	if strings.HasPrefix(name, "projects/") {
		return name, nil
	}
	project, err := s.sourceProjectName()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/repos/%s", project, name), nil
}

// sourceCmds holds the commands and flag values of one invocation's "source"
// command group, so invocations don't share state (#2070).
type sourceCmds struct {
	inv *invocation

	sourceCmd                *cobra.Command
	sourceProjectConfigsCmd  *cobra.Command
	flagSourcePCFormat       string
	flagSourcePCConfigFile   string
	flagSourcePCUpdateMask   string
	sourcePCDescribeCmd      *cobra.Command
	sourcePCUpdateCmd        *cobra.Command
	sourceReposCmd           *cobra.Command
	flagSourceRepoFormat     string
	flagSourceRepoConfigFile string
	flagSourceRepoUpdateMask string
	flagSourceRepoPageSize   int64
	sourceRepoCreateCmd      *cobra.Command
	sourceRepoDeleteCmd      *cobra.Command
	sourceRepoDescribeCmd    *cobra.Command
	sourceRepoListCmd        *cobra.Command
	sourceRepoUpdateCmd      *cobra.Command
	sourceRepoGetIamCmd      *cobra.Command
	sourceRepoSetIamCmd      *cobra.Command
}

// newSourceCmd returns the "source" command group for inv.
func newSourceCmd(inv *invocation) *cobra.Command {
	return newSourceCmds(inv).sourceCmd
}

// newSourceCmds builds the "source" command group for inv and returns its state.
func newSourceCmds(inv *invocation) *sourceCmds {
	s := &sourceCmds{inv: inv}
	s.sourceCmd = &cobra.Command{Use: "source", Short: "Manage Cloud Source Repositories"}
	s.sourceProjectConfigsCmd = &cobra.Command{Use: "project-configs", Short: "Manage per-project source repository configuration"}
	s.sourcePCDescribeCmd = &cobra.Command{
		Use: "describe", Short: "Describe the source repositories project config",
		Args: cobra.NoArgs, RunE: s.runSourcePCDescribe,
	}
	s.sourcePCUpdateCmd = &cobra.Command{
		Use: "update", Short: "Update the source repositories project config",
		Args: cobra.NoArgs, RunE: s.runSourcePCUpdate,
	}
	s.sourceReposCmd = &cobra.Command{Use: "repos", Short: "Manage Cloud Source Repositories"}
	s.sourceRepoCreateCmd = &cobra.Command{
		Use: "create REPO", Short: "Create a source repository",
		Args: cobra.ExactArgs(1), RunE: s.runSourceRepoCreate,
	}
	s.sourceRepoDeleteCmd = &cobra.Command{
		Use: "delete REPO", Short: "Delete a source repository",
		Args: cobra.ExactArgs(1), RunE: s.runSourceRepoDelete,
	}
	s.sourceRepoDescribeCmd = &cobra.Command{
		Use: "describe REPO", Short: "Describe a source repository",
		Args: cobra.ExactArgs(1), RunE: s.runSourceRepoDescribe,
	}
	s.sourceRepoListCmd = &cobra.Command{
		Use: "list", Short: "List source repositories in the current project",
		Args: cobra.NoArgs, RunE: s.runSourceRepoList,
	}
	s.sourceRepoUpdateCmd = &cobra.Command{
		Use: "update REPO", Short: "Update a source repository",
		Args: cobra.ExactArgs(1), RunE: s.runSourceRepoUpdate,
	}
	s.sourceRepoGetIamCmd = &cobra.Command{
		Use: "get-iam-policy REPO", Short: "Get the IAM policy for a source repository",
		Args: cobra.ExactArgs(1), RunE: s.runSourceRepoGetIam,
	}
	s.sourceRepoSetIamCmd = &cobra.Command{
		Use: "set-iam-policy REPO POLICY_FILE", Short: "Set the IAM policy for a source repository",
		Args: cobra.ExactArgs(2), RunE: s.runSourceRepoSetIam,
	}
	s.initSource()
	s.initSourceProjectConfigs()
	s.initSourceRepos()
	return s
}
