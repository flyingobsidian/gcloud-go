package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/flyingobsidian/gcloud-go/internal/auth"
	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/flyingobsidian/gcloud-go/internal/iap"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

var sshCmd = &cobra.Command{
	Use:   "ssh [USER@]INSTANCE [-- REMOTE_COMMAND]",
	Short: "SSH into a Compute Engine instance",
	Args:  cobra.MinimumNArgs(1),
	RunE:  runSSH,
}

var (
	flagSSHTunnelThroughIAP      bool
	flagSSHInternalIP            bool
	flagSSHKeyFile               string
	flagSSHCommand               string
	flagSSHFlag                  []string
	flagSSHDryRun                bool
	flagSSHPlain                 bool
	flagSSHStrictHostKeyChecking string
)

func init() {
	sshCmd.Flags().BoolVar(&flagSSHTunnelThroughIAP, "tunnel-through-iap", false, "Tunnel through Identity-Aware Proxy")
	sshCmd.Flags().BoolVar(&flagSSHInternalIP, "internal-ip", false, "Connect using internal IP")
	sshCmd.Flags().StringVar(&flagSSHKeyFile, "ssh-key-file", "", "SSH private key file")
	sshCmd.Flags().StringVar(&flagSSHCommand, "command", "", "Command to run on the instance")
	sshCmd.Flags().StringArrayVar(&flagSSHFlag, "ssh-flag", nil, "Extra flags to pass to ssh")
	sshCmd.Flags().BoolVar(&flagSSHDryRun, "dry-run", false, "Print the ssh command without running it")
	sshCmd.Flags().BoolVar(&flagSSHPlain, "plain", false, "Suppress managed SSH key setup")
	sshCmd.Flags().StringVar(&flagSSHStrictHostKeyChecking, "strict-host-key-checking", "", "Override StrictHostKeyChecking (yes, no, ask)")

	computeCmd.AddCommand(sshCmd)
}

func runSSH(cmd *cobra.Command, args []string) error {
	user, instance := parseUserInstance(args[0])
	project, zone, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	// Resolve instance to verify it exists and get its IPs.
	inst, err := svc.Instances.Get(project, zone, instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting instance %s: %w", instance, err)
	}

	// Ensure SSH key exists (unless --plain or --ssh-key-file is set).
	if !flagSSHPlain && flagSSHKeyFile == "" {
		keyPath := googleSSHKeyPath()
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			keyPath, err = generateSSHKey(keyPath)
			if err != nil {
				return fmt.Errorf("generating SSH key: %w", err)
			}
			sshUser := user
			if sshUser == "" {
				if u, err := osUser(); err == nil {
					sshUser = u
				}
			}
			if sshUser != "" {
				if err := pushSSHKeyToProject(ctx, svc, project, sshUser, keyPath+".pub"); err != nil {
					fmt.Fprintf(os.Stderr, "WARNING: could not push SSH key to project metadata: %v\n", err)
				} else {
					fmt.Fprintln(os.Stderr, "Waiting for SSH key to propagate.")
					time.Sleep(5 * time.Second)
				}
			}
		}
	}

	// Build SSH args.
	var sshArgs []string
	if flagSSHPlain {
		sshArgs = []string{}
	} else {
		sshArgs = buildSSHOpts(flagSSHKeyFile)
	}

	if flagSSHStrictHostKeyChecking != "" {
		sshArgs = append(sshArgs, "-o", "StrictHostKeyChecking="+flagSSHStrictHostKeyChecking)
	}

	for _, f := range flagSSHFlag {
		sshArgs = append(sshArgs, f)
	}

	var target string

	if flagSSHTunnelThroughIAP {
		ln, err := startIAPTunnel(ctx, project, zone, instance)
		if err != nil {
			return err
		}
		defer ln.Close()

		localPort := ln.Addr().(*net.TCPAddr).Port
		sshArgs = append(sshArgs, "-p", strconv.Itoa(localPort))
		target = "localhost"
	} else if flagSSHInternalIP {
		target = getInternalIP(inst)
		if target == "" {
			return fmt.Errorf("instance %s has no internal IP", instance)
		}
	} else {
		target = getExternalIP(inst)
		if target == "" {
			return fmt.Errorf("instance %s has no external IP; consider --tunnel-through-iap", instance)
		}
	}

	if user != "" {
		target = user + "@" + target
	}
	sshArgs = append(sshArgs, target)

	// Append remote command if provided via --command or after --.
	if flagSSHCommand != "" {
		sshArgs = append(sshArgs, "--", flagSSHCommand)
	} else if dashIdx := cmd.ArgsLenAtDash(); dashIdx >= 0 {
		sshArgs = append(sshArgs, "--")
		sshArgs = append(sshArgs, args[dashIdx:]...)
	}

	if flagSSHDryRun {
		fmt.Println("ssh " + fmt.Sprintf("%v", sshArgs))
		return nil
	}

	return execSSH(ctx, sshArgs)
}

func parseUserInstance(arg string) (user, instance string) {
	for i, c := range arg {
		if c == '@' {
			return arg[:i], arg[i+1:]
		}
	}
	return "", arg
}

func googleSSHKeyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", "google_compute_engine")
}

func googleKnownHostsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", "google_compute_known_hosts")
}

func buildSSHOpts(keyFile string) []string {
	opts := []string{}

	// Use gcloud's default SSH key unless overridden.
	if keyFile == "" {
		keyFile = googleSSHKeyPath()
	}
	if keyFile != "" {
		if _, err := os.Stat(keyFile); err == nil {
			opts = append(opts, "-i", keyFile, "-o", "IdentitiesOnly=yes")
		}
	}

	// Use gcloud's known_hosts file.
	knownHosts := googleKnownHostsPath()
	if knownHosts != "" {
		opts = append(opts, "-o", "UserKnownHostsFile="+knownHosts)
	}

	opts = append(opts,
		"-o", "StrictHostKeyChecking=no",
		"-o", "CheckHostIP=no",
	)
	return opts
}

func startIAPTunnel(ctx context.Context, project, zone, instance string) (net.Listener, error) {
	ts, err := auth.TokenSource(ctx, flagAccount,
		"https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return nil, fmt.Errorf("obtaining token source: %w", err)
	}

	ln, err := iap.Listen(ctx, iap.TunnelConfig{
		Project:     project,
		Zone:        zone,
		Instance:    instance,
		Port:        22,
		TokenSource: ts,
	}, 0)
	if err != nil {
		return nil, fmt.Errorf("starting IAP tunnel: %w", err)
	}
	return ln, nil
}

func execSSH(ctx context.Context, args []string) error {
	sshBin, err := exec.LookPath("ssh")
	if err != nil {
		return fmt.Errorf("ssh not found in PATH: %w", err)
	}

	c := exec.CommandContext(ctx, sshBin, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func getExternalIP(inst *compute.Instance) string {
	for _, ni := range inst.NetworkInterfaces {
		for _, ac := range ni.AccessConfigs {
			if ac.NatIP != "" {
				return ac.NatIP
			}
		}
	}
	return ""
}

func getInternalIP(inst *compute.Instance) string {
	if len(inst.NetworkInterfaces) > 0 {
		return inst.NetworkInterfaces[0].NetworkIP
	}
	return ""
}

func osUser() (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}
	return u.Username, nil
}

// generateSSHKey creates an SSH keypair. It tries the preferred path first,
// falling back to /tmp/gcloud-go/ if the directory can't be created.
func generateSSHKey(preferredPath string) (string, error) {
	keyPath := preferredPath
	dir := filepath.Dir(keyPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		// Fall back to /tmp.
		dir = filepath.Join(os.TempDir(), "gcloud-go")
		if err := os.MkdirAll(dir, 0700); err != nil {
			return "", fmt.Errorf("creating SSH key directory: %w", err)
		}
		keyPath = filepath.Join(dir, "google_compute_engine")
	}

	fmt.Fprintf(os.Stderr, "WARNING: The SSH key file for gcloud does not exist.\n")
	fmt.Fprintf(os.Stderr, "WARNING: SSH keygen will be executed to generate a key.\n")

	cmd := exec.Command("ssh-keygen", "-t", "rsa", "-b", "3072", "-f", keyPath, "-N", "", "-q")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("running ssh-keygen: %w", err)
	}

	return keyPath, nil
}

// pushSSHKeyToProject adds the public key to the project's SSH metadata.
func pushSSHKeyToProject(ctx context.Context, svc *compute.Service, project, sshUser, pubKeyPath string) error {
	pubKey, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return fmt.Errorf("reading public key: %w", err)
	}

	// Format: "user:ssh-rsa AAAA... comment"
	entry := sshUser + ":" + strings.TrimSpace(string(pubKey))

	// Get current project metadata.
	proj, err := svc.Projects.Get(project).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting project metadata: %w", err)
	}

	// Find or create the ssh-keys metadata item.
	var sshKeysItem *compute.MetadataItems
	for _, item := range proj.CommonInstanceMetadata.Items {
		if item.Key == "ssh-keys" {
			sshKeysItem = item
			break
		}
	}

	if sshKeysItem == nil {
		sshKeysItem = &compute.MetadataItems{Key: "ssh-keys", Value: &entry}
		proj.CommonInstanceMetadata.Items = append(proj.CommonInstanceMetadata.Items, sshKeysItem)
	} else {
		// Append to existing keys if not already present.
		existing := ""
		if sshKeysItem.Value != nil {
			existing = *sshKeysItem.Value
		}
		if !strings.Contains(existing, strings.TrimSpace(string(pubKey))) {
			combined := existing + "\n" + entry
			sshKeysItem.Value = &combined
		}
	}

	fmt.Fprintf(os.Stderr, "Updating project ssh metadata...\n")
	_, err = svc.Projects.SetCommonInstanceMetadata(project, proj.CommonInstanceMetadata).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("updating project metadata: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Updating project ssh metadata...done.\n")
	return nil
}
