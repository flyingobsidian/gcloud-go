package cmd

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	managedkafka "google.golang.org/api/managedkafka/v1"
)

// applyBrokerDiskFlag validates --broker-disk-size-gib (using the same
// byte-unit parser as gcloud-python 576.0.0) and threads it onto the Cluster
// request body. The managedkafka v1 discovery in this SDK release does not yet
// expose a broker_capacity_config field on Cluster; when the client library
// adds the field we can populate cluster.BrokerCapacityConfig.DiskSizeGib
// directly. Until then the flag is accepted, validated, and logged so callers
// see a clear diagnostic rather than a silent no-op.
func applyBrokerDiskFlag(_ *managedkafka.Cluster) error {
	if flagMKBrokerDiskGiB == "" {
		return nil
	}
	gib, err := parseBrokerDiskSizeGiB(flagMKBrokerDiskGiB)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr,
		"Warning: --broker-disk-size-gib=%dGiB parsed successfully; the current managedkafka v1 SDK does not yet expose brokerCapacityConfig.diskSizeGib, so include it in --config-file for now.\n",
		gib,
	)
	return nil
}

// --- Managed Kafka broker-disk flags (#1783) ---
//
// gcloud-python 576.0.0 introduced a shared "byte units" parser for the
// broker-disk flag, and 582.0.0 promoted broker-disk to GA. gcloud-go
// mirrors both changes here.

// parseBrokerDiskSizeGiB accepts a raw --broker-disk-size-gib value from the
// CLI and returns the size in gibibytes. The gcloud-python util accepts a
// plain integer or an integer plus a byte-unit suffix (``B``, ``KiB``, ``KB``,
// ``MiB``, ``MB``, ``GiB``, ``GB``, ``TiB``, ``TB``, etc.) and normalises the
// value to GiB. Fractional GiB values are rejected to match the API contract.
func parseBrokerDiskSizeGiB(raw string) (int64, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, fmt.Errorf("empty broker-disk-size value")
	}

	// Fast path: plain integer.
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		if n < 0 {
			return 0, fmt.Errorf("broker-disk-size must be non-negative")
		}
		return n, nil
	}

	// Split the numeric prefix from the byte-unit suffix.
	unitStart := 0
	for unitStart < len(s) && (s[unitStart] == '-' || s[unitStart] == '+' || (s[unitStart] >= '0' && s[unitStart] <= '9') || s[unitStart] == '.') {
		unitStart++
	}
	if unitStart == 0 {
		return 0, fmt.Errorf("broker-disk-size %q: missing number", raw)
	}
	numPart := s[:unitStart]
	unit := strings.TrimSpace(s[unitStart:])
	// Compute size in bytes so unit conversions are consistent, then convert to GiB.
	value, err := strconv.ParseFloat(numPart, 64)
	if err != nil {
		return 0, fmt.Errorf("broker-disk-size %q: invalid number: %w", raw, err)
	}
	if value < 0 {
		return 0, fmt.Errorf("broker-disk-size must be non-negative")
	}
	unit = strings.ToLower(unit)
	multiplierBytes := float64(0)
	const (
		kib = 1024.0
		mib = kib * 1024.0
		gib = mib * 1024.0
		tib = gib * 1024.0
		pib = tib * 1024.0
	)
	switch unit {
	case "", "b":
		multiplierBytes = 1
	case "k", "kb":
		multiplierBytes = 1000
	case "ki", "kib":
		multiplierBytes = kib
	case "m", "mb":
		multiplierBytes = 1000 * 1000
	case "mi", "mib":
		multiplierBytes = mib
	case "g", "gb":
		multiplierBytes = 1000 * 1000 * 1000
	case "gi", "gib":
		multiplierBytes = gib
	case "t", "tb":
		multiplierBytes = 1000 * 1000 * 1000 * 1000
	case "ti", "tib":
		multiplierBytes = tib
	case "p", "pb":
		multiplierBytes = 1000 * 1000 * 1000 * 1000 * 1000
	case "pi", "pib":
		multiplierBytes = pib
	default:
		return 0, fmt.Errorf("broker-disk-size %q: unknown unit %q", raw, unit)
	}
	bytes := value * multiplierBytes
	// Convert to GiB, matching the api_field expected by broker-disk-size-gib.
	gibValue := bytes / gib
	rounded := int64(gibValue)
	if float64(rounded) != gibValue {
		return 0, fmt.Errorf("broker-disk-size %q must be an integer number of GiB", raw)
	}
	return rounded, nil
}
