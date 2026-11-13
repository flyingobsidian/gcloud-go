package cli

import "github.com/spf13/cobra"

// --- gcloud edge-cache (#333) ---
//
// Media CDN Edge Cache is served by the networkservices.googleapis.com API.
// Only the IAM-only surfaces are generated in google.golang.org/api; the CRUD
// endpoints (keysets, origins, services) are reached via the shared restClient
// against edgeCacheRest so subcommand files can call them uniformly.

func edgeCacheRest() *restClient {
	return newRESTClient("https://networkservices.googleapis.com/v1")
}

func (s *edgeCacheCmds) initEdgeCache() {

}

// edgeCacheCmds holds the commands and flag values of one invocation's "edgeCache"
// command group, so invocations don't share state (#2070).
type edgeCacheCmds struct {
	inv *invocation

	edgeCacheCmd                      *cobra.Command
	edgeCacheKeysetsCmd               *cobra.Command
	flagEdgeCacheKeysetsLocation      string
	flagEdgeCacheKeysetsFormat        string
	flagEdgeCacheKeysetsConfigFile    string
	flagEdgeCacheKeysetsUpdateMask    string
	flagEdgeCacheKeysetsDestination   string
	flagEdgeCacheKeysetsSource        string
	flagEdgeCacheKeysetsPageSize      int64
	edgeCacheKeysetsCreateCmd         *cobra.Command
	edgeCacheKeysetsDeleteCmd         *cobra.Command
	edgeCacheKeysetsDescribeCmd       *cobra.Command
	edgeCacheKeysetsExportCmd         *cobra.Command
	edgeCacheKeysetsImportCmd         *cobra.Command
	edgeCacheKeysetsListCmd           *cobra.Command
	edgeCacheKeysetsUpdateCmd         *cobra.Command
	edgeCacheOperationsCmd            *cobra.Command
	flagEdgeCacheOperationsLocation   string
	flagEdgeCacheOperationsFormat     string
	flagEdgeCacheOperationsPageSize   int64
	edgeCacheOperationsDescribeCmd    *cobra.Command
	edgeCacheOperationsListCmd        *cobra.Command
	edgeCacheOriginsCmd               *cobra.Command
	flagEdgeCacheOriginsLocation      string
	flagEdgeCacheOriginsFormat        string
	flagEdgeCacheOriginsConfigFile    string
	flagEdgeCacheOriginsUpdateMask    string
	flagEdgeCacheOriginsDestination   string
	flagEdgeCacheOriginsSource        string
	flagEdgeCacheOriginsPageSize      int64
	edgeCacheOriginsCreateCmd         *cobra.Command
	edgeCacheOriginsDeleteCmd         *cobra.Command
	edgeCacheOriginsDescribeCmd       *cobra.Command
	edgeCacheOriginsExportCmd         *cobra.Command
	edgeCacheOriginsImportCmd         *cobra.Command
	edgeCacheOriginsListCmd           *cobra.Command
	edgeCacheOriginsUpdateCmd         *cobra.Command
	edgeCacheServicesCmd              *cobra.Command
	flagEdgeCacheServicesLocation     string
	flagEdgeCacheServicesFormat       string
	flagEdgeCacheServicesConfigFile   string
	flagEdgeCacheServicesUpdateMask   string
	flagEdgeCacheServicesDestination  string
	flagEdgeCacheServicesSource       string
	flagEdgeCacheServicesInvalidateFn string
	flagEdgeCacheServicesPageSize     int64
	edgeCacheServicesDeleteCmd        *cobra.Command
	edgeCacheServicesDescribeCmd      *cobra.Command
	edgeCacheServicesExportCmd        *cobra.Command
	edgeCacheServicesImportCmd        *cobra.Command
	edgeCacheServicesInvalidateCmd    *cobra.Command
	edgeCacheServicesListCmd          *cobra.Command
	edgeCacheServicesUpdateCmd        *cobra.Command
}

// newEdgeCacheCmd returns the "edgeCache" command group for inv.
func newEdgeCacheCmd(inv *invocation) *cobra.Command {
	return newEdgeCacheCmds(inv).edgeCacheCmd
}

// newEdgeCacheCmds builds the "edgeCache" command group for inv and returns its state.
func newEdgeCacheCmds(inv *invocation) *edgeCacheCmds {
	s := &edgeCacheCmds{inv: inv}
	s.edgeCacheCmd = &cobra.Command{Use: "edge-cache", Short: "Manage Google Cloud Media CDN Edge Cache"}
	s.edgeCacheKeysetsCmd = &cobra.Command{Use: "keysets", Short: "Manage EdgeCacheKeyset resources"}
	s.edgeCacheKeysetsCreateCmd = &cobra.Command{
		Use: "create KEYSET", Short: "Create an EdgeCacheKeyset",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheKeysetsCreate,
	}
	s.edgeCacheKeysetsDeleteCmd = &cobra.Command{
		Use: "delete KEYSET", Short: "Delete an EdgeCacheKeyset",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheKeysetsDelete,
	}
	s.edgeCacheKeysetsDescribeCmd = &cobra.Command{
		Use: "describe KEYSET", Short: "Describe an EdgeCacheKeyset",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheKeysetsDescribe,
	}
	s.edgeCacheKeysetsExportCmd = &cobra.Command{
		Use: "export KEYSET", Short: "Export an EdgeCacheKeyset (as YAML or JSON) to a file or stdout",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheKeysetsExport,
	}
	s.edgeCacheKeysetsImportCmd = &cobra.Command{
		Use: "import KEYSET", Short: "Import (create or update) an EdgeCacheKeyset from a file",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheKeysetsImport,
	}
	s.edgeCacheKeysetsListCmd = &cobra.Command{
		Use: "list", Short: "List EdgeCacheKeyset resources",
		Args: cobra.NoArgs, RunE: s.runEdgeCacheKeysetsList,
	}
	s.edgeCacheKeysetsUpdateCmd = &cobra.Command{
		Use: "update KEYSET", Short: "Update an EdgeCacheKeyset",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheKeysetsUpdate,
	}
	s.edgeCacheOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Media CDN Edge Cache long-running operations"}
	s.edgeCacheOperationsDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe an Edge Cache long-running operation",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheOperationsDescribe,
	}
	s.edgeCacheOperationsListCmd = &cobra.Command{
		Use: "list", Short: "List Edge Cache long-running operations",
		Args: cobra.NoArgs, RunE: s.runEdgeCacheOperationsList,
	}
	s.edgeCacheOriginsCmd = &cobra.Command{Use: "origins", Short: "Manage EdgeCacheOrigin resources"}
	s.edgeCacheOriginsCreateCmd = &cobra.Command{
		Use: "create ORIGIN", Short: "Create an EdgeCacheOrigin",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheOriginsCreate,
	}
	s.edgeCacheOriginsDeleteCmd = &cobra.Command{
		Use: "delete ORIGIN", Short: "Delete an EdgeCacheOrigin",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheOriginsDelete,
	}
	s.edgeCacheOriginsDescribeCmd = &cobra.Command{
		Use: "describe ORIGIN", Short: "Describe an EdgeCacheOrigin",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheOriginsDescribe,
	}
	s.edgeCacheOriginsExportCmd = &cobra.Command{
		Use: "export ORIGIN", Short: "Export an EdgeCacheOrigin (as YAML or JSON) to a file or stdout",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheOriginsExport,
	}
	s.edgeCacheOriginsImportCmd = &cobra.Command{
		Use: "import ORIGIN", Short: "Import (create or update) an EdgeCacheOrigin from a file",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheOriginsImport,
	}
	s.edgeCacheOriginsListCmd = &cobra.Command{
		Use: "list", Short: "List EdgeCacheOrigin resources",
		Args: cobra.NoArgs, RunE: s.runEdgeCacheOriginsList,
	}
	s.edgeCacheOriginsUpdateCmd = &cobra.Command{
		Use: "update ORIGIN", Short: "Update an EdgeCacheOrigin",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheOriginsUpdate,
	}
	s.edgeCacheServicesCmd = &cobra.Command{Use: "services", Short: "Manage EdgeCacheService resources"}
	s.edgeCacheServicesDeleteCmd = &cobra.Command{
		Use: "delete SERVICE", Short: "Delete an EdgeCacheService",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheServicesDelete,
	}
	s.edgeCacheServicesDescribeCmd = &cobra.Command{
		Use: "describe SERVICE", Short: "Describe an EdgeCacheService",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheServicesDescribe,
	}
	s.edgeCacheServicesExportCmd = &cobra.Command{
		Use: "export SERVICE", Short: "Export an EdgeCacheService (as YAML or JSON) to a file or stdout",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheServicesExport,
	}
	s.edgeCacheServicesImportCmd = &cobra.Command{
		Use: "import SERVICE", Short: "Import (create or update) an EdgeCacheService from a file",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheServicesImport,
	}
	s.edgeCacheServicesInvalidateCmd = &cobra.Command{
		Use: "invalidate-cache SERVICE", Short: "Invalidate cache for an EdgeCacheService",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheServicesInvalidate,
	}
	s.edgeCacheServicesListCmd = &cobra.Command{
		Use: "list", Short: "List EdgeCacheService resources",
		Args: cobra.NoArgs, RunE: s.runEdgeCacheServicesList,
	}
	s.edgeCacheServicesUpdateCmd = &cobra.Command{
		Use: "update SERVICE", Short: "Update an EdgeCacheService",
		Args: cobra.ExactArgs(1), RunE: s.runEdgeCacheServicesUpdate,
	}
	s.initEdgeCache()
	s.initEdgeCacheKeysets()
	s.initEdgeCacheOperations()
	s.initEdgeCacheOrigins()
	s.initEdgeCacheServices()
	return s
}
