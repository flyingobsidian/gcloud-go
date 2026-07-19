package cmd

import (
	"fmt"
	"strings"
)

// odbLocationParent returns projects/PROJECT/locations/LOCATION.
func odbLocationParent(location string) (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	if location == "" {
		return "", fmt.Errorf("--location is required")
	}
	return fmt.Sprintf("projects/%s/locations/%s", project, location), nil
}

// odbResource returns projects/.../COLLECTION/ID.
func odbResource(location, collection, id string) (string, error) {
	if strings.HasPrefix(id, "projects/") {
		return id, nil
	}
	parent, err := odbLocationParent(location)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", parent, collection, id), nil
}
