package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	datacatalog "google.golang.org/api/datacatalog/v1"
)

// --- gcloud data-catalog (#321) ---

func (s *dataCatalogCmds) initDataCatalog() {

}

func dcLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func dcChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func dcIamMemberFlags(c *cobra.Command, member, role, condExpr, condTitle, condDesc *string) {
	c.Flags().StringVar(member, "member", "", "IAM member (e.g. user:alice@example.com) (required)")
	c.Flags().StringVar(role, "role", "", "IAM role to bind (required)")
	c.Flags().StringVar(condExpr, "condition-expression", "", "CEL expression for a conditional binding")
	c.Flags().StringVar(condTitle, "condition-title", "", "Title for a conditional binding")
	c.Flags().StringVar(condDesc, "condition-description", "", "Description for a conditional binding")
	_ = c.MarkFlagRequired("member")
	_ = c.MarkFlagRequired("role")
}

func dcBuildCondition(expr, title, desc string) *datacatalog.Expr {
	if expr == "" && title == "" && desc == "" {
		return nil
	}
	return &datacatalog.Expr{Expression: expr, Title: title, Description: desc}
}

func dcCondsEqual(a, b *datacatalog.Expr) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Expression == b.Expression && a.Title == b.Title && a.Description == b.Description
}

func dcAddBinding(policy *datacatalog.Policy, role, member string, cond *datacatalog.Expr) {
	for _, b := range policy.Bindings {
		if b.Role != role || !dcCondsEqual(b.Condition, cond) {
			continue
		}
		for _, m := range b.Members {
			if m == member {
				return
			}
		}
		b.Members = append(b.Members, member)
		return
	}
	policy.Bindings = append(policy.Bindings, &datacatalog.Binding{
		Role:      role,
		Members:   []string{member},
		Condition: cond,
	})
}

func dcRemoveBinding(policy *datacatalog.Policy, role, member string, cond *datacatalog.Expr, allConds bool) bool {
	changed := false
	kept := policy.Bindings[:0]
	for _, b := range policy.Bindings {
		match := b.Role == role && (allConds || dcCondsEqual(b.Condition, cond))
		if !match {
			kept = append(kept, b)
			continue
		}
		newMembers := b.Members[:0]
		for _, m := range b.Members {
			if m == member {
				changed = true
				continue
			}
			newMembers = append(newMembers, m)
		}
		b.Members = newMembers
		if len(b.Members) > 0 {
			kept = append(kept, b)
		} else {
			changed = true
		}
	}
	policy.Bindings = kept
	return changed
}

func (s *dataCatalogCmds) dcUpdatedIam(who string) {
	fmt.Fprintf(s.inv.stderr, "Updated IAM policy for %s.\n", who)
}

// dataCatalogCmds holds the commands and flag values of one invocation's "dataCatalog"
// command group, so invocations don't share state (#2070).
type dataCatalogCmds struct {
	inv *invocation

	dataCatalogCmd            *cobra.Command
	dcEntriesCmd              *cobra.Command
	flagDCEntryLocation       string
	flagDCEntryGroup          string
	flagDCEntryFormat         string
	flagDCEntryConfigFile     string
	flagDCEntryUpdateMask     string
	flagDCEntryPageSize       int64
	flagDCEntryLookupSql      string
	flagDCEntryLookupLinked   string
	flagDCEntryLookupFQN      string
	flagDCEntryLookupProject  string
	flagDCEntryLookupLoc      string
	dcEntriesCreateCmd        *cobra.Command
	dcEntriesDeleteCmd        *cobra.Command
	dcEntriesDescribeCmd      *cobra.Command
	dcEntriesListCmd          *cobra.Command
	dcEntriesUpdateCmd        *cobra.Command
	dcEntriesLookupCmd        *cobra.Command
	dcEntriesStarCmd          *cobra.Command
	dcEntriesUnstarCmd        *cobra.Command
	dcEntryGroupsCmd          *cobra.Command
	flagDCEGLocation          string
	flagDCEGFormat            string
	flagDCEGConfigFile        string
	flagDCEGUpdateMask        string
	flagDCEGPageSize          int64
	flagDCEGIamMember         string
	flagDCEGIamRole           string
	flagDCEGIamCondExpr       string
	flagDCEGIamCondT          string
	flagDCEGIamCondD          string
	flagDCEGIamAllCond        bool
	dcEGCreateCmd             *cobra.Command
	dcEGDeleteCmd             *cobra.Command
	dcEGDescribeCmd           *cobra.Command
	dcEGListCmd               *cobra.Command
	dcEGUpdateCmd             *cobra.Command
	dcEGGetIamCmd             *cobra.Command
	dcEGSetIamCmd             *cobra.Command
	dcEGAddIamCmd             *cobra.Command
	dcEGRemoveIamCmd          *cobra.Command
	flagDCSearchQuery         string
	flagDCSearchProjects      []string
	flagDCSearchOrgIds        []string
	flagDCSearchGCPPublic     bool
	flagDCSearchOrderBy       string
	flagDCSearchPageSize      int64
	flagDCSearchAdminSearch   bool
	flagDCSearchIncludeTagsPT bool
	flagDCSearchFormat        string
	dataCatalogSearchCmd      *cobra.Command
	dcTTCmd                   *cobra.Command
	flagDCTTLocation          string
	flagDCTTFormat            string
	flagDCTTConfigFile        string
	flagDCTTUpdateMask        string
	flagDCTTForce             bool
	flagDCTTIamMember         string
	flagDCTTIamRole           string
	flagDCTTIamCondExpr       string
	flagDCTTIamCondT          string
	flagDCTTIamCondD          string
	flagDCTTIamAllCond        bool
	dcTTCreateCmd             *cobra.Command
	dcTTDeleteCmd             *cobra.Command
	dcTTDescribeCmd           *cobra.Command
	dcTTUpdateCmd             *cobra.Command
	dcTTGetIamCmd             *cobra.Command
	dcTTSetIamCmd             *cobra.Command
	dcTTAddIamCmd             *cobra.Command
	dcTTRemoveIamCmd          *cobra.Command
	dcTTFieldsCmd             *cobra.Command
	flagDCTTFieldConfigFile   string
	flagDCTTFieldUpdateMask   string
	flagDCTTFieldNewID        string
	flagDCTTFieldEnumRename   string
	dcTTFieldCreateCmd        *cobra.Command
	dcTTFieldDeleteCmd        *cobra.Command
	dcTTFieldUpdateCmd        *cobra.Command
	dcTTFieldRenameCmd        *cobra.Command
	dcTTFieldRenameEnumCmd    *cobra.Command
	flagDCTTFieldTemplate     string
	dcTagsCmd                 *cobra.Command
	flagDCTagLocation         string
	flagDCTagEntryGroup       string
	flagDCTagEntry            string
	flagDCTagFormat           string
	flagDCTagConfigFile       string
	flagDCTagUpdateMask       string
	flagDCTagPageSize         int64
	dcTagCreateCmd            *cobra.Command
	dcTagDeleteCmd            *cobra.Command
	dcTagListCmd              *cobra.Command
	dcTagUpdateCmd            *cobra.Command
	dcTaxCmd                  *cobra.Command
	flagDCTaxLocation         string
	flagDCTaxFormat           string
	flagDCTaxFilter           string
	flagDCTaxPageSize         int64
	flagDCTaxTaxIds           []string
	flagDCTaxSerialized       bool
	flagDCTaxImportFile       string
	flagDCTaxCrossRegional    string
	flagDCTaxIamMember        string
	flagDCTaxIamRole          string
	flagDCTaxIamCondExpr      string
	flagDCTaxIamCondT         string
	flagDCTaxIamCondD         string
	flagDCTaxIamAllCond       bool
	dcTaxDescribeCmd          *cobra.Command
	dcTaxListCmd              *cobra.Command
	dcTaxExportCmd            *cobra.Command
	dcTaxImportCmd            *cobra.Command
	dcTaxGetIamCmd            *cobra.Command
	dcTaxSetIamCmd            *cobra.Command
	dcTaxAddIamCmd            *cobra.Command
	dcTaxRemoveIamCmd         *cobra.Command
	dcPTCmd                   *cobra.Command
	flagDCPTTaxonomy          string
	flagDCPTConfigFile        string
	flagDCPTUpdateMask        string
	flagDCPTPageSize          int64
	flagDCPTIamMember         string
	flagDCPTIamRole           string
	flagDCPTIamCondExpr       string
	flagDCPTIamCondT          string
	flagDCPTIamCondD          string
	flagDCPTIamAllCond        bool
	dcPTCreateCmd             *cobra.Command
	dcPTDeleteCmd             *cobra.Command
	dcPTDescribeCmd           *cobra.Command
	dcPTListCmd               *cobra.Command
	dcPTUpdateCmd             *cobra.Command
	dcPTGetIamCmd             *cobra.Command
	dcPTSetIamCmd             *cobra.Command
	dcPTAddIamCmd             *cobra.Command
	dcPTRemoveIamCmd          *cobra.Command
}

// newDataCatalogCmd returns the "dataCatalog" command group for inv.
func newDataCatalogCmd(inv *invocation) *cobra.Command {
	return newDataCatalogCmds(inv).dataCatalogCmd
}

// newDataCatalogCmds builds the "dataCatalog" command group for inv and returns its state.
func newDataCatalogCmds(inv *invocation) *dataCatalogCmds {
	s := &dataCatalogCmds{inv: inv}
	s.dataCatalogCmd = &cobra.Command{Use: "data-catalog", Short: "Manage Data Catalog"}
	s.dcEntriesCmd = &cobra.Command{Use: "entries", Short: "Manage Data Catalog entries"}
	s.dcEntriesCreateCmd = &cobra.Command{
		Use: "create ENTRY", Short: "Create a Data Catalog entry",
		Args: cobra.ExactArgs(1), RunE: s.runDCEntryCreate,
	}
	s.dcEntriesDeleteCmd = &cobra.Command{
		Use: "delete ENTRY", Short: "Delete a Data Catalog entry",
		Args: cobra.ExactArgs(1), RunE: s.runDCEntryDelete,
	}
	s.dcEntriesDescribeCmd = &cobra.Command{
		Use: "describe ENTRY", Short: "Describe a Data Catalog entry",
		Args: cobra.ExactArgs(1), RunE: s.runDCEntryDescribe,
	}
	s.dcEntriesListCmd = &cobra.Command{
		Use: "list", Short: "List Data Catalog entries in an entry-group",
		Args: cobra.NoArgs, RunE: s.runDCEntryList,
	}
	s.dcEntriesUpdateCmd = &cobra.Command{
		Use: "update ENTRY", Short: "Update a Data Catalog entry",
		Args: cobra.ExactArgs(1), RunE: s.runDCEntryUpdate,
	}
	s.dcEntriesLookupCmd = &cobra.Command{
		Use: "lookup", Short: "Look up a Data Catalog entry by resource / SQL / fully-qualified name",
		Args: cobra.NoArgs, RunE: s.runDCEntryLookup,
	}
	s.dcEntriesStarCmd = &cobra.Command{
		Use: "star ENTRY", Short: "Star (bookmark) an entry",
		Args: cobra.ExactArgs(1), RunE: s.runDCEntryStar,
	}
	s.dcEntriesUnstarCmd = &cobra.Command{
		Use: "unstar ENTRY", Short: "Unstar an entry",
		Args: cobra.ExactArgs(1), RunE: s.runDCEntryUnstar,
	}
	s.dcEntryGroupsCmd = &cobra.Command{Use: "entry-groups", Short: "Manage Data Catalog entry groups"}
	s.dcEGCreateCmd = &cobra.Command{
		Use: "create ENTRY_GROUP", Short: "Create an entry group",
		Args: cobra.ExactArgs(1), RunE: s.runDCEGCreate,
	}
	s.dcEGDeleteCmd = &cobra.Command{
		Use: "delete ENTRY_GROUP", Short: "Delete an entry group",
		Args: cobra.ExactArgs(1), RunE: s.runDCEGDelete,
	}
	s.dcEGDescribeCmd = &cobra.Command{
		Use: "describe ENTRY_GROUP", Short: "Describe an entry group",
		Args: cobra.ExactArgs(1), RunE: s.runDCEGDescribe,
	}
	s.dcEGListCmd = &cobra.Command{
		Use: "list", Short: "List entry groups",
		Args: cobra.NoArgs, RunE: s.runDCEGList,
	}
	s.dcEGUpdateCmd = &cobra.Command{
		Use: "update ENTRY_GROUP", Short: "Update an entry group",
		Args: cobra.ExactArgs(1), RunE: s.runDCEGUpdate,
	}
	s.dcEGGetIamCmd = &cobra.Command{
		Use: "get-iam-policy ENTRY_GROUP", Short: "Get the IAM policy for an entry group",
		Args: cobra.ExactArgs(1), RunE: s.runDCEGGetIam,
	}
	s.dcEGSetIamCmd = &cobra.Command{
		Use: "set-iam-policy ENTRY_GROUP POLICY_FILE", Short: "Set the IAM policy for an entry group",
		Args: cobra.ExactArgs(2), RunE: s.runDCEGSetIam,
	}
	s.dcEGAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding ENTRY_GROUP", Short: "Add an IAM policy binding to an entry group",
		Args: cobra.ExactArgs(1), RunE: s.runDCEGAddIam,
	}
	s.dcEGRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding ENTRY_GROUP", Short: "Remove an IAM policy binding from an entry group",
		Args: cobra.ExactArgs(1), RunE: s.runDCEGRemoveIam,
	}
	s.dataCatalogSearchCmd = &cobra.Command{
		Use:   "search QUERY",
		Short: "Search Data Catalog",
		Args:  cobra.MaximumNArgs(1),
		RunE:  s.runDCSearch,
	}
	s.dcTTCmd = &cobra.Command{Use: "tag-templates", Short: "Manage Data Catalog tag templates"}
	s.dcTTCreateCmd = &cobra.Command{
		Use: "create TAG_TEMPLATE", Short: "Create a tag template",
		Args: cobra.ExactArgs(1), RunE: s.runDCTTCreate,
	}
	s.dcTTDeleteCmd = &cobra.Command{
		Use: "delete TAG_TEMPLATE", Short: "Delete a tag template",
		Args: cobra.ExactArgs(1), RunE: s.runDCTTDelete,
	}
	s.dcTTDescribeCmd = &cobra.Command{
		Use: "describe TAG_TEMPLATE", Short: "Describe a tag template",
		Args: cobra.ExactArgs(1), RunE: s.runDCTTDescribe,
	}
	s.dcTTUpdateCmd = &cobra.Command{
		Use: "update TAG_TEMPLATE", Short: "Update a tag template",
		Args: cobra.ExactArgs(1), RunE: s.runDCTTUpdate,
	}
	s.dcTTGetIamCmd = &cobra.Command{
		Use: "get-iam-policy TAG_TEMPLATE", Short: "Get the IAM policy for a tag template",
		Args: cobra.ExactArgs(1), RunE: s.runDCTTGetIam,
	}
	s.dcTTSetIamCmd = &cobra.Command{
		Use: "set-iam-policy TAG_TEMPLATE POLICY_FILE", Short: "Set the IAM policy for a tag template",
		Args: cobra.ExactArgs(2), RunE: s.runDCTTSetIam,
	}
	s.dcTTAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding TAG_TEMPLATE", Short: "Add an IAM policy binding to a tag template",
		Args: cobra.ExactArgs(1), RunE: s.runDCTTAddIam,
	}
	s.dcTTRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding TAG_TEMPLATE", Short: "Remove an IAM policy binding from a tag template",
		Args: cobra.ExactArgs(1), RunE: s.runDCTTRemoveIam,
	}
	s.dcTTFieldsCmd = &cobra.Command{Use: "fields", Short: "Manage tag template fields"}
	s.dcTTFieldCreateCmd = &cobra.Command{
		Use: "create FIELD", Short: "Create a tag template field",
		Args: cobra.ExactArgs(1), RunE: s.runDCTTFieldCreate,
	}
	s.dcTTFieldDeleteCmd = &cobra.Command{
		Use: "delete FIELD", Short: "Delete a tag template field",
		Args: cobra.ExactArgs(1), RunE: s.runDCTTFieldDelete,
	}
	s.dcTTFieldUpdateCmd = &cobra.Command{
		Use: "update FIELD", Short: "Update a tag template field",
		Args: cobra.ExactArgs(1), RunE: s.runDCTTFieldUpdate,
	}
	s.dcTTFieldRenameCmd = &cobra.Command{
		Use: "rename FIELD", Short: "Rename a tag template field ID",
		Args: cobra.ExactArgs(1), RunE: s.runDCTTFieldRename,
	}
	s.dcTTFieldRenameEnumCmd = &cobra.Command{
		Use: "rename-enum-value FIELD ENUM_VALUE", Short: "Rename an enum value on a tag template field",
		Args: cobra.ExactArgs(2), RunE: s.runDCTTFieldRenameEnum,
	}
	s.dcTagsCmd = &cobra.Command{Use: "tags", Short: "Manage Data Catalog tags"}
	s.dcTagCreateCmd = &cobra.Command{
		Use: "create", Short: "Create a tag",
		Args: cobra.NoArgs, RunE: s.runDCTagCreate,
	}
	s.dcTagDeleteCmd = &cobra.Command{
		Use: "delete TAG", Short: "Delete a tag",
		Args: cobra.ExactArgs(1), RunE: s.runDCTagDelete,
	}
	s.dcTagListCmd = &cobra.Command{
		Use: "list", Short: "List tags on an entry or entry group",
		Args: cobra.NoArgs, RunE: s.runDCTagList,
	}
	s.dcTagUpdateCmd = &cobra.Command{
		Use: "update TAG", Short: "Update a tag",
		Args: cobra.ExactArgs(1), RunE: s.runDCTagUpdate,
	}
	s.dcTaxCmd = &cobra.Command{Use: "taxonomies", Short: "Manage Data Catalog policy taxonomies"}
	s.dcTaxDescribeCmd = &cobra.Command{
		Use: "describe TAXONOMY", Short: "Describe a taxonomy",
		Args: cobra.ExactArgs(1), RunE: s.runDCTaxDescribe,
	}
	s.dcTaxListCmd = &cobra.Command{
		Use: "list", Short: "List taxonomies",
		Args: cobra.NoArgs, RunE: s.runDCTaxList,
	}
	s.dcTaxExportCmd = &cobra.Command{
		Use: "export", Short: "Export taxonomies (returns serialized JSON)",
		Args: cobra.NoArgs, RunE: s.runDCTaxExport,
	}
	s.dcTaxImportCmd = &cobra.Command{
		Use: "import", Short: "Import taxonomies (from a serialized JSON file or a cross-regional source)",
		Args: cobra.NoArgs, RunE: s.runDCTaxImport,
	}
	s.dcTaxGetIamCmd = &cobra.Command{
		Use: "get-iam-policy TAXONOMY", Short: "Get the IAM policy for a taxonomy",
		Args: cobra.ExactArgs(1), RunE: s.runDCTaxGetIam,
	}
	s.dcTaxSetIamCmd = &cobra.Command{
		Use: "set-iam-policy TAXONOMY POLICY_FILE", Short: "Set the IAM policy for a taxonomy",
		Args: cobra.ExactArgs(2), RunE: s.runDCTaxSetIam,
	}
	s.dcTaxAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding TAXONOMY", Short: "Add an IAM policy binding to a taxonomy",
		Args: cobra.ExactArgs(1), RunE: s.runDCTaxAddIam,
	}
	s.dcTaxRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding TAXONOMY", Short: "Remove an IAM policy binding from a taxonomy",
		Args: cobra.ExactArgs(1), RunE: s.runDCTaxRemoveIam,
	}
	s.dcPTCmd = &cobra.Command{Use: "policy-tags", Short: "Manage taxonomy policy tags"}
	s.dcPTCreateCmd = &cobra.Command{
		Use: "create POLICY_TAG", Short: "Create a policy tag",
		Args: cobra.ExactArgs(1), RunE: s.runDCPTCreate,
	}
	s.dcPTDeleteCmd = &cobra.Command{
		Use: "delete POLICY_TAG", Short: "Delete a policy tag",
		Args: cobra.ExactArgs(1), RunE: s.runDCPTDelete,
	}
	s.dcPTDescribeCmd = &cobra.Command{
		Use: "describe POLICY_TAG", Short: "Describe a policy tag",
		Args: cobra.ExactArgs(1), RunE: s.runDCPTDescribe,
	}
	s.dcPTListCmd = &cobra.Command{
		Use: "list", Short: "List policy tags",
		Args: cobra.NoArgs, RunE: s.runDCPTList,
	}
	s.dcPTUpdateCmd = &cobra.Command{
		Use: "update POLICY_TAG", Short: "Update a policy tag",
		Args: cobra.ExactArgs(1), RunE: s.runDCPTUpdate,
	}
	s.dcPTGetIamCmd = &cobra.Command{
		Use: "get-iam-policy POLICY_TAG", Short: "Get the IAM policy for a policy tag",
		Args: cobra.ExactArgs(1), RunE: s.runDCPTGetIam,
	}
	s.dcPTSetIamCmd = &cobra.Command{
		Use: "set-iam-policy POLICY_TAG POLICY_FILE", Short: "Set the IAM policy for a policy tag",
		Args: cobra.ExactArgs(2), RunE: s.runDCPTSetIam,
	}
	s.dcPTAddIamCmd = &cobra.Command{
		Use: "add-iam-policy-binding POLICY_TAG", Short: "Add an IAM policy binding to a policy tag",
		Args: cobra.ExactArgs(1), RunE: s.runDCPTAddIam,
	}
	s.dcPTRemoveIamCmd = &cobra.Command{
		Use: "remove-iam-policy-binding POLICY_TAG", Short: "Remove an IAM policy binding from a policy tag",
		Args: cobra.ExactArgs(1), RunE: s.runDCPTRemoveIam,
	}
	s.initDataCatalog()
	s.initDataCatalogEntries()
	s.initDataCatalogEntryGroups()
	s.initDataCatalogSearch()
	s.initDataCatalogTagTemplates()
	s.initDataCatalogTagTemplates5()
	s.initDataCatalogTags()
	s.initDataCatalogTaxonomies()
	s.initDataCatalogTaxonomies8()
	return s
}
