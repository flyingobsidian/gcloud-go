package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

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
	flagSSHTunnelThroughIAP bool
	flagSSHInternalIP       bool
	flagSSHKeyFile          string
)

func init() {
	sshCmd.Flags().BoolVar(&flagSSHTunnelThroughIAP, "tunnel-through-iap", false, "Tunnel through Identity-Aware Proxy")
	sshCmd.Flags().BoolVar(&flagSSHInternalIP, "internal-ip", false, "Connect using internal IP")
	sshCmd.Flags().StringVar(&flagSSHKeyFile, "ssh-key-file", "", "SSH private key file")

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

	// Build SSH args.
	sshArgs := buildSSHOpts(flagSSHKeyFile)

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

	// Append remote command if provided after --.
	if dashIdx := cmd.ArgsLenAtDash(); dashIdx >= 0 {
		sshArgs = append(sshArgs, "--")
		sshArgs = append(sshArgs, args[dashIdx:]...)
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
		opts = append(opts, "-i", keyFile)
	}

	// Use gcloud's known_hosts file.
	knownHosts := googleKnownHostsPath()
	if knownHosts != "" {
		opts = append(opts, "-o", "UserKnownHostsFile="+knownHosts)
	}

	opts = append(opts,
		"-o", "StrictHostKeyChecking=no",
		"-o", "CheckHostIP=no",
		"-o", "IdentitiesOnly=yes",
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
