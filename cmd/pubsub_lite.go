package cmd

import (
	"fmt"
	"strings"
)

// pubsubLiteRegion returns the region derived from a Pub/Sub Lite location.
// Callers pass either a region (e.g. "us-central1") or a zone (e.g.
// "us-central1-a"). Zonal locations are collapsed to their parent region so
// callers can route requests to <region>-pubsublite.googleapis.com.
func pubsubLiteRegion(location string) (string, error) {
	if location == "" {
		return "", fmt.Errorf("--location is required for Pub/Sub Lite commands")
	}
	parts := strings.Split(location, "-")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid --location %q (want region like us-central1 or zone like us-central1-a)", location)
	}
	return strings.Join(parts[:2], "-"), nil
}

func pubsubLiteLocationParent(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func pubsubLiteChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}
