package oslogin

import (
	"testing"

	"google.golang.org/api/compute/v1"
)

func strPtr(s string) *string { return &s }

func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name string
		inst *compute.Instance
		proj *compute.Project
		want bool
	}{
		{
			name: "no metadata",
			inst: &compute.Instance{},
			proj: &compute.Project{},
			want: false,
		},
		{
			name: "project metadata enables",
			inst: &compute.Instance{},
			proj: &compute.Project{
				CommonInstanceMetadata: &compute.Metadata{
					Items: []*compute.MetadataItems{
						{Key: "enable-oslogin", Value: strPtr("TRUE")},
					},
				},
			},
			want: true,
		},
		{
			name: "instance metadata enables",
			inst: &compute.Instance{
				Metadata: &compute.Metadata{
					Items: []*compute.MetadataItems{
						{Key: "enable-oslogin", Value: strPtr("true")},
					},
				},
			},
			proj: &compute.Project{},
			want: true,
		},
		{
			name: "instance overrides project to false",
			inst: &compute.Instance{
				Metadata: &compute.Metadata{
					Items: []*compute.MetadataItems{
						{Key: "enable-oslogin", Value: strPtr("false")},
					},
				},
			},
			proj: &compute.Project{
				CommonInstanceMetadata: &compute.Metadata{
					Items: []*compute.MetadataItems{
						{Key: "enable-oslogin", Value: strPtr("true")},
					},
				},
			},
			want: false,
		},
		{
			name: "instance overrides project to true",
			inst: &compute.Instance{
				Metadata: &compute.Metadata{
					Items: []*compute.MetadataItems{
						{Key: "enable-oslogin", Value: strPtr("true")},
					},
				},
			},
			proj: &compute.Project{
				CommonInstanceMetadata: &compute.Metadata{
					Items: []*compute.MetadataItems{
						{Key: "enable-oslogin", Value: strPtr("false")},
					},
				},
			},
			want: true,
		},
		{
			name: "case insensitive key",
			inst: &compute.Instance{
				Metadata: &compute.Metadata{
					Items: []*compute.MetadataItems{
						{Key: "Enable-OsLogin", Value: strPtr("True")},
					},
				},
			},
			proj: &compute.Project{},
			want: true,
		},
		{
			name: "nil project",
			inst: &compute.Instance{},
			proj: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEnabled(tt.inst, tt.proj)
			if got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}
