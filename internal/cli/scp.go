package cli

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"

	ioslogin "github.com/flyingobsidian/gcloud-go/internal/oslogin"
	"github.com/spf13/cobra"
)

func (s *computeCmds) initScp() {
	s.scpCmd.Flags().BoolVar(&s.flagSCPTunnelThroughIAP, "tunnel-through-iap", false, "Tunnel through Identity-Aware Proxy")
	s.scpCmd.Flags().BoolVar(&s.flagSCPInternalIP, "internal-ip", false, "Connect using internal IP")
	s.scpCmd.Flags().StringVar(&s.flagSCPKeyFile, "ssh-key-file", "", "SSH private key file")
	s.scpCmd.Flags().BoolVar(&s.flagSCPRecurse, "recurse", false, "Upload directories recursively")
	s.scpCmd.Flags().IntVar(&s.flagSCPPort, "port", 0, "SSH port on the remote host")
	s.scpCmd.Flags().BoolVar(&s.flagSCPCompress, "compress", false, "Enable compression")
	s.scpCmd.Flags().StringArrayVar(&s.flagSCPFlag, "scp-flag", nil, "Extra flags to pass to scp")
	s.scpCmd.Flags().BoolVar(&s.flagSCPPlain, "plain", false, "Suppress managed SSH key setup")
	s.scpCmd.Flags().StringVar(&s.flagSCPStrictHostKeyChecking, "strict-host-key-checking", "", "Override StrictHostKeyChecking (yes, no, ask)")
	s.scpCmd.Flags().BoolVar(&s.flagSCPDryRun, "dry-run", false, "Print the scp command without running it")

	s.computeCmd.AddCommand(s.scpCmd)
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

// parseSCPArgs splits SRC... DEST into targets and checks that they form a
// valid copy, matching gcloud's SCPCommand.Verify with single_remote=True:
// either local sources to a remote destination, or sources all on the same
// remote to a local destination.
func parseSCPArgs(args []string) ([]scpTarget, scpTarget, error) {
	if len(args) < 2 {
		return nil, scpTarget{}, fmt.Errorf("at least one source and a destination are required")
	}
	srcs := make([]scpTarget, 0, len(args)-1)
	for _, a := range args[:len(args)-1] {
		srcs = append(srcs, parseSCPTarget(a))
	}
	dst := parseSCPTarget(args[len(args)-1])

	if dst.IsRemote {
		for _, src := range srcs {
			if src.IsRemote {
				return nil, scpTarget{}, fmt.Errorf("all sources must be local files when destination is remote")
			}
		}
		return srcs, dst, nil
	}
	for _, src := range srcs {
		if !src.IsRemote {
			return nil, scpTarget{}, fmt.Errorf("source(s) must be remote when destination is local")
		}
		if src.User != srcs[0].User || src.Instance != srcs[0].Instance {
			return nil, scpTarget{}, fmt.Errorf("all sources must refer to the same remote when destination is local")
		}
	}
	return srcs, dst, nil
}

func (s *computeCmds) runSCP(cmd *cobra.Command, args []string) error {
	if s.flagSCPTunnelThroughIAP && s.flagSCPInternalIP {
		return fmt.Errorf("--tunnel-through-iap and --internal-ip are mutually exclusive")
	}
	srcs, dst, err := parseSCPArgs(args)
	if err != nil {
		return err
	}

	// Determine which target is remote to resolve the instance; all remote
	// targets share the same user and instance.
	remoteTarget := &dst
	if !dst.IsRemote {
		remoteTarget = &srcs[0]
	}

	project, zone, err := s.inv.projectZone()
	if err != nil {
		return err
	}

	ctx := cmd.Context()
	svc, err := s.inv.clients().Compute(ctx)
	if err != nil {
		return err
	}

	inst, err := svc.Instances.Get(project, zone, remoteTarget.Instance).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("getting instance %s: %w", remoteTarget.Instance, err)
	}

	// Check OS Login and manage SSH keys (matching SSH command behavior).
	proj, err := svc.Projects.Get(project).Context(ctx).Do()
	if err != nil {
		fmt.Fprintf(s.inv.stderr, "WARNING: could not get project metadata: %v\n", err)
	}
	useOSLogin := proj != nil && ioslogin.IsEnabled(inst, proj)

	if !s.flagSCPPlain && s.flagSCPKeyFile == "" {
		keyPath := s.inv.googleSSHKeyPath()
		keyGenerated := false
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			keyPath, err = s.generateSSHKey(keyPath)
			if err != nil {
				return fmt.Errorf("generating SSH key: %w", err)
			}
			keyGenerated = true
		}

		if keyGenerated {
			if useOSLogin {
				email := s.resolveAccountEmail()
				if email != "" {
					osLoginSvc, err := s.inv.clients().OSLogin(ctx)
					if err != nil {
						fmt.Fprintf(s.inv.stderr, "WARNING: could not create OS Login service: %v\n", err)
					} else {
						posixUser, err := ioslogin.ImportSSHKey(ctx, osLoginSvc, email, keyPath+".pub", project)
						if err != nil {
							fmt.Fprintf(s.inv.stderr, "WARNING: could not import SSH key via OS Login: %v\n", err)
						} else if posixUser != "" && remoteTarget.User == "" {
							remoteTarget.User = posixUser
						}
					}
				}
			} else {
				sshUser := remoteTarget.User
				if sshUser == "" {
					if u, err := user.Current(); err == nil {
						sshUser = u.Username
					}
				}
				if sshUser != "" {
					if err := s.pushSSHKeyToProject(ctx, svc, project, sshUser, keyPath+".pub"); err != nil {
						fmt.Fprintf(s.inv.stderr, "WARNING: could not push SSH key to project metadata: %v\n", err)
					} else {
						fmt.Fprintln(s.inv.stderr, "Waiting for SSH key to propagate.")
						time.Sleep(5 * time.Second)
					}
				}
			}
		} else if useOSLogin && remoteTarget.User == "" {
			email := s.resolveAccountEmail()
			if email != "" {
				osLoginSvc, err := s.inv.clients().OSLogin(ctx)
				if err != nil {
					fmt.Fprintf(s.inv.stderr, "WARNING: could not create OS Login service: %v\n", err)
				} else {
					posixUser, err := ioslogin.PosixUsername(ctx, osLoginSvc, email, project)
					if err != nil {
						fmt.Fprintf(s.inv.stderr, "WARNING: could not resolve OS Login username: %v\n", err)
					} else {
						remoteTarget.User = posixUser
					}
				}
			}
		}
	} else if useOSLogin && remoteTarget.User == "" && !s.flagSCPPlain {
		email := s.resolveAccountEmail()
		if email != "" {
			osLoginSvc, err := s.inv.clients().OSLogin(ctx)
			if err != nil {
				fmt.Fprintf(s.inv.stderr, "WARNING: could not create OS Login service: %v\n", err)
			} else {
				posixUser, err := ioslogin.PosixUsername(ctx, osLoginSvc, email, project)
				if err != nil {
					fmt.Fprintf(s.inv.stderr, "WARNING: could not resolve OS Login username: %v\n", err)
				} else {
					remoteTarget.User = posixUser
				}
			}
		}
	}

	// Resolve the target host IP first so we can check known_hosts.
	var host string
	var iapListener net.Listener

	if s.flagSCPTunnelThroughIAP {
		ln, err := s.startIAPTunnel(ctx, project, zone, remoteTarget.Instance)
		if err != nil {
			return err
		}
		defer ln.Close()
		iapListener = ln
		host = "localhost"
	} else if s.flagSCPInternalIP {
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

	// Build SCP args.
	scpArgs := []string{}
	if s.flagSCPRecurse {
		scpArgs = append(scpArgs, "-r")
	}

	if s.flagSCPPlain {
		// --plain: no managed key setup, just minimal args.
	} else {
		// Use gcloud's default SSH key unless overridden.
		keyFile := s.flagSCPKeyFile
		if keyFile == "" {
			keyFile = s.inv.googleSSHKeyPath()
		}
		if keyFile != "" {
			if _, err := os.Stat(keyFile); err == nil {
				scpArgs = append(scpArgs, "-i", keyFile, "-o", "IdentitiesOnly=yes")
			}
		}

		// Use gcloud's known_hosts file.
		if knownHosts := s.inv.googleKnownHostsPath(); knownHosts != "" {
			scpArgs = append(scpArgs, "-o", "UserKnownHostsFile="+knownHosts)
		}

		// Match gcloud Python behavior: verify host key if host is already known.
		if s.inv.knownHostsHasHost(host) {
			scpArgs = append(scpArgs, "-o", "StrictHostKeyChecking=yes")
		} else {
			scpArgs = append(scpArgs, "-o", "StrictHostKeyChecking=no")
		}
		scpArgs = append(scpArgs, "-o", "CheckHostIP=no")
	}

	if s.flagSCPStrictHostKeyChecking != "" {
		scpArgs = append(scpArgs, "-o", "StrictHostKeyChecking="+s.flagSCPStrictHostKeyChecking)
	}

	if s.flagSCPCompress {
		scpArgs = append(scpArgs, "-C")
	}

	for _, f := range s.flagSCPFlag {
		scpArgs = append(scpArgs, f)
	}

	if iapListener != nil {
		localPort := iapListener.Addr().(*net.TCPAddr).Port
		scpArgs = append(scpArgs, "-P", strconv.Itoa(localPort))
	}

	if s.flagSCPPort != 0 && !s.flagSCPTunnelThroughIAP {
		scpArgs = append(scpArgs, "-P", strconv.Itoa(s.flagSCPPort))
	}

	// Build source and destination SCP arguments.
	scpArgs = append(scpArgs, scpTargetArgs(srcs, dst, remoteTarget.User, host)...)

	if s.flagSCPDryRun {
		fmt.Fprintln(s.inv.stdout, shellJoin("scp", scpArgs))
		return nil
	}

	scpBin, err := exec.LookPath("scp")
	if err != nil {
		return fmt.Errorf("scp not found in PATH: %w", err)
	}

	c := exec.CommandContext(ctx, scpBin, scpArgs...)
	c.Stdin = s.inv.stdin
	c.Stdout = s.inv.stdout
	c.Stderr = s.inv.stderr
	return c.Run()
}

// scpTargetArgs formats the sources and destination for scp, applying the
// resolved remote user (which may come from OS Login) to every remote target.
func scpTargetArgs(srcs []scpTarget, dst scpTarget, remoteUser, host string) []string {
	targets := append(append([]scpTarget{}, srcs...), dst)
	out := make([]string, 0, len(targets))
	for _, t := range targets {
		if t.IsRemote {
			t.User = remoteUser
		}
		out = append(out, formatSCPArg(t, host))
	}
	return out
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
