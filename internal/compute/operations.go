package compute

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/compute/v1"
)

const (
	pollInterval = 2 * time.Second
	maxWait      = 30 * time.Minute
)

// WaitForZoneOp polls a zone operation until it completes or times out.
func WaitForZoneOp(ctx context.Context, svc *compute.Service, project, zone, opName string) error {
	deadline := time.Now().Add(maxWait)
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("operation %s timed out after %v", opName, maxWait)
		}

		op, err := svc.ZoneOperations.Get(project, zone, opName).Context(ctx).Do()
		if err != nil {
			return fmt.Errorf("polling operation %s: %w", opName, err)
		}

		if op.Status == "DONE" {
			if op.Error != nil && len(op.Error.Errors) > 0 {
				return fmt.Errorf("operation %s failed: %s", opName, op.Error.Errors[0].Message)
			}
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}
