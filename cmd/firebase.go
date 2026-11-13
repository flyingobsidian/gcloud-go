package cmd

import "github.com/spf13/cobra"

// --- gcloud firebase (#340) ---

func (s *firebaseCmds) initFirebase() {
	s.firebaseCmd.AddCommand(s.firebaseTestCmd)

}

// firebaseCmds holds the commands and flag values of one invocation's "firebase"
// command group, so invocations don't share state (#2070).
type firebaseCmds struct {
	inv *invocation

	firebaseCmd                            *cobra.Command
	firebaseTestCmd                        *cobra.Command
	firebaseTestAndroidCmd                 *cobra.Command
	flagFtaFormat                          string
	flagFtaConfigFile                      string
	flagFtaProject                         string
	firebaseTestAndroidModelsCmd           *cobra.Command
	firebaseTestAndroidModelsListCmd       *cobra.Command
	firebaseTestAndroidVersionsCmd         *cobra.Command
	firebaseTestAndroidVersionsListCmd     *cobra.Command
	firebaseTestAndroidLocalesCmd          *cobra.Command
	firebaseTestAndroidLocalesListCmd      *cobra.Command
	firebaseTestAndroidRunCmd              *cobra.Command
	firebaseTestIosCmd                     *cobra.Command
	flagFtiFormat                          string
	flagFtiConfigFile                      string
	flagFtiProject                         string
	firebaseTestIosModelsCmd               *cobra.Command
	firebaseTestIosModelsListCmd           *cobra.Command
	firebaseTestIosVersionsCmd             *cobra.Command
	firebaseTestIosVersionsListCmd         *cobra.Command
	firebaseTestIosLocalesCmd              *cobra.Command
	firebaseTestIosLocalesListCmd          *cobra.Command
	firebaseTestIosRunCmd                  *cobra.Command
	firebaseTestNetworkProfilesCmd         *cobra.Command
	flagFtnpFormat                         string
	flagFtnpID                             string
	firebaseTestNetworkProfilesListCmd     *cobra.Command
	firebaseTestNetworkProfilesDescribeCmd *cobra.Command
}

// newFirebaseCmd returns the "firebase" command group for inv.
func newFirebaseCmd(inv *invocation) *cobra.Command {
	return newFirebaseCmds(inv).firebaseCmd
}

// newFirebaseCmds builds the "firebase" command group for inv and returns its state.
func newFirebaseCmds(inv *invocation) *firebaseCmds {
	s := &firebaseCmds{inv: inv}
	s.firebaseCmd = &cobra.Command{Use: "firebase", Short: "Work with Google Firebase"}
	s.firebaseTestCmd = &cobra.Command{Use: "test", Short: "Interact with Firebase Test Lab"}
	s.firebaseTestAndroidCmd = &cobra.Command{Use: "android", Short: "Android application testing"}
	s.firebaseTestAndroidModelsCmd = &cobra.Command{
		Use: "models", Short: "Manage supported Android device models",
	}
	s.firebaseTestAndroidModelsListCmd = &cobra.Command{
		Use: "list", Short: "List supported Android device models",
		Args: cobra.NoArgs, RunE: s.runFtaModelsList,
	}
	s.firebaseTestAndroidVersionsCmd = &cobra.Command{
		Use: "versions", Short: "Manage supported Android OS versions",
	}
	s.firebaseTestAndroidVersionsListCmd = &cobra.Command{
		Use: "list", Short: "List supported Android OS versions",
		Args: cobra.NoArgs, RunE: s.runFtaVersionsList,
	}
	s.firebaseTestAndroidLocalesCmd = &cobra.Command{
		Use: "locales", Short: "Manage supported Android locales",
	}
	s.firebaseTestAndroidLocalesListCmd = &cobra.Command{
		Use: "locales", Short: "List supported Android locales",
		Args: cobra.NoArgs, RunE: s.runFtaLocalesList,
	}
	s.firebaseTestAndroidRunCmd = &cobra.Command{
		Use: "run", Short: "Submit an Android test matrix (loads TestMatrix body from --config-file)",
		Args: cobra.NoArgs, RunE: s.runFtaRun,
	}
	s.firebaseTestIosCmd = &cobra.Command{Use: "ios", Short: "iOS application testing"}
	s.firebaseTestIosModelsCmd = &cobra.Command{Use: "models", Short: "Manage supported iOS device models"}
	s.firebaseTestIosModelsListCmd = &cobra.Command{
		Use: "list", Short: "List supported iOS device models",
		Args: cobra.NoArgs, RunE: s.runFtiModelsList,
	}
	s.firebaseTestIosVersionsCmd = &cobra.Command{Use: "versions", Short: "Manage supported iOS versions"}
	s.firebaseTestIosVersionsListCmd = &cobra.Command{
		Use: "list", Short: "List supported iOS versions",
		Args: cobra.NoArgs, RunE: s.runFtiVersionsList,
	}
	s.firebaseTestIosLocalesCmd = &cobra.Command{Use: "locales", Short: "Manage supported iOS locales"}
	s.firebaseTestIosLocalesListCmd = &cobra.Command{
		Use: "list", Short: "List supported iOS locales",
		Args: cobra.NoArgs, RunE: s.runFtiLocalesList,
	}
	s.firebaseTestIosRunCmd = &cobra.Command{
		Use: "run", Short: "Submit an iOS test matrix (loads TestMatrix body from --config-file)",
		Args: cobra.NoArgs, RunE: s.runFtiRun,
	}
	s.firebaseTestNetworkProfilesCmd = &cobra.Command{Use: "network-profiles", Short: "List network profiles available to Firebase Test Lab"}
	s.firebaseTestNetworkProfilesListCmd = &cobra.Command{
		Use: "list", Short: "List network profiles",
		Args: cobra.NoArgs, RunE: s.runFtnpList,
	}
	s.firebaseTestNetworkProfilesDescribeCmd = &cobra.Command{
		Use: "describe PROFILE_ID", Short: "Describe a network profile",
		Args: cobra.ExactArgs(1), RunE: s.runFtnpDescribe,
	}
	s.initFirebase()
	s.initFirebaseTestAndroidCmd()
	s.initFirebaseTestIosCmd()
	s.initFirebaseTestNetworkProfiles()
	return s
}
