package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud healthcare (#344) ---

func (s *healthcareCmds) initHealthcare() {
	// consent-stores (#1220), dicom-stores (#1222), fhir-stores (#1223),
	// hl7v2-stores (#1224), datasets (#1221) and operations (#1225) are
	// implemented in dedicated files. Specialty verbs (deidentify/export/
	// import/rollback/search/etc.) will land under separate issues.

}

// hcLocationParent returns projects/PROJECT/locations/LOCATION.
func (s *healthcareCmds) hcLocationParent(location string) (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	if location == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location), nil
}

// hcDatasetName returns projects/.../datasets/DATASET.
func (s *healthcareCmds) hcDatasetName(location, dataset string) (string, error) {
	if strings.HasPrefix(dataset, "projects/") {
		return dataset, nil
	}
	parent, err := s.hcLocationParent(location)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/datasets/%s", parent, dataset), nil
}

// healthcareCmds holds the commands and flag values of one invocation's "healthcare"
// command group, so invocations don't share state (#2070).
type healthcareCmds struct {
	inv *invocation

	healthcareCmd              *cobra.Command
	healthcareConsentStoresCmd *cobra.Command
	flagHcCsLocation           string
	flagHcCsDataset            string
	flagHcCsFormat             string
	flagHcCsConfigFile         string
	flagHcCsUpdateMask         string
	flagHcCsPageSize           int64
	hcCsCreateCmd              *cobra.Command
	hcCsDeleteCmd              *cobra.Command
	hcCsDescribeCmd            *cobra.Command
	hcCsListCmd                *cobra.Command
	hcCsUpdateCmd              *cobra.Command
	hcCsGetIamCmd              *cobra.Command
	hcCsSetIamCmd              *cobra.Command
	healthcareDatasetsCmd      *cobra.Command
	flagHcDsLocation           string
	flagHcDsFormat             string
	flagHcDsConfigFile         string
	flagHcDsUpdateMask         string
	flagHcDsPageSize           int64
	flagHcDsDestDataset        string
	healthcareDsCreateCmd      *cobra.Command
	healthcareDsDeleteCmd      *cobra.Command
	healthcareDsDescribeCmd    *cobra.Command
	healthcareDsListCmd        *cobra.Command
	healthcareDsUpdateCmd      *cobra.Command
	healthcareDsGetIamCmd      *cobra.Command
	healthcareDsSetIamCmd      *cobra.Command
	healthcareDsDeidentifyCmd  *cobra.Command
	healthcareDicomStoresCmd   *cobra.Command
	flagHcDsxLocation          string
	flagHcDsxDataset           string
	flagHcDsxFormat            string
	flagHcDsxConfigFile        string
	flagHcDsxUpdateMask        string
	flagHcDsxPageSize          int64
	hcDsxCreateCmd             *cobra.Command
	hcDsxDeleteCmd             *cobra.Command
	hcDsxDescribeCmd           *cobra.Command
	hcDsxListCmd               *cobra.Command
	hcDsxUpdateCmd             *cobra.Command
	hcDsxGetIamCmd             *cobra.Command
	hcDsxSetIamCmd             *cobra.Command
	healthcareFhirStoresCmd    *cobra.Command
	flagHcFsLocation           string
	flagHcFsDataset            string
	flagHcFsFormat             string
	flagHcFsConfigFile         string
	flagHcFsUpdateMask         string
	flagHcFsPageSize           int64
	hcFsCreateCmd              *cobra.Command
	hcFsDeleteCmd              *cobra.Command
	hcFsDescribeCmd            *cobra.Command
	hcFsListCmd                *cobra.Command
	hcFsUpdateCmd              *cobra.Command
	hcFsGetIamCmd              *cobra.Command
	hcFsSetIamCmd              *cobra.Command
	healthcareHl7v2StoresCmd   *cobra.Command
	flagHcH7Location           string
	flagHcH7Dataset            string
	flagHcH7Format             string
	flagHcH7ConfigFile         string
	flagHcH7UpdateMask         string
	flagHcH7PageSize           int64
	hcH7CreateCmd              *cobra.Command
	hcH7DeleteCmd              *cobra.Command
	hcH7DescribeCmd            *cobra.Command
	hcH7ListCmd                *cobra.Command
	hcH7UpdateCmd              *cobra.Command
	hcH7GetIamCmd              *cobra.Command
	hcH7SetIamCmd              *cobra.Command
	healthcareOperationsCmd    *cobra.Command
	flagHcOpLocation           string
	flagHcOpDataset            string
	flagHcOpFormat             string
	flagHcOpFilter             string
	flagHcOpPageSize           int64
	healthcareOpCancelCmd      *cobra.Command
	healthcareOpDescribeCmd    *cobra.Command
	healthcareOpListCmd        *cobra.Command
}

// newHealthcareCmd returns the "healthcare" command group for inv.
func newHealthcareCmd(inv *invocation) *cobra.Command {
	return newHealthcareCmds(inv).healthcareCmd
}

// newHealthcareCmds builds the "healthcare" command group for inv and returns its state.
func newHealthcareCmds(inv *invocation) *healthcareCmds {
	s := &healthcareCmds{inv: inv}
	s.healthcareCmd = &cobra.Command{Use: "healthcare", Short: "Manage Cloud Healthcare API"}
	s.healthcareConsentStoresCmd = &cobra.Command{Use: "consent-stores", Short: "Manage Healthcare consent stores"}
	s.hcCsCreateCmd = &cobra.Command{
		Use: "create CONSENT_STORE", Short: "Create a Healthcare consent store",
		Args: cobra.ExactArgs(1), RunE: s.runHcCsCreate,
	}
	s.hcCsDeleteCmd = &cobra.Command{
		Use: "delete CONSENT_STORE", Short: "Delete a Healthcare consent store",
		Args: cobra.ExactArgs(1), RunE: s.runHcCsDelete,
	}
	s.hcCsDescribeCmd = &cobra.Command{
		Use: "describe CONSENT_STORE", Short: "Describe a Healthcare consent store",
		Args: cobra.ExactArgs(1), RunE: s.runHcCsDescribe,
	}
	s.hcCsListCmd = &cobra.Command{
		Use: "list", Short: "List Healthcare consent stores in a dataset",
		Args: cobra.NoArgs, RunE: s.runHcCsList,
	}
	s.hcCsUpdateCmd = &cobra.Command{
		Use: "update CONSENT_STORE", Short: "Update a Healthcare consent store",
		Args: cobra.ExactArgs(1), RunE: s.runHcCsUpdate,
	}
	s.hcCsGetIamCmd = &cobra.Command{
		Use: "get-iam-policy CONSENT_STORE", Short: "Get the IAM policy for a Healthcare consent store",
		Args: cobra.ExactArgs(1), RunE: s.runHcCsGetIam,
	}
	s.hcCsSetIamCmd = &cobra.Command{
		Use: "set-iam-policy CONSENT_STORE POLICY_FILE", Short: "Set the IAM policy for a Healthcare consent store",
		Args: cobra.ExactArgs(2), RunE: s.runHcCsSetIam,
	}
	s.healthcareDatasetsCmd = &cobra.Command{Use: "datasets", Short: "Manage Cloud Healthcare datasets"}
	s.healthcareDsCreateCmd = &cobra.Command{
		Use: "create DATASET", Short: "Create a healthcare dataset",
		Args: cobra.ExactArgs(1), RunE: s.runHcDsCreate,
	}
	s.healthcareDsDeleteCmd = &cobra.Command{
		Use: "delete DATASET", Short: "Delete a healthcare dataset",
		Args: cobra.ExactArgs(1), RunE: s.runHcDsDelete,
	}
	s.healthcareDsDescribeCmd = &cobra.Command{
		Use: "describe DATASET", Short: "Describe a healthcare dataset",
		Args: cobra.ExactArgs(1), RunE: s.runHcDsDescribe,
	}
	s.healthcareDsListCmd = &cobra.Command{
		Use: "list", Short: "List healthcare datasets in a location",
		Args: cobra.NoArgs, RunE: s.runHcDsList,
	}
	s.healthcareDsUpdateCmd = &cobra.Command{
		Use: "update DATASET", Short: "Update a healthcare dataset",
		Args: cobra.ExactArgs(1), RunE: s.runHcDsUpdate,
	}
	s.healthcareDsGetIamCmd = &cobra.Command{
		Use: "get-iam-policy DATASET", Short: "Get the IAM policy for a healthcare dataset",
		Args: cobra.ExactArgs(1), RunE: s.runHcDsGetIam,
	}
	s.healthcareDsSetIamCmd = &cobra.Command{
		Use: "set-iam-policy DATASET POLICY_FILE", Short: "Set the IAM policy for a healthcare dataset",
		Args: cobra.ExactArgs(2), RunE: s.runHcDsSetIam,
	}
	s.healthcareDsDeidentifyCmd = &cobra.Command{
		Use: "deidentify DATASET", Short: "De-identify a healthcare dataset into another dataset",
		Args: cobra.ExactArgs(1), RunE: s.runHcDsDeidentify,
	}
	s.healthcareDicomStoresCmd = &cobra.Command{Use: "dicom-stores", Short: "Manage Healthcare DICOM stores"}
	s.hcDsxCreateCmd = &cobra.Command{
		Use: "create DICOM_STORE", Short: "Create a Healthcare DICOM store",
		Args: cobra.ExactArgs(1), RunE: s.runHcDsxCreate,
	}
	s.hcDsxDeleteCmd = &cobra.Command{
		Use: "delete DICOM_STORE", Short: "Delete a Healthcare DICOM store",
		Args: cobra.ExactArgs(1), RunE: s.runHcDsxDelete,
	}
	s.hcDsxDescribeCmd = &cobra.Command{
		Use: "describe DICOM_STORE", Short: "Describe a Healthcare DICOM store",
		Args: cobra.ExactArgs(1), RunE: s.runHcDsxDescribe,
	}
	s.hcDsxListCmd = &cobra.Command{
		Use: "list", Short: "List Healthcare DICOM stores in a dataset",
		Args: cobra.NoArgs, RunE: s.runHcDsxList,
	}
	s.hcDsxUpdateCmd = &cobra.Command{
		Use: "update DICOM_STORE", Short: "Update a Healthcare DICOM store",
		Args: cobra.ExactArgs(1), RunE: s.runHcDsxUpdate,
	}
	s.hcDsxGetIamCmd = &cobra.Command{
		Use: "get-iam-policy DICOM_STORE", Short: "Get the IAM policy for a Healthcare DICOM store",
		Args: cobra.ExactArgs(1), RunE: s.runHcDsxGetIam,
	}
	s.hcDsxSetIamCmd = &cobra.Command{
		Use: "set-iam-policy DICOM_STORE POLICY_FILE", Short: "Set the IAM policy for a Healthcare DICOM store",
		Args: cobra.ExactArgs(2), RunE: s.runHcDsxSetIam,
	}
	s.healthcareFhirStoresCmd = &cobra.Command{Use: "fhir-stores", Short: "Manage Healthcare FHIR stores"}
	s.hcFsCreateCmd = &cobra.Command{
		Use: "create FHIR_STORE", Short: "Create a Healthcare FHIR store",
		Args: cobra.ExactArgs(1), RunE: s.runHcFsCreate,
	}
	s.hcFsDeleteCmd = &cobra.Command{
		Use: "delete FHIR_STORE", Short: "Delete a Healthcare FHIR store",
		Args: cobra.ExactArgs(1), RunE: s.runHcFsDelete,
	}
	s.hcFsDescribeCmd = &cobra.Command{
		Use: "describe FHIR_STORE", Short: "Describe a Healthcare FHIR store",
		Args: cobra.ExactArgs(1), RunE: s.runHcFsDescribe,
	}
	s.hcFsListCmd = &cobra.Command{
		Use: "list", Short: "List Healthcare FHIR stores in a dataset",
		Args: cobra.NoArgs, RunE: s.runHcFsList,
	}
	s.hcFsUpdateCmd = &cobra.Command{
		Use: "update FHIR_STORE", Short: "Update a Healthcare FHIR store",
		Args: cobra.ExactArgs(1), RunE: s.runHcFsUpdate,
	}
	s.hcFsGetIamCmd = &cobra.Command{
		Use: "get-iam-policy FHIR_STORE", Short: "Get the IAM policy for a Healthcare FHIR store",
		Args: cobra.ExactArgs(1), RunE: s.runHcFsGetIam,
	}
	s.hcFsSetIamCmd = &cobra.Command{
		Use: "set-iam-policy FHIR_STORE POLICY_FILE", Short: "Set the IAM policy for a Healthcare FHIR store",
		Args: cobra.ExactArgs(2), RunE: s.runHcFsSetIam,
	}
	s.healthcareHl7v2StoresCmd = &cobra.Command{Use: "hl7v2-stores", Short: "Manage Healthcare HL7v2 stores"}
	s.hcH7CreateCmd = &cobra.Command{
		Use: "create HL7V2_STORE", Short: "Create a Healthcare HL7v2 store",
		Args: cobra.ExactArgs(1), RunE: s.runHcH7Create,
	}
	s.hcH7DeleteCmd = &cobra.Command{
		Use: "delete HL7V2_STORE", Short: "Delete a Healthcare HL7v2 store",
		Args: cobra.ExactArgs(1), RunE: s.runHcH7Delete,
	}
	s.hcH7DescribeCmd = &cobra.Command{
		Use: "describe HL7V2_STORE", Short: "Describe a Healthcare HL7v2 store",
		Args: cobra.ExactArgs(1), RunE: s.runHcH7Describe,
	}
	s.hcH7ListCmd = &cobra.Command{
		Use: "list", Short: "List Healthcare HL7v2 stores in a dataset",
		Args: cobra.NoArgs, RunE: s.runHcH7List,
	}
	s.hcH7UpdateCmd = &cobra.Command{
		Use: "update HL7V2_STORE", Short: "Update a Healthcare HL7v2 store",
		Args: cobra.ExactArgs(1), RunE: s.runHcH7Update,
	}
	s.hcH7GetIamCmd = &cobra.Command{
		Use: "get-iam-policy HL7V2_STORE", Short: "Get the IAM policy for a Healthcare HL7v2 store",
		Args: cobra.ExactArgs(1), RunE: s.runHcH7GetIam,
	}
	s.hcH7SetIamCmd = &cobra.Command{
		Use: "set-iam-policy HL7V2_STORE POLICY_FILE", Short: "Set the IAM policy for a Healthcare HL7v2 store",
		Args: cobra.ExactArgs(2), RunE: s.runHcH7SetIam,
	}
	s.healthcareOperationsCmd = &cobra.Command{Use: "operations", Short: "Manage Cloud Healthcare long-running operations"}
	s.healthcareOpCancelCmd = &cobra.Command{
		Use: "cancel OPERATION", Short: "Cancel a Cloud Healthcare operation",
		Args: cobra.ExactArgs(1), RunE: s.runHcOpCancel,
	}
	s.healthcareOpDescribeCmd = &cobra.Command{
		Use: "describe OPERATION", Short: "Describe a Cloud Healthcare operation",
		Args: cobra.ExactArgs(1), RunE: s.runHcOpDescribe,
	}
	s.healthcareOpListCmd = &cobra.Command{
		Use: "list", Short: "List Cloud Healthcare operations for a dataset",
		Args: cobra.NoArgs, RunE: s.runHcOpList,
	}
	s.initHealthcare()
	s.initHealthcareConsentStores()
	s.initHealthcareDatasets()
	s.initHealthcareDicomStores()
	s.initHealthcareFhirStores()
	s.initHealthcareHl7v2Stores()
	s.initHealthcareOperations()
	return s
}
