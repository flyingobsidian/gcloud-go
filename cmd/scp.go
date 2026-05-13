package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
)

var scpCmd = &cobra.Command{
	Use:   "scp [[USER@]INSTANCE:]SRC [[USER@]INSTANCE:]DST",
	Short: "Copy files to/from a Compute Engine instance via SCP",
	Args:  cobra.ExactArgs(2),
	RunE:  runSCP,
}

var (
	flagSCPTunnelThroughIAP bool
	flagSCPInternalIP       bool
	flagSCPKeyFile          string
	flagSCPRecurse          bool
)

func init() {
	scpCmd.Flags().BoolVar(&flagSCPTunnelThroughIAP, "tunnel-through-iap", false, "Tunnel through Identity-Aware Proxy")
	scpCmd.Flags().BoolVar(&flagSCPInternalIP, "internal-ip", false, "Connect using internal IP")
	scpCmd.Flags().StringVar(&flagSCPKeyFile, "ssh-key-file", "", "SSH private key file")
	scpCmd.Flags().BoolVar(&flagSCPRecurse, "recurse", false, "Upload directories recursively")

	computeCmd.AddCommand(scpCmd)
}

// scpTarget parses a SCP argument into optional user, instance, and path components.
// Format: [[USER@]INSTANCE:]PATH
type scpTarget struct {
	User     string
	Instance string
	Path     string
	IsRemote bool
}

func parseSCPTarget(arg string) scpTarget {
	// Check for colon (remote target indicator).
	colonIdx := strings.Index(arg, ":")
	if colonIdx < 0 {
		return scpTarget{Path: arg}
	}

	remote := arg[:colonIdx]
	path := arg[colonIdx+1:]
	user, instance := parseUserInstance(remote)
	return scpTarget{User: user, Instance: instance, Path: path, IsRemote: true}
}

func runSCP(cmd *cobra.Command, args []string) error {
	src := parseSCPTarget(args[0])
	dst := parseSCPTarget(args[1])

	// Determine which target is remote to resolve the instance.
	var remoteTarget *scpTarget
	if src.IsRemote {
		remoteTarget = &src
	} else if dst.IsRemote {
		remoteTarget = &dst
	} else {
		return fmt.Errorf("at least one argument must be a remote target (INSTANCE:PATH)")
	}

	project, zone, err := resolveProjectZone()
	if err != nil {
		return err
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	inst, err := svc.Instances.Get(project, zone, remoteTarget.Instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting instance %s: %w", remoteTarget.Instance, err)
	}

	// Build SCP args.
	scpArgs := []string{}
	if flagSCPRecurse {
		scpArgs = append(scpArgs, "-r")
	}

	// Use gcloud's default SSH key unless overridden.
	keyFile := flagSCPKeyFile
	if keyFile == "" {
		keyFile = googleSSHKeyPath()
	}
	if keyFile != "" {
		scpArgs = append(scpArgs, "-i", keyFile)
	}

	// Use gcloud's known_hosts file.
	if knownHosts := googleKnownHostsPath(); knownHosts != "" {
		scpArgs = append(scpArgs, "-o", "UserKnownHostsFile="+knownHosts)
	}

	scpArgs = append(scpArgs,
		"-o", "StrictHostKeyChecking=no",
		"-o", "CheckHostIP=no",
		"-o", "IdentitiesOnly=yes",
	)

	var host string

	if flagSCPTunnelThroughIAP {
		ln, err := startIAPTunnel(ctx, project, zone, remoteTarget.Instance)
		if err != nil {
			return err
		}
		defer ln.Close()

		localPort := ln.Addr().(*net.TCPAddr).Port
		scpArgs = append(scpArgs, "-P", strconv.Itoa(localPort))
		host = "localhost"
	} else if flagSCPInternalIP {
		host = getInternalIP(inst)
		if host == "" {
			return fmt.Errorf("instance %s has no internal IP", remoteTarget.Instance)
		}
	} else {
		host = getExternalIP(inst)
		if host == "" {
			return fmt.Errorf("instance %s has no external IP; consider --tunnel-through-iap", remoteTarget.Instance)
		}
	}

	// Build source and destination SCP arguments.
	scpArgs = append(scpArgs, formatSCPArg(src, host), formatSCPArg(dst, host))

	scpBin, err := exec.LookPath("scp")
	if err != nil {
		return fmt.Errorf("scp not found in PATH: %w", err)
	}

	c := exec.CommandContext(ctx, scpBin, scpArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return c.Run()
}

func formatSCPArg(t scpTarget, host string) string {
	if !t.IsRemote {
		return t.Path
	}
	prefix := host
	if t.User != "" {
		prefix = t.User + "@" + host
	}
	return prefix + ":" + t.Path
}
