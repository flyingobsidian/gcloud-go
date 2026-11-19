package cli

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

func (s *computeCmds) initInstancesMetadata() {
	s.instancesAddMetadataCmd.Flags().StringToStringVar(&s.flagAddMetadataKV, "metadata", nil, "Metadata key=value pairs to add")
	s.instancesAddMetadataCmd.Flags().StringToStringVar(&s.flagAddMetadataFromFile, "metadata-from-file", nil,
		"Metadata key=filepath pairs; a filepath of - reads standard input")

	s.instancesRemoveMetadataCmd.Flags().StringVar(&s.flagRemoveMetadataKeys, "keys", "", "Comma-separated list of metadata keys to remove")
	s.instancesRemoveMetadataCmd.Flags().BoolVar(&s.flagRemoveMetadataAll, "all", false, "Remove all metadata")

	s.instancesCmd.AddCommand(s.instancesAddMetadataCmd)
	s.instancesCmd.AddCommand(s.instancesRemoveMetadataCmd)
}

func (s *computeCmds) runInstancesAddMetadata(cmd *cobra.Command, args []string) error {
	add, err := s.addMetadataDict(s.flagAddMetadataKV, s.flagAddMetadataFromFile)
	if err != nil {
		return err
	}
	instance := args[0]
	project, zone, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	inst, err := svc.Instances.Get(project, zone, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting instance: %w", err)
	}

	metadata := inst.Metadata
	if metadata == nil {
		metadata = &compute.Metadata{}
	}

	mergeMetadata(metadata, add)

	op, err := svc.Instances.SetMetadata(project, zone, instance, metadata).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting metadata: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Updated metadata for instance [%s].\n", instance)
	return nil
}

func (s *computeCmds) runInstancesRemoveMetadata(cmd *cobra.Command, args []string) error {
	if !s.flagRemoveMetadataAll && s.flagRemoveMetadataKeys == "" {
		return fmt.Errorf("one of --keys or --all must be specified")
	}

	instance := args[0]
	project, zone, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	inst, err := svc.Instances.Get(project, zone, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting instance: %w", err)
	}

	metadata := inst.Metadata
	if metadata == nil {
		metadata = &compute.Metadata{}
	}

	if s.flagRemoveMetadataAll {
		metadata.Items = nil
	} else {
		keysToRemove := make(map[string]bool)
		for _, k := range strings.Split(s.flagRemoveMetadataKeys, ",") {
			keysToRemove[strings.TrimSpace(k)] = true
		}
		var filtered []*compute.MetadataItems
		for _, item := range metadata.Items {
			if !keysToRemove[item.Key] {
				filtered = append(filtered, item)
			}
		}
		metadata.Items = filtered
	}

	op, err := svc.Instances.SetMetadata(project, zone, instance, metadata).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("setting metadata: %w", err)
	}

	if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
		return err
	}
	fmt.Fprintf(s.inv.stdout, "Updated metadata for instance [%s].\n", instance)
	return nil
}

// addMetadataDict builds the metadata to add from --metadata and
// --metadata-from-file, at least one of which is required.
func (s *computeCmds) addMetadataDict(kv, fromFile map[string]string) (map[string]string, error) {
	if len(kv) == 0 && len(fromFile) == 0 {
		return nil, fmt.Errorf("At least one of [--metadata] or [--metadata-from-file] must be provided.")
	}
	return s.constructMetadataDict(kv, fromFile)
}

// mergeMetadata sets each key of add on metadata, overwriting existing
// values and appending new keys.
func mergeMetadata(metadata *compute.Metadata, add map[string]string) {
	for _, k := range slices.Sorted(maps.Keys(add)) {
		val := add[k]
		found := false
		for _, item := range metadata.Items {
			if item.Key == k {
				item.Value = &val
				found = true
				break
			}
		}
		if !found {
			metadata.Items = append(metadata.Items, &compute.MetadataItems{Key: k, Value: &val})
		}
	}
}
