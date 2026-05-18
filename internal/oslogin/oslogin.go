package oslogin

import (
	"context"
	"fmt"
	"os"
	"strings"

	"google.golang.org/api/compute/v1"
	osloginapi "google.golang.org/api/oslogin/v1"
)

// IsEnabled checks whether OS Login is enabled for an instance by inspecting
// both instance-level and project-level metadata for the "enable-oslogin" key.
func IsEnabled(inst *compute.Instance, proj *compute.Project) bool {
	// Instance metadata takes precedence.
	if inst.Metadata != nil {
		for _, item := range inst.Metadata.Items {
			if strings.EqualFold(item.Key, "enable-oslogin") && item.Value != nil {
				return strings.EqualFold(*item.Value, "true")
			}
		}
	}
	// Fall back to project metadata.
	if proj != nil && proj.CommonInstanceMetadata != nil {
		for _, item := range proj.CommonInstanceMetadata.Items {
			if strings.EqualFold(item.Key, "enable-oslogin") && item.Value != nil {
				return strings.EqualFold(*item.Value, "true")
			}
		}
	}
	return false
}

// PosixUsername retrieves the POSIX username for the authenticated user in the
// given project via the OS Login API's GetLoginProfile.
func PosixUsername(ctx context.Context, svc *osloginapi.Service, email, project string) (string, error) {
	name := "users/" + email
	resp, err := svc.Users.GetLoginProfile(name).ProjectId(project).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("getting OS Login profile: %w", err)
	}
	for _, pa := range resp.PosixAccounts {
		if pa.Primary {
			return pa.Username, nil
		}
	}
	// If no primary, use the first one.
	if len(resp.PosixAccounts) > 0 {
		return resp.PosixAccounts[0].Username, nil
	}
	return "", fmt.Errorf("no POSIX account found for %s in project %s", email, project)
}

// ImportSSHKey imports the public key at pubKeyPath into the OS Login API for
// the given user/project. Returns the POSIX username from the response.
func ImportSSHKey(ctx context.Context, svc *osloginapi.Service, email, pubKeyPath, project string) (string, error) {
	pubKey, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return "", fmt.Errorf("reading public key: %w", err)
	}

	key := &osloginapi.SshPublicKey{
		Key: strings.TrimSpace(string(pubKey)),
	}

	resp, err := svc.Users.ImportSshPublicKey("users/"+email, key).ProjectId(project).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("importing SSH key via OS Login: %w", err)
	}

	// Extract POSIX username from the login profile in the response.
	if resp.LoginProfile != nil {
		for _, pa := range resp.LoginProfile.PosixAccounts {
			if pa.Primary {
				return pa.Username, nil
			}
		}
		if len(resp.LoginProfile.PosixAccounts) > 0 {
			return resp.LoginProfile.PosixAccounts[0].Username, nil
		}
	}
	return "", nil
}
