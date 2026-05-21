package cmd

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/flyingobsidian/gcloud-go/internal/gcp"
	ioslogin "github.com/flyingobsidian/gcloud-go/internal/oslogin"
	"github.com/spf13/cobra"
)

var scpCmd = &cobra.Command{
	Use:   "scp [[USER@]INSTANCE:]SRC [[USER@]INSTANCE:]DST",
	Short: "Copy files to/from a Compute Engine instance via SCP",
	Args:  cobra.ExactArgs(2),
	RunE:  runSCP,
}

var (
	flagSCPTunnelThroughIAP      bool
	flagSCPInternalIP            bool
	flagSCPKeyFile               string
	flagSCPRecurse               bool
	flagSCPPort                  int
	flagSCPCompress              bool
	flagSCPFlag                  []string
	flagSCPPlain                 bool
	flagSCPStrictHostKeyChecking string
	flagSCPDryRun                bool
)

func init() {
	scpCmd.Flags().BoolVar(&flagSCPTunnelThroughIAP, "tunnel-through-iap", false, "Tunnel through Identity-Aware Proxy")
	scpCmd.Flags().BoolVar(&flagSCPInternalIP, "internal-ip", false, "Connect using internal IP")
	scpCmd.Flags().StringVar(&flagSCPKeyFile, "ssh-key-file", "", "SSH private key file")
	scpCmd.Flags().BoolVar(&flagSCPRecurse, "recurse", false, "Upload directories recursively")
	scpCmd.Flags().IntVar(&flagSCPPort, "port", 0, "SSH port on the remote host")
	scpCmd.Flags().BoolVar(&flagSCPCompress, "compress", false, "Enable compression")
	scpCmd.Flags().StringArrayVar(&flagSCPFlag, "scp-flag", nil, "Extra flags to pass to scp")
	scpCmd.Flags().BoolVar(&flagSCPPlain, "plain", false, "Suppress managed SSH key setup")
	scpCmd.Flags().StringVar(&flagSCPStrictHostKeyChecking, "strict-host-key-checking", "", "Override StrictHostKeyChecking (yes, no, ask)")
	scpCmd.Flags().BoolVar(&flagSCPDryRun, "dry-run", false, "Print the scp command without running it")

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
	if flagSCPTunnelThroughIAP && flagSCPInternalIP {
		return fmt.Errorf("--tunnel-through-iap and --internal-ip are mutually exclusive")
	}
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

	// Check OS Login and manage SSH keys (matching SSH command behavior).
	proj, err := svc.Projects.Get(project).Context(ctx).Do()
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: could not get project metadata: %v\n", err)
	}
	useOSLogin := proj != nil && ioslogin.IsEnabled(inst, proj)

	if !flagSCPPlain && flagSCPKeyFile == "" {
		keyPath := googleSSHKeyPath()
		keyGenerated := false
		if _, err := os.Stat(keyPath); os.IsNotExist(err) {
			keyPath, err = generateSSHKey(keyPath)
			if err != nil {
				return fmt.Errorf("generating SSH key: %w", err)
			}
			keyGenerated = true
		}

		if keyGenerated {
			if useOSLogin {
				email := resolveAccountEmail()
				if email != "" {
					osLoginSvc, err := gcp.OSLoginService(ctx, flagAccount)
					if err != nil {
						fmt.Fprintf(os.Stderr, "WARNING: could not create OS Login service: %v\n", err)
					} else {
						posixUser, err := ioslogin.ImportSSHKey(ctx, osLoginSvc, email, keyPath+".pub", project)
						if err != nil {
							fmt.Fprintf(os.Stderr, "WARNING: could not import SSH key via OS Login: %v\n", err)
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
					if err := pushSSHKeyToProject(ctx, svc, project, sshUser, keyPath+".pub"); err != nil {
						fmt.Fprintf(os.Stderr, "WARNING: could not push SSH key to project metadata: %v\n", err)
					} else {
						fmt.Fprintln(os.Stderr, "Waiting for SSH key to propagate.")
						time.Sleep(5 * time.Second)
					}
				}
			}
		} else if useOSLogin && remoteTarget.User == "" {
			email := resolveAccountEmail()
			if email != "" {
				osLoginSvc, err := gcp.OSLoginService(ctx, flagAccount)
				if err != nil {
					fmt.Fprintf(os.Stderr, "WARNING: could not create OS Login service: %v\n", err)
				} else {
					posixUser, err := ioslogin.PosixUsername(ctx, osLoginSvc, email, project)
					if err != nil {
						fmt.Fprintf(os.Stderr, "WARNING: could not resolve OS Login username: %v\n", err)
					} else {
						remoteTarget.User = posixUser
					}
				}
			}
		}
	} else if useOSLogin && remoteTarget.User == "" && !flagSCPPlain {
		email := resolveAccountEmail()
		if email != "" {
			osLoginSvc, err := gcp.OSLoginService(ctx, flagAccount)
			if err != nil {
				fmt.Fprintf(os.Stderr, "WARNING: could not create OS Login service: %v\n", err)
			} else {
				posixUser, err := ioslogin.PosixUsername(ctx, osLoginSvc, email, project)
				if err != nil {
					fmt.Fprintf(os.Stderr, "WARNING: could not resolve OS Login username: %v\n", err)
				} else {
					remoteTarget.User = posixUser
				}
			}
		}
	}

	// Resolve the target host IP first so we can check known_hosts.
	var host string
	var iapListener net.Listener

	if flagSCPTunnelThroughIAP {
		ln, err := startIAPTunnel(ctx, project, zone, remoteTarget.Instance)
		if err != nil {
			return err
		}
		defer ln.Close()
		iapListener = ln
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

	// Build SCP args.
	scpArgs := []string{}
	if flagSCPRecurse {
		scpArgs = append(scpArgs, "-r")
	}

	if flagSCPPlain {
		// --plain: no managed key setup, just minimal args.
	} else {
		// Use gcloud's default SSH key unless overridden.
		keyFile := flagSCPKeyFile
		if keyFile == "" {
			keyFile = googleSSHKeyPath()
		}
		if keyFile != "" {
			if _, err := os.Stat(keyFile); err == nil {
				scpArgs = append(scpArgs, "-i", keyFile, "-o", "IdentitiesOnly=yes")
			}
		}

		// Use gcloud's known_hosts file.
		if knownHosts := googleKnownHostsPath(); knownHosts != "" {
			scpArgs = append(scpArgs, "-o", "UserKnownHostsFile="+knownHosts)
		}

		// Match gcloud Python behavior: verify host key if host is already known.
		if knownHostsHasHost(host) {
			scpArgs = append(scpArgs, "-o", "StrictHostKeyChecking=yes")
		} else {
			scpArgs = append(scpArgs, "-o", "StrictHostKeyChecking=no")
		}
		scpArgs = append(scpArgs, "-o", "CheckHostIP=no")
	}

	if flagSCPStrictHostKeyChecking != "" {
		scpArgs = append(scpArgs, "-o", "StrictHostKeyChecking="+flagSCPStrictHostKeyChecking)
	}

	if flagSCPCompress {
		scpArgs = append(scpArgs, "-C")
	}

	for _, f := range flagSCPFlag {
		scpArgs = append(scpArgs, f)
	}

	if iapListener != nil {
		localPort := iapListener.Addr().(*net.TCPAddr).Port
		scpArgs = append(scpArgs, "-P", strconv.Itoa(localPort))
	}

	if flagSCPPort != 0 && !flagSCPTunnelThroughIAP {
		scpArgs = append(scpArgs, "-P", strconv.Itoa(flagSCPPort))
	}

	// Build source and destination SCP arguments.
	scpArgs = append(scpArgs, formatSCPArg(src, host), formatSCPArg(dst, host))

	if flagSCPDryRun {
		fmt.Println(shellJoin("scp", scpArgs))
		return nil
	}

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
