package cli

import (
	"fmt"
	"strings"
)

// aiParent returns "projects/PROJECT/locations/REGION" from the resolved
// project (via config) and the supplied --region value. REGION is required.
func (s *aiCmds) aiParent(region string) (string, error) {
	project, err := s.inv.project()
	if err != nil {
		return "", err
	}
	if region == "" {
		return "", fmt.Errorf("--region is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, region), nil
}

// aiChild builds "PARENT/COLLECTION/ID" unless ID is already a full resource
// path (starts with "projects/" or "publishers/").
func aiChild(collection, id, parent string) string {
	if strings.HasPrefix(id, "projects/") || strings.HasPrefix(id, "publishers/") {
		return id
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id)
}
