package cmd

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	compute "google.golang.org/api/compute/v1"
)

func TestComputeImageProjects(t *testing.T) {
	if got := computeImageProjects("mine", false, false); strings.Join(got, ",") != "mine" {
		t.Errorf("--no-standard-images: got %v, want only the project", got)
	}
	got := computeImageProjects("mine", true, false)
	if got[0] != "mine" || len(got) != 1+len(computePublicImageProjects) {
		t.Fatalf("standard: got %v", got)
	}
	if got[1] != "almalinux-cloud" || got[len(got)-1] != "windows-sql-cloud" {
		t.Errorf("public projects not sorted: %v", got)
	}
	if dup := computeImageProjects("debian-cloud", true, false); len(dup) != len(computePublicImageProjects) {
		t.Errorf("project listed twice when it is a public image project: %v", dup)
	}
}

func TestListComputeImages(t *testing.T) {
	images := map[string][]*compute.Image{
		"mine": {
			{Name: "active", Deprecated: &compute.DeprecationStatus{State: "ACTIVE"}},
			{Name: "plain"},
			{Name: "old", Deprecated: &compute.DeprecationStatus{State: "DEPRECATED"}},
		},
		"debian-cloud": {{Name: "debian-13"}},
	}
	var listed []string
	list := func(_ context.Context, project string) ([]*compute.Image, error) {
		listed = append(listed, project)
		return images[project], nil
	}
	names := func(imgs []*compute.Image) string {
		var n []string
		for _, i := range imgs {
			n = append(n, i.Name)
		}
		return strings.Join(n, ",")
	}

	got, err := listComputeImages(context.Background(), []string{"mine", "debian-cloud"}, false, list)
	if err != nil {
		t.Fatal(err)
	}
	if names(got) != "active,plain,debian-13" || strings.Join(listed, ",") != "mine,debian-cloud" {
		t.Errorf("got %s from %v", names(got), listed)
	}
	got, err = listComputeImages(context.Background(), []string{"mine"}, true, list)
	if err != nil || names(got) != "active,plain,old" {
		t.Errorf("--show-deprecated: got %s, %v", names(got), err)
	}

	fail := func(context.Context, string) ([]*compute.Image, error) { return nil, errors.New("boom") }
	if _, err := listComputeImages(context.Background(), []string{"mine"}, false, fail); err == nil || !strings.Contains(err.Error(), "project mine") {
		t.Errorf("err = %v", err)
	}
}

func TestComputeImagesDefaultTable(t *testing.T) {
	imgs := []*compute.Image{{
		Name: "app-v1", Family: "app", Status: "READY",
		SelfLink: "https://www.googleapis.com/compute/v1/projects/mine/global/images/app-v1",
	}}
	var out bytes.Buffer
	if err := imagesLister.emit(&out, &bytes.Buffer{}, imgs, "", false); err != nil {
		t.Fatal(err)
	}
	want := "NAME    PROJECT  FAMILY  DEPRECATED  STATUS\napp-v1  mine     app                 READY\n"
	if out.String() != want {
		t.Errorf("table =\n%q\nwant\n%q", out.String(), want)
	}
}

func TestComputeImagesListFlags(t *testing.T) {
	list, _, err := computeCmd.Find([]string{"images", "list"})
	if err != nil || list.RunE == nil {
		t.Fatalf("compute images list not registered: %v", err)
	}
	if err := list.ParseFlags([]string{"--no-standard-images"}); err != nil {
		t.Fatalf("--no-standard-images rejected: %v", err)
	}
	defer func() {
		flagImagesListNoStandard = false
		list.Flags().Lookup("no-standard-images").Changed = false
	}()
	if !flagImagesListNoStandard {
		t.Error("--no-standard-images not set")
	}
}
