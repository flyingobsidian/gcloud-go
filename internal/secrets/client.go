package secrets

import (
	"fmt"
)

// SecretParent returns the parent resource for secrets in a project.
func SecretParent(project, location string) string {
	if location != "" {
		return fmt.Sprintf("projects/%s/locations/%s", project, location)
	}
	return fmt.Sprintf("projects/%s", project)
}

// SecretName returns the full resource name for a secret.
func SecretName(project, secretID, location string) string {
	if location != "" {
		return fmt.Sprintf("projects/%s/locations/%s/secrets/%s", project, location, secretID)
	}
	return fmt.Sprintf("projects/%s/secrets/%s", project, secretID)
}

// VersionName returns the full resource name for a secret version.
func VersionName(project, secretID, version, location string) string {
	if location != "" {
		return fmt.Sprintf("projects/%s/locations/%s/secrets/%s/versions/%s", project, location, secretID, version)
	}
	return fmt.Sprintf("projects/%s/secrets/%s/versions/%s", project, secretID, version)
}
