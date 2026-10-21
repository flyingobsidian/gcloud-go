package cmd

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	cloudscheduler "google.golang.org/api/cloudscheduler/v1"
	"google.golang.org/api/googleapi"
)

func TestSchedulerAppLocation(t *testing.T) {
	loc := func(region string) *cloudscheduler.Location {
		return &cloudscheduler.Location{LocationId: region, Labels: map[string]string{schedulerRegionLabel: region}}
	}
	cases := []struct {
		name        string
		appExists   bool
		locations   []*cloudscheduler.Location
		listErr     error
		want        string
		wantErr     error
		wantWarning bool
	}{
		{name: "no app", wantErr: errSchedulerNoLocation},
		{name: "single location", appExists: true, locations: []*cloudscheduler.Location{loc("europe-west2")}, want: "europe-west2"},
		{name: "several locations warn", appExists: true,
			locations: []*cloudscheduler.Location{loc("europe-west2"), loc("us-central1")}, want: "europe-west2", wantWarning: true},
		{name: "location id without label", appExists: true,
			locations: []*cloudscheduler.Location{{LocationId: "asia-east1"}}, want: "asia-east1"},
		{name: "no locations", appExists: true, wantErr: errSchedulerNoLocation},
		{name: "not found", appExists: true, listErr: &googleapi.Error{Code: http.StatusNotFound}, wantErr: errSchedulerNoLocation},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			listed := false
			var buf bytes.Buffer
			got, err := schedulerAppLocation(context.Background(), &buf,
				func(context.Context) bool { return tc.appExists },
				func(context.Context) ([]*cloudscheduler.Location, error) {
					listed = true
					return tc.locations, tc.listErr
				})
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("err = %v, want %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Errorf("location = %q, want %q", got, tc.want)
			}
			if listed != tc.appExists {
				t.Errorf("listed locations = %v, want %v", listed, tc.appExists)
			}
			if gotWarning := strings.Contains(buf.String(), "We are using the App Engine app location (europe-west2)"); gotWarning != tc.wantWarning {
				t.Errorf("warning output %q, want warning %v", buf.String(), tc.wantWarning)
			}
		})
	}
}

func TestSchedulerAppLocationListError(t *testing.T) {
	_, err := schedulerAppLocation(context.Background(), &bytes.Buffer{},
		func(context.Context) bool { return true },
		func(context.Context) ([]*cloudscheduler.Location, error) { return nil, errors.New("boom") })
	if err == nil || errors.Is(err, errSchedulerNoLocation) || !strings.Contains(err.Error(), "boom") {
		t.Errorf("err = %v, want wrapped list error", err)
	}
}
