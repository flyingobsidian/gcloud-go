package cmd

import (
	"fmt"
	"sync"
	"testing"
)

func TestOrganizationsListInProcess(t *testing.T) {
	inv, stdout, _ := newFakeAPIInvocation(t, fakeCRM(t))
	if err := runInvocation(inv, "organizations", "list"); err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("%-40s %-20s %s\n", "DISPLAY_NAME", "ID", "DIRECTORY_CUSTOMER_ID") +
		fmt.Sprintf("%-40s %-20s %s\n", "example.com", "20", "C01")
	if got := stdout.String(); got != want {
		t.Errorf("organizations list:\n%s\nwant:\n%s", got, want)
	}
}

func TestOrganizationsConcurrentInProcess(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Go(func() {
			inv, stdout, _ := newFakeAPIInvocation(t, fakeCRM(t))
			id := fmt.Sprint(1000 + i)
			if err := runInvocation(inv, "organizations", "describe", id); err != nil {
				t.Error(err)
				return
			}
			want := fmt.Sprintf("displayName: org-%s\nname: organizations/%s\n", id, id)
			if got := stdout.String(); got != want {
				t.Errorf("organizations describe %s = %q, want %q", id, got, want)
			}
		})
	}
	wg.Wait()
}
