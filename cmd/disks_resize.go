package cmd

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	icompute "github.com/flyingobsidian/gcloud-go/internal/compute"
	"github.com/spf13/cobra"
	"google.golang.org/api/compute/v1"
)

// --- gcloud compute disks resize (#1727) ---
//
// Grows one or more persistent disks in place (shrinking is not supported by
// the Compute API). Matches gcloud-python's surface:
//
//	gcloud compute disks resize DISK_NAME [DISK_NAME ...] \
//	    --size=SIZE (--zone=ZONE | --region=REGION)
//
// SIZE is a whole number optionally suffixed with GB or TB (default GB).

var (
	flagDiskResizeSize   string
	flagDiskResizeZone   string
	flagDiskResizeRegion string
	flagDiskResizeAsync  bool
	flagDiskResizeFormat string
)

var disksResizeCmd = &cobra.Command{
	Use:   "resize DISK_NAME [DISK_NAME ...]",
	Short: "Resize (grow) one or more Compute Engine disks",
	Long: "Resize one or more Compute Engine persistent disks. Disk size may " +
		"only be increased; shrinking is not supported by the Compute API. " +
		"This change is not reversible.",
	Args: cobra.MinimumNArgs(1),
	RunE: runDisksResize,
}

func init() {
	disksResizeCmd.Flags().StringVar(&flagDiskResizeSize, "size", "",
		"New size for the disks. Whole number + optional GB or TB suffix (default GB). (required)")
	_ = disksResizeCmd.MarkFlagRequired("size")
	disksResizeCmd.Flags().StringVar(&flagDiskResizeZone, "zone", "",
		"Zone of the disks (mutually exclusive with --region)")
	disksResizeCmd.Flags().StringVar(&flagDiskResizeRegion, "region", "",
		"Region of the disks (mutually exclusive with --zone)")
	disksResizeCmd.Flags().BoolVar(&flagDiskResizeAsync, "async", false,
		"Return immediately without waiting for the resize operation to complete")
	disksResizeCmd.Flags().StringVar(&flagDiskResizeFormat, "format", "",
		"Output format (default yaml; use 'none' to suppress the resized-disk dump)")
	disksCmd.AddCommand(disksResizeCmd)
}

// parseDiskSizeGB converts a gcloud disk-size string like "200", "200GB", or
// "1TB" to gigabytes as an int64. The input is case-insensitive; whitespace
// and an optional leading "+" are tolerated.
func parseDiskSizeGB(size string) (int64, error) {
	s := strings.TrimSpace(strings.ToUpper(size))
	s = strings.TrimPrefix(s, "+")
	if s == "" {
		return 0, fmt.Errorf("--size is required")
	}
	mult := int64(1)
	switch {
	case strings.HasSuffix(s, "TB"):
		s = strings.TrimSuffix(s, "TB")
		mult = 1024
	case strings.HasSuffix(s, "GB"):
		s = strings.TrimSuffix(s, "GB")
	}
	s = strings.TrimSpace(s)
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid --size %q: %w", size, err)
	}
	if n <= 0 {
		return 0, fmt.Errorf("--size must be positive, got %q", size)
	}
	return n * mult, nil
}

func runDisksResize(cmd *cobra.Command, args []string) error {
	if flagDiskResizeZone != "" && flagDiskResizeRegion != "" {
		return fmt.Errorf("--zone and --region are mutually exclusive")
	}

	sizeGB, err := parseDiskSizeGB(flagDiskResizeSize)
	if err != nil {
		return err
	}

	project, err := resolveProject()
	if err != nil {
		return err
	}

	zone := flagDiskResizeZone
	if zone == "" && flagDiskResizeRegion == "" {
		// Match gcloud-python fallback: honour compute/zone from config /
		// environment.
		zone = resolveZone()
	}
	if zone == "" && flagDiskResizeRegion == "" {
		return fmt.Errorf("one of --zone or --region is required")
	}

	if !flagQuiet {
		fmt.Fprintln(os.Stderr, "This command increases disk size. This change is not reversible.")
		fmt.Fprintln(os.Stderr, "For more information, see:")
		fmt.Fprintln(os.Stderr, "https://cloud.google.com/sdk/gcloud/reference/compute/disks/resize")
		fmt.Fprintln(os.Stderr)
		fmt.Fprint(os.Stderr, "Do you want to continue (Y/n)? ")
		var answer string
		fmt.Scanln(&answer)
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "" && answer != "y" && answer != "yes" {
			fmt.Fprintln(os.Stderr, "Aborted.")
			return nil
		}
	}

	ctx := context.Background()
	svc, err := icompute.NewService(ctx, flagAccount)
	if err != nil {
		return err
	}

	format := flagDiskResizeFormat
	if format == "" {
		format = "yaml"
	}

	for _, disk := range args {
		if err := resizeOneDisk(ctx, svc, project, zone, flagDiskResizeRegion, disk, sizeGB, format); err != nil {
			return err
		}
	}
	return nil
}

func resizeOneDisk(ctx context.Context, svc *compute.Service, project, zone, region, disk string, sizeGB int64, format string) error {
	if zone != "" {
		req := &compute.DisksResizeRequest{SizeGb: sizeGB}
		op, err := svc.Disks.Resize(project, zone, disk, req).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("resizing disk %s: %w", disk, err)
		}
		if !flagDiskResizeAsync {
			if err := icompute.WaitForZoneOp(ctx, svc, project, zone, op.Name); err != nil {
				return err
			}
		}
		return emitResizedDisk(ctx, svc, project, zone, "", disk, format)
	}

	// regional disk
	req := &compute.RegionDisksResizeRequest{SizeGb: sizeGB}
	op, err := svc.RegionDisks.Resize(project, region, disk, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("resizing regional disk %s: %w", disk, err)
	}
	if !flagDiskResizeAsync {
		if err := waitForRegionOp(ctx, svc, project, region, op.Name); err != nil {
			return err
		}
	}
	return emitResizedDisk(ctx, svc, project, "", region, disk, format)
}

func emitResizedDisk(ctx context.Context, svc *compute.Service, project, zone, region, disk, format string) error {
	scope := "zones/" + zone
	if zone == "" {
		scope = "regions/" + region
	}
	self := fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/%s/disks/%s", project, scope, path.Base(disk))
	fmt.Fprintf(os.Stderr, "Updated [%s].\n", self)

	if strings.EqualFold(format, "none") {
		return nil
	}
	var (
		got *compute.Disk
		err error
	)
	if zone != "" {
		got, err = svc.Disks.Get(project, zone, disk).Context(ctx).Do()
	} else {
		got, err = svc.RegionDisks.Get(project, region, disk).Context(ctx).Do()
	}
	if err != nil {
		return fmt.Errorf("re-reading disk after resize: %w", err)
	}
	return emitFormatted(got, format)
}
