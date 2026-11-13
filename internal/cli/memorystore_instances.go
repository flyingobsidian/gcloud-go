package cli

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud memorystore instances (#978) ---

func (s *memorystoreCmds) initMemorystoreInstances() {
	all := []*cobra.Command{
		s.memstoreInstCreateCmd, s.memstoreInstDeleteCmd, s.memstoreInstDescribeCmd, s.memstoreInstListCmd, s.memstoreInstUpdateCmd,
		s.memstoreInstBackupCmd, s.memstoreInstFinishMigrationCmd, s.memstoreInstStartMigrationCmd,
		s.memstoreInstGetCACmd, s.memstoreInstGetSharedCACmd, s.memstoreInstRescheduleMaintCmd,
	}
	for _, c := range all {
		c.Flags().StringVar(&s.flagMemstoreInstLocation, "location", "", "Memorystore location (required)")
		_ = c.MarkFlagRequired("location")
		c.Flags().StringVar(&s.flagMemstoreInstFormat, "format", "", "Output format")
	}
	for _, c := range []*cobra.Command{s.memstoreInstCreateCmd, s.memstoreInstUpdateCmd, s.memstoreInstStartMigrationCmd, s.memstoreInstRescheduleMaintCmd} {
		c.Flags().StringVar(&s.flagMemstoreInstConfigFile, "config-file", "",
			"Path to a YAML/JSON file with the request body (required)")
		_ = c.MarkFlagRequired("config-file")
	}
	// backup and finish-migration accept optional config, defaulting to {}.
	s.memstoreInstBackupCmd.Flags().StringVar(&s.flagMemstoreInstConfigFile, "config-file", "",
		"Optional path to a YAML/JSON file with the backup request body")
	s.memstoreInstFinishMigrationCmd.Flags().StringVar(&s.flagMemstoreInstConfigFile, "config-file", "",
		"Optional path to a YAML/JSON file with the finish-migration request body")
	s.memstoreInstUpdateCmd.Flags().StringVar(&s.flagMemstoreInstUpdateMask, "update-mask", "",
		"Comma-separated list of fields to update (defaults to every populated field)")
	s.memstoreInstListCmd.Flags().Int64Var(&s.flagMemstoreInstPageSize, "page-size", 0, "Maximum results per page")
	// --zone-distribution-config-zones added in gcloud-python 574.0.0.
	s.memstoreInstCreateCmd.Flags().StringSliceVar(&s.flagMemstoreInstZDCZones, "zone-distribution-config-zones", nil,
		"Zones for a MULTI_ZONE cluster (comma-separated); overrides zoneDistributionConfig.zones from --config-file")

	s.memstoreInstCmd.AddCommand(all...)
	s.memorystoreCmd.AddCommand(s.memstoreInstCmd)
}

func (s *memorystoreCmds) memstoreInstParent() (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, s.flagMemstoreInstLocation), nil
}

func (s *memorystoreCmds) memstoreInstName(id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := s.memstoreInstParent()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/instances/%s", parent, id), nil
}

// memstoreInstOptionalBody returns the parsed --config-file body, or {} if the flag
// was not set. Used by RPCs where the body is optional.
func (s *memorystoreCmds) memstoreInstOptionalBody() (map[string]any, error) {
	body := map[string]any{}
	if s.flagMemstoreInstConfigFile == "" {
		return body, nil
	}
	if err := loadYAMLOrJSONInto(s.flagMemstoreInstConfigFile, &body); err != nil {
		return nil, err
	}
	return body, nil
}

func (s *memorystoreCmds) runMemstoreInstCreate(cmd *cobra.Command, args []string) error {
	parent, err := s.memstoreInstParent()
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagMemstoreInstConfigFile, &body); err != nil {
		return err
	}
	applyZoneDistributionZonesToMap(body, s.flagMemstoreInstZDCZones)
	q := url.Values{}
	q.Set("instanceId", args[0])
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodPost, "/"+parent+"/instances", q, body, &op); err != nil {
		return fmt.Errorf("creating instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Create request issued for instance [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagMemstoreInstFormat)
}

// applyZoneDistributionZonesToMap sets body.zoneDistributionConfig.zones from
// the CLI slice, preserving any mode/zone already present. Added in
// gcloud-python 574.0.0.
func applyZoneDistributionZonesToMap(body map[string]any, zones []string) {
	if len(zones) == 0 {
		return
	}
	zdc, _ := body["zoneDistributionConfig"].(map[string]any)
	if zdc == nil {
		zdc = map[string]any{}
	}
	converted := make([]any, len(zones))
	for i, z := range zones {
		converted[i] = z
	}
	zdc["zones"] = converted
	body["zoneDistributionConfig"] = zdc
}

func (s *memorystoreCmds) runMemstoreInstDelete(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodDelete, "/"+name, nil, nil, &op); err != nil {
		return fmt.Errorf("deleting instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Delete request issued for instance [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagMemstoreInstFormat)
}

func (s *memorystoreCmds) runMemstoreInstDescribe(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodGet, "/"+name, nil, nil, &got); err != nil {
		return fmt.Errorf("describing instance: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMemstoreInstFormat)
}

func (s *memorystoreCmds) runMemstoreInstList(cmd *cobra.Command, args []string) error {
	parent, err := s.memstoreInstParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	items, err := s.inv.rest(memorystoreRest()).paginate(ctx, "/"+parent+"/instances", nil, "instances", s.flagMemstoreInstPageSize)
	if err != nil {
		return fmt.Errorf("listing instances: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, items, s.flagMemstoreInstFormat)
}

func (s *memorystoreCmds) runMemstoreInstUpdate(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreInstName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagMemstoreInstConfigFile, &body); err != nil {
		return err
	}
	mask := s.flagMemstoreInstUpdateMask
	if mask == "" {
		mask = joinMask(dcTopLevelKeys(body))
	}
	q := url.Values{}
	if mask != "" {
		q.Set("updateMask", mask)
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodPatch, "/"+name, q, body, &op); err != nil {
		return fmt.Errorf("updating instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Update request issued for instance [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagMemstoreInstFormat)
}

func (s *memorystoreCmds) runMemstoreInstBackup(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreInstName(args[0])
	if err != nil {
		return err
	}
	body, err := s.memstoreInstOptionalBody()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodPost, "/"+name+":backup", nil, body, &op); err != nil {
		return fmt.Errorf("backing up instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Backup request issued for instance [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagMemstoreInstFormat)
}

func (s *memorystoreCmds) runMemstoreInstFinishMigration(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreInstName(args[0])
	if err != nil {
		return err
	}
	body, err := s.memstoreInstOptionalBody()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodPost, "/"+name+":finishMigration", nil, body, &op); err != nil {
		return fmt.Errorf("finishing migration for instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Finish-migration request issued for instance [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagMemstoreInstFormat)
}

func (s *memorystoreCmds) runMemstoreInstStartMigration(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreInstName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagMemstoreInstConfigFile, &body); err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodPost, "/"+name+":startMigration", nil, body, &op); err != nil {
		return fmt.Errorf("starting migration for instance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Start-migration request issued for instance [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagMemstoreInstFormat)
}

func (s *memorystoreCmds) runMemstoreInstGetCA(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreInstName(args[0])
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodGet, "/"+name+"/certificateAuthority", nil, nil, &got); err != nil {
		return fmt.Errorf("getting certificate authority: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMemstoreInstFormat)
}

func (s *memorystoreCmds) runMemstoreInstGetSharedCA(cmd *cobra.Command, args []string) error {
	parent, err := s.memstoreInstParent()
	if err != nil {
		return err
	}
	ctx := cmd.Context()
	var got map[string]any
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodGet, "/"+parent+":getSharedRegionalCertificateAuthority", nil, nil, &got); err != nil {
		return fmt.Errorf("getting shared regional certificate authority: %w", err)
	}
	return emitFormattedTo(s.inv.stdout, got, s.flagMemstoreInstFormat)
}

func (s *memorystoreCmds) runMemstoreInstRescheduleMaint(cmd *cobra.Command, args []string) error {
	name, err := s.memstoreInstName(args[0])
	if err != nil {
		return err
	}
	body := map[string]any{}
	if err := loadYAMLOrJSONInto(s.flagMemstoreInstConfigFile, &body); err != nil {
		return err
	}
	ctx := cmd.Context()
	var op map[string]any
	if err := s.inv.rest(memorystoreRest()).do(ctx, http.MethodPost, "/"+name+":rescheduleMaintenance", nil, body, &op); err != nil {
		return fmt.Errorf("rescheduling maintenance: %w", err)
	}
	fmt.Fprintf(s.inv.stdout, "Reschedule-maintenance request issued for instance [%s].\n", args[0])
	return emitFormattedTo(s.inv.stdout, op, s.flagMemstoreInstFormat)
}
