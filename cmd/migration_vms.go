package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	"github.com/spf13/cobra"
	vmmigration "google.golang.org/api/vmmigration/v1"
)

// --- gcloud migration vms target-projects and image-imports (#895, #896) ---

var (
	flagMVMLocation      string
	flagMVMFormat        string
	flagMVMFilter        string
	flagMVMAsync         bool
	flagMVMTargetProj    string
	flagMVMDescription   string
	flagMVMUpdateMask    string
	flagMVMConfigFile    string
	flagMVMSourceFile    string
	flagMVMImageName     string
	flagMVMImportTP      string
	flagMVMFamily        string
	flagMVMLabels        []string
	flagMVMAddLicenses   []string
	flagMVMSingleRegion  bool
	flagMVMKmsKey        string
	flagMVMSkipOSAdapt   bool
	flagMVMGeneralize    bool
	flagMVMLicenseType   string
	flagMVMBootConv      string
)

func mvmService(ctx context.Context) (*vmmigration.Service, error) {
	return gcp.VMMigrationService(ctx, flagAccount)
}

// mvmLocationParent returns projects/PROJECT/locations/LOCATION.
func mvmLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func mvmResourceName(parent, collection, id string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}

func mvmResolveLocationParent(defaultToGlobal bool) (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	loc := flagMVMLocation
	if loc == "" && defaultToGlobal {
		loc = "global"
	}
	if loc == "" {
		return "", fmt.Errorf("--location is required")
	}
	return mvmLocationParent(project, loc), nil
}

func mvmWaitOp(ctx context.Context, svc *vmmigration.Service, op *vmmigration.Operation) (*vmmigration.Operation, error) {
	for !op.Done {
		time.Sleep(3 * time.Second)
		got, err := svc.Projects.Locations.Operations.Get(op.Name).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("polling operation %s: %w", op.Name, err)
		}
		op = got
	}
	if op.Error != nil {
		return op, fmt.Errorf("operation %s failed: %s", op.Name, op.Error.Message)
	}
	return op, nil
}

func mvmFinishOp(ctx context.Context, svc *vmmigration.Service, op *vmmigration.Operation, verb, target string) error {
	if flagMVMAsync {
		fmt.Fprintf(os.Stderr, "%s in progress (operation: %s).\n", verb, op.Name)
		return emitFormatted(op, flagMVMFormat)
	}
	final, err := mvmWaitOp(ctx, svc, op)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "%s [%s] completed.\n", verb, target)
	return emitFormatted(final, flagMVMFormat)
}

// --- target-projects ---

var mvmTargetProjectsCmd = &cobra.Command{
	Use: "target-projects", Short: "Manage VM Migration target projects",
}

var (
	mvmTPAddCmd = &cobra.Command{
		Use: "add TARGET_PROJECT", Short: "Create a target project",
		Args: cobra.ExactArgs(1), RunE: runMVMTPCreate,
	}
	mvmTPDeleteCmd = &cobra.Command{
		Use: "delete TARGET_PROJECT", Short: "Delete a target project",
		Args: cobra.ExactArgs(1), RunE: runMVMTPDelete,
	}
	mvmTPDescribeCmd = &cobra.Command{
		Use: "describe TARGET_PROJECT", Short: "Describe a target project",
		Args: cobra.ExactArgs(1), RunE: runMVMTPDescribe,
	}
	mvmTPListCmd = &cobra.Command{
		Use: "list", Short: "List target projects",
		Args: cobra.NoArgs, RunE: runMVMTPList,
	}
	mvmTPUpdateCmd = &cobra.Command{
		Use: "update TARGET_PROJECT", Short: "Update a target project",
		Args: cobra.ExactArgs(1), RunE: runMVMTPUpdate,
	}
)

// --- image-imports ---

var mvmImageImportsCmd = &cobra.Command{
	Use: "image-imports", Short: "Manage VM Migration image imports",
}

var (
	mvmIICreateCmd = &cobra.Command{
		Use: "create IMAGE_IMPORT", Short: "Create an image import",
		Args: cobra.ExactArgs(1), RunE: runMVMIICreate,
	}
	mvmIIDeleteCmd = &cobra.Command{
		Use: "delete IMAGE_IMPORT", Short: "Delete an image import",
		Args: cobra.ExactArgs(1), RunE: runMVMIIDelete,
	}
	mvmIIDescribeCmd = &cobra.Command{
		Use: "describe IMAGE_IMPORT", Short: "Describe an image import",
		Args: cobra.ExactArgs(1), RunE: runMVMIIDescribe,
	}
	mvmIIListCmd = &cobra.Command{
		Use: "list", Short: "List image imports",
		Args: cobra.NoArgs, RunE: runMVMIIList,
	}
)

func init() {
	addLoc := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagMVMLocation, "location", "", "Location (region); defaults to 'global' for target-projects")
		}
	}
	addFmt := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagMVMFormat, "format", "", "Output format")
		}
	}
	addFilter := func(cmds ...*cobra.Command) {
		for _, c := range cmds {
			c.Flags().StringVar(&flagMVMFilter, "filter", "", "Server-side list filter")
		}
	}

	// target-projects
	addLoc(mvmTPAddCmd, mvmTPDeleteCmd, mvmTPDescribeCmd, mvmTPListCmd, mvmTPUpdateCmd)
	addFmt(mvmTPAddCmd, mvmTPDescribeCmd, mvmTPListCmd, mvmTPUpdateCmd)
	addFilter(mvmTPListCmd)
	mvmTPAddCmd.Flags().StringVar(&flagMVMTargetProj, "target-project", "", "The Compute Engine project ID or number (required)")
	_ = mvmTPAddCmd.MarkFlagRequired("target-project")
	mvmTPAddCmd.Flags().StringVar(&flagMVMDescription, "description", "", "Optional description")
	mvmTPAddCmd.Flags().BoolVar(&flagMVMAsync, "async", false, "Do not wait for the operation to finish")
	mvmTPDeleteCmd.Flags().BoolVar(&flagMVMAsync, "async", false, "Do not wait for the operation to finish")
	mvmTPUpdateCmd.Flags().StringVar(&flagMVMTargetProj, "target-project", "", "Update the Compute Engine project ID")
	mvmTPUpdateCmd.Flags().StringVar(&flagMVMDescription, "description", "", "Update the description")
	mvmTPUpdateCmd.Flags().StringVar(&flagMVMUpdateMask, "update-mask", "", "Comma-separated list of fields to update")
	mvmTPUpdateCmd.Flags().BoolVar(&flagMVMAsync, "async", false, "Do not wait for the operation to finish")
	mvmTargetProjectsCmd.AddCommand(mvmTPAddCmd, mvmTPDeleteCmd, mvmTPDescribeCmd, mvmTPListCmd, mvmTPUpdateCmd)
	migrationVMsCmd.AddCommand(mvmTargetProjectsCmd)

	// image-imports
	addLoc(mvmIICreateCmd, mvmIIDeleteCmd, mvmIIDescribeCmd, mvmIIListCmd)
	addFmt(mvmIICreateCmd, mvmIIDescribeCmd, mvmIIListCmd)
	addFilter(mvmIIListCmd)
	mvmIICreateCmd.Flags().StringVar(&flagMVMSourceFile, "source-file", "", "Cloud Storage URI (gs://bucket/file.vmdk) of the source image (required)")
	_ = mvmIICreateCmd.MarkFlagRequired("source-file")
	mvmIICreateCmd.Flags().StringVar(&flagMVMImageName, "image-name", "", "Target Compute Engine image name (defaults to the import resource ID)")
	mvmIICreateCmd.Flags().StringVar(&flagMVMImportTP, "target-project", "", "Target project resource path (projects/*/locations/global/targetProjects/*) (required)")
	_ = mvmIICreateCmd.MarkFlagRequired("target-project")
	mvmIICreateCmd.Flags().StringVar(&flagMVMDescription, "description", "", "Optional description for the target image")
	mvmIICreateCmd.Flags().StringVar(&flagMVMFamily, "family-name", "", "Compute Engine image family")
	mvmIICreateCmd.Flags().StringSliceVar(&flagMVMLabels, "labels", nil, "KEY=VALUE labels on the target image")
	mvmIICreateCmd.Flags().StringSliceVar(&flagMVMAddLicenses, "additional-licenses", nil, "Additional Compute Engine license URIs")
	mvmIICreateCmd.Flags().BoolVar(&flagMVMSingleRegion, "single-region-storage", false, "Store image in a single region")
	mvmIICreateCmd.Flags().StringVar(&flagMVMKmsKey, "kms-key", "", "Cloud KMS key used to encrypt the image import (projects/.../cryptoKeys/...)")
	mvmIICreateCmd.Flags().BoolVar(&flagMVMSkipOSAdapt, "skip-os-adaptation", false, "Skip OS adaptation")
	mvmIICreateCmd.Flags().BoolVar(&flagMVMGeneralize, "generalize", false, "Generalize the image (Windows)")
	mvmIICreateCmd.Flags().StringVar(&flagMVMLicenseType, "license-type", "", "License type (default, payg, byol)")
	mvmIICreateCmd.Flags().StringVar(&flagMVMBootConv, "boot-conversion", "", "Boot conversion (none, bios-to-efi)")
	mvmIICreateCmd.Flags().StringVar(&flagMVMConfigFile, "config-file", "", "JSON/YAML ImageImport body override")
	mvmIICreateCmd.Flags().BoolVar(&flagMVMAsync, "async", false, "Do not wait for the operation to finish")
	mvmIIDeleteCmd.Flags().BoolVar(&flagMVMAsync, "async", false, "Do not wait for the operation to finish")
	mvmImageImportsCmd.AddCommand(mvmIICreateCmd, mvmIIDeleteCmd, mvmIIDescribeCmd, mvmIIListCmd)
	migrationVMsCmd.AddCommand(mvmImageImportsCmd)
}

// --- target-projects impl ---

func runMVMTPCreate(cmd *cobra.Command, args []string) error {
	parent, err := mvmResolveLocationParent(true)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
	if err != nil {
		return err
	}
	body := &vmmigration.TargetProject{
		Project:     flagMVMTargetProj,
		Description: flagMVMDescription,
	}
	op, err := svc.Projects.Locations.TargetProjects.Create(parent, body).TargetProjectId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating target project: %w", err)
	}
	return mvmFinishOp(ctx, svc, op, "Create target project", args[0])
}

func runMVMTPDelete(cmd *cobra.Command, args []string) error {
	parent, err := mvmResolveLocationParent(true)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
	if err != nil {
		return err
	}
	name := mvmResourceName(parent, "targetProjects", args[0])
	op, err := svc.Projects.Locations.TargetProjects.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting target project: %w", err)
	}
	return mvmFinishOp(ctx, svc, op, "Delete target project", args[0])
}

func runMVMTPDescribe(cmd *cobra.Command, args []string) error {
	parent, err := mvmResolveLocationParent(true)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
	if err != nil {
		return err
	}
	name := mvmResourceName(parent, "targetProjects", args[0])
	got, err := svc.Projects.Locations.TargetProjects.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing target project: %w", err)
	}
	return emitFormatted(got, flagMVMFormat)
}

func runMVMTPList(cmd *cobra.Command, args []string) error {
	parent, err := mvmResolveLocationParent(true)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
	if err != nil {
		return err
	}
	var all []*vmmigration.TargetProject
	pageToken := ""
	for {
		call := svc.Projects.Locations.TargetProjects.List(parent).Context(ctx)
		if flagMVMFilter != "" {
			call = call.Filter(flagMVMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing target projects: %w", err)
		}
		all = append(all, resp.TargetProjects...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagMVMFormat != "" {
		return emitFormatted(all, flagMVMFormat)
	}
	fmt.Printf("%-40s %-30s %s\n", "NAME", "PROJECT", "CREATE_TIME")
	for _, tp := range all {
		fmt.Printf("%-40s %-30s %s\n", path.Base(tp.Name), tp.Project, tp.CreateTime)
	}
	return nil
}

func runMVMTPUpdate(cmd *cobra.Command, args []string) error {
	parent, err := mvmResolveLocationParent(true)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
	if err != nil {
		return err
	}
	body := &vmmigration.TargetProject{}
	var mask []string
	if flagMVMTargetProj != "" {
		body.Project = flagMVMTargetProj
		mask = append(mask, "project")
	}
	if flagMVMDescription != "" {
		body.Description = flagMVMDescription
		mask = append(mask, "description")
	}
	if flagMVMUpdateMask != "" {
		mask = strings.Split(flagMVMUpdateMask, ",")
	}
	name := mvmResourceName(parent, "targetProjects", args[0])
	call := svc.Projects.Locations.TargetProjects.Patch(name, body).Context(ctx)
	if len(mask) > 0 {
		call = call.UpdateMask(strings.Join(mask, ","))
	}
	op, err := call.Do()
	if err != nil {
		return fmt.Errorf("updating target project: %w", err)
	}
	return mvmFinishOp(ctx, svc, op, "Update target project", args[0])
}

// --- image-imports impl ---

func runMVMIICreate(cmd *cobra.Command, args []string) error {
	parent, err := mvmResolveLocationParent(false)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
	if err != nil {
		return err
	}
	body := &vmmigration.ImageImport{
		CloudStorageUri: flagMVMSourceFile,
	}
	if flagMVMConfigFile != "" {
		if err := loadYAMLOrJSONInto(flagMVMConfigFile, body); err != nil {
			return err
		}
	}
	if body.DiskImageTargetDefaults == nil {
		body.DiskImageTargetDefaults = &vmmigration.DiskImageTargetDetails{}
	}
	target := body.DiskImageTargetDefaults
	if flagMVMImageName != "" {
		target.ImageName = flagMVMImageName
	} else if target.ImageName == "" {
		target.ImageName = args[0]
	}
	target.TargetProject = flagMVMImportTP
	if flagMVMDescription != "" {
		target.Description = flagMVMDescription
	}
	if flagMVMFamily != "" {
		target.FamilyName = flagMVMFamily
	}
	if len(flagMVMLabels) > 0 {
		if target.Labels == nil {
			target.Labels = map[string]string{}
		}
		for _, kv := range flagMVMLabels {
			k, v, ok := strings.Cut(kv, "=")
			if !ok {
				return fmt.Errorf("invalid --labels entry %q (want KEY=VALUE)", kv)
			}
			target.Labels[k] = v
		}
	}
	if len(flagMVMAddLicenses) > 0 {
		target.AdditionalLicenses = append(target.AdditionalLicenses, flagMVMAddLicenses...)
	}
	target.SingleRegionStorage = flagMVMSingleRegion
	if flagMVMKmsKey != "" {
		body.Encryption = &vmmigration.Encryption{KmsKey: flagMVMKmsKey}
	}
	if flagMVMSkipOSAdapt {
		target.DataDiskImageImport = &vmmigration.DataDiskImageImport{}
		target.OsAdaptationParameters = nil
	} else if flagMVMGeneralize || flagMVMLicenseType != "" || flagMVMBootConv != "" {
		if target.OsAdaptationParameters == nil {
			target.OsAdaptationParameters = &vmmigration.ImageImportOsAdaptationParameters{}
		}
		p := target.OsAdaptationParameters
		p.Generalize = flagMVMGeneralize
		switch strings.ToLower(flagMVMLicenseType) {
		case "":
		case "default":
			p.LicenseType = "COMPUTE_ENGINE_LICENSE_TYPE_DEFAULT"
		case "payg":
			p.LicenseType = "COMPUTE_ENGINE_LICENSE_TYPE_PAYG"
		case "byol":
			p.LicenseType = "COMPUTE_ENGINE_LICENSE_TYPE_BYOL"
		default:
			return fmt.Errorf("invalid --license-type %q (want default, payg, or byol)", flagMVMLicenseType)
		}
		switch strings.ToLower(flagMVMBootConv) {
		case "":
		case "none":
			p.BootConversion = "NONE"
		case "bios-to-efi":
			p.BootConversion = "BIOS_TO_EFI"
		default:
			return fmt.Errorf("invalid --boot-conversion %q (want none or bios-to-efi)", flagMVMBootConv)
		}
	}
	op, err := svc.Projects.Locations.ImageImports.Create(parent, body).ImageImportId(args[0]).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("creating image import: %w", err)
	}
	return mvmFinishOp(ctx, svc, op, "Create image import", args[0])
}

func runMVMIIDelete(cmd *cobra.Command, args []string) error {
	parent, err := mvmResolveLocationParent(false)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
	if err != nil {
		return err
	}
	name := mvmResourceName(parent, "imageImports", args[0])
	op, err := svc.Projects.Locations.ImageImports.Delete(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("deleting image import: %w", err)
	}
	return mvmFinishOp(ctx, svc, op, "Delete image import", args[0])
}

func runMVMIIDescribe(cmd *cobra.Command, args []string) error {
	parent, err := mvmResolveLocationParent(false)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
	if err != nil {
		return err
	}
	name := mvmResourceName(parent, "imageImports", args[0])
	got, err := svc.Projects.Locations.ImageImports.Get(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("describing image import: %w", err)
	}
	return emitFormatted(got, flagMVMFormat)
}

func runMVMIIList(cmd *cobra.Command, args []string) error {
	parent, err := mvmResolveLocationParent(false)
	if err != nil {
		return err
	}
	ctx := context.Background()
	svc, err := mvmService(ctx)
	if err != nil {
		return err
	}
	var all []*vmmigration.ImageImport
	pageToken := ""
	for {
		call := svc.Projects.Locations.ImageImports.List(parent).Context(ctx)
		if flagMVMFilter != "" {
			call = call.Filter(flagMVMFilter)
		}
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		resp, err := call.Do()
		if err != nil {
			return fmt.Errorf("listing image imports: %w", err)
		}
		all = append(all, resp.ImageImports...)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
	if flagMVMFormat != "" {
		return emitFormatted(all, flagMVMFormat)
	}
	fmt.Printf("%-40s %-60s %s\n", "NAME", "CLOUD_STORAGE_URI", "CREATE_TIME")
	for _, ii := range all {
		fmt.Printf("%-40s %-60s %s\n", path.Base(ii.Name), ii.CloudStorageUri, ii.CreateTime)
	}
	return nil
}
