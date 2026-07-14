package cmd

import (
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestArtifactsYumSubcommandsRegistered(t *testing.T) {
	var yum *cobra.Command
	for _, c := range artifactsCmd.Commands() {
		if c.Name() == "yum" {
			yum = c
			break
		}
	}
	if yum == nil {
		t.Fatal("artifacts yum not registered")
	}
	want := []string{"import", "upload"}
	got := map[string]bool{}
	for _, c := range yum.Commands() {
		got[c.Name()] = true
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("artifacts yum %s not registered", name)
		}
	}
}

func TestExpandYumGcsSources(t *testing.T) {
	tests := []struct {
		name         string
		in           []string
		wantURIs     []string
		wantWildcard bool
	}{
		{"single", []string{"gs://b/p.rpm"}, []string{"gs://b/p.rpm"}, false},
		{"comma-separated", []string{"gs://b/a.rpm,gs://b/b.rpm"}, []string{"gs://b/a.rpm", "gs://b/b.rpm"}, false},
		{"wildcard sets flag", []string{"gs://b/dir/*"}, []string{"gs://b/dir/*"}, true},
		{"empty entries dropped", []string{"gs://b/x.rpm,,", ""}, []string{"gs://b/x.rpm"}, false},
		{"trims whitespace", []string{"gs://b/a.rpm ,  gs://b/b.rpm"}, []string{"gs://b/a.rpm", "gs://b/b.rpm"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uris, w := expandYumGcsSources(tt.in)
			if !reflect.DeepEqual(uris, tt.wantURIs) {
				t.Errorf("uris = %v, want %v", uris, tt.wantURIs)
			}
			if w != tt.wantWildcard {
				t.Errorf("useWildcards = %v, want %v", w, tt.wantWildcard)
			}
		})
	}
}

func TestArtifactsYumImportRequiresGcsSource(t *testing.T) {
	// Cobra registers the required flag; verify the annotation is set so a bare
	// `artifacts yum import` correctly errors out at parse time.
	f := artifactsYumImportCmd.Flag("gcs-source")
	if f == nil {
		t.Fatal("--gcs-source flag not registered on artifacts yum import")
	}
	req := f.Annotations[cobra.BashCompOneRequiredFlag]
	if len(req) == 0 || req[0] != "true" {
		t.Errorf("--gcs-source not marked required (annotations=%v)", f.Annotations)
	}
}

func TestArtifactsYumUploadRequiresSource(t *testing.T) {
	f := artifactsYumUploadCmd.Flag("source")
	if f == nil {
		t.Fatal("--source flag not registered on artifacts yum upload")
	}
	req := f.Annotations[cobra.BashCompOneRequiredFlag]
	if len(req) == 0 || req[0] != "true" {
		t.Errorf("--source not marked required (annotations=%v)", f.Annotations)
	}
}
