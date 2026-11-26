package cli

import (
	"bytes"
	"encoding/json"
	"time"

	storage "google.golang.org/api/storage/v1"
)

// --- storage bucket display records (#2016) ---
//
// gcloud-python's `storage buckets list` returns, for each bucket, the dict
// that resource_util.get_display_dict_for_resource builds from a
// GcsBucketResource (full_resource_formatter.BucketDisplayTitlesAndDefaults):
// snake_case keys, trimmed ACL entries and reformatted timestamps, with unset
// fields omitted. --raw returns the API's bucket metadata instead.

// storageDisplayTimeFormat is gcloud-python's
// get_formatted_string_from_datetime_object (%Y-%m-%dT%H:%M:%S%z).
const storageDisplayTimeFormat = "2006-01-02T15:04:05-0700"

// storageDisplayTime reformats an RFC 3339 API timestamp as gcloud-python
// displays a top-level datetime; ok is false when ts is unset or invalid.
func storageDisplayTime(ts string) (string, bool) {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return "", false
	}
	return t.Format(storageDisplayTimeFormat), true
}

// storageISOTime reformats an RFC 3339 timestamp as Python's
// datetime.isoformat(), which apitools uses for DateTimeFields nested in a
// message dict: microseconds only when non-zero, and a +HH:MM offset.
func storageISOTime(ts string) string {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return ts
	}
	if t.Nanosecond()/1000 == 0 {
		return t.Format("2006-01-02T15:04:05-07:00")
	}
	return t.Format("2006-01-02T15:04:05.000000-07:00")
}

// storageNestedTimeKeys are the Bucket fields that are DateTimeFields in
// gcloud-python's storage v1 messages.
func storageNestedTimeKeys() map[string]bool {
	return map[string]bool{
		"effectiveTime": true, "lockedTime": true,
		"toggleTime": true, "terminalStorageClassUpdateTime": true,
		// Top-level Bucket timestamps, for --raw.
		"timeCreated": true, "updated": true, "softDeleteTime": true, "hardDeleteTime": true,
	}
}

// storageMessageDict converts a bucket sub-message to the dict
// gcloud-python's _message_to_dict produces: nil for an unset message or an
// empty list, with nested timestamps in isoformat.
func storageMessageDict(v any) any {
	data, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var out any
	if err := dec.Decode(&out); err != nil || out == nil {
		return nil
	}
	if l, ok := out.([]any); ok && len(l) == 0 {
		return nil
	}
	storageFixNestedTimes(out, storageNestedTimeKeys())
	return out
}

func storageFixNestedTimes(v any, keys map[string]bool) {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if s, ok := val.(string); ok && keys[k] {
				t[k] = storageISOTime(s)
				continue
			}
			storageFixNestedTimes(val, keys)
		}
	case []any:
		for _, e := range t {
			storageFixNestedTimes(e, keys)
		}
	}
}

// storageFormattedACL is gcloud-python's _get_formatted_acl: the ACL entries
// without kind, bucket, id, selfLink and etag.
func storageFormattedACL(acl any) any {
	d := storageMessageDict(acl)
	entries, ok := d.([]any)
	if !ok {
		return d
	}
	for _, e := range entries {
		if m, ok := e.(map[string]any); ok {
			for _, k := range []string{"kind", "bucket", "id", "selfLink", "etag"} {
				delete(m, k)
			}
		}
	}
	return entries
}

// storageBucketDisplay returns gcloud-python's display dict for b.
func storageBucketDisplay(b *storage.Bucket) map[string]any {
	d := map[string]any{"storage_url": "gs://" + b.Name + "/"}
	set := func(k string, v any) {
		switch t := v.(type) {
		case nil:
			return
		case string:
			if t == "" {
				return
			}
		}
		d[k] = v
	}
	setTime := func(k, ts string) {
		if s, ok := storageDisplayTime(ts); ok {
			d[k] = s
		}
	}

	set("name", b.Name)
	set("default_storage_class", b.StorageClass)
	set("location_type", b.LocationType)
	set("location", b.Location)
	if b.CustomPlacementConfig != nil && len(b.CustomPlacementConfig.DataLocations) > 0 {
		d["data_locations"] = b.CustomPlacementConfig.DataLocations
	}
	if b.Versioning != nil {
		d["versioning_enabled"] = b.Versioning.Enabled
	}
	if b.Logging != nil {
		set("logging_config", storageMessageDict(b.Logging))
	}
	if b.Website != nil {
		set("website_config", storageMessageDict(b.Website))
	}
	set("cors_config", storageMessageDict(b.Cors))
	if b.Lifecycle != nil {
		set("lifecycle_config", storageMessageDict(b.Lifecycle))
	}
	if b.Billing != nil {
		d["requester_pays"] = b.Billing.RequesterPays
	}
	if b.ObjectRetention != nil {
		set("per_object_retention", storageMessageDict(b.ObjectRetention))
	}
	if b.RetentionPolicy != nil {
		set("retention_policy", storageMessageDict(b.RetentionPolicy))
	}
	if b.DefaultEventBasedHold {
		d["default_event_based_hold"] = true
	}
	if b.IpFilter != nil {
		set("ip_filter_config", storageMessageDict(b.IpFilter))
	}
	if len(b.Labels) > 0 {
		d["labels"] = b.Labels
	}
	if e := b.Encryption; e != nil {
		set("default_kms_key", e.DefaultKmsKeyName)
		if e.GoogleManagedEncryptionEnforcementConfig != nil {
			set("gmek_enforcement", storageMessageDict(e.GoogleManagedEncryptionEnforcementConfig))
		}
		if e.CustomerManagedEncryptionEnforcementConfig != nil {
			set("cmek_enforcement", storageMessageDict(e.CustomerManagedEncryptionEnforcementConfig))
		}
		if e.CustomerSuppliedEncryptionEnforcementConfig != nil {
			set("csek_enforcement", storageMessageDict(e.CustomerSuppliedEncryptionEnforcementConfig))
		}
	}
	setTime("creation_time", b.TimeCreated)
	setTime("update_time", b.Updated)
	setTime("soft_delete_time", b.SoftDeleteTime)
	setTime("hard_delete_time", b.HardDeleteTime)
	if b.Generation != 0 {
		d["generation"] = b.Generation
	}
	if b.Metageneration != 0 {
		d["metageneration"] = b.Metageneration
	}
	if ic := b.IamConfiguration; ic != nil {
		if ic.UniformBucketLevelAccess != nil {
			d["uniform_bucket_level_access"] = ic.UniformBucketLevelAccess.Enabled
		}
		set("public_access_prevention", ic.PublicAccessPrevention)
	}
	set("rpo", b.Rpo)
	if b.Autoclass != nil {
		set("autoclass", storageMessageDict(b.Autoclass))
		if b.Autoclass.Enabled {
			setTime("autoclass_enabled_time", b.Autoclass.ToggleTime)
		}
	}
	if b.SatisfiesPZS {
		d["satisfies_pzs"] = true
	}
	if b.SoftDeletePolicy != nil {
		set("soft_delete_policy", storageMessageDict(b.SoftDeletePolicy))
	}
	set("acl", storageFormattedACL(b.Acl))
	set("default_acl", storageFormattedACL(b.DefaultObjectAcl))
	return d
}
