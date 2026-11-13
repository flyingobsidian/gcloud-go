package cli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
)

// fakeScheduler serves the parts of the Cloud Scheduler and App Engine APIs
// the scheduler commands use. Every project has an App Engine app in
// europe-west1.
func fakeScheduler(t *testing.T) http.Handler {
	t.Helper()
	write := func(w http.ResponseWriter, v any) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(v)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer fake-token" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		p := strings.TrimPrefix(r.URL.Path, "/v1/")
		switch parts := strings.Split(p, "/"); {
		case len(parts) == 2 && parts[0] == "apps":
			write(w, map[string]any{"id": parts[1], "locationId": "europe-west"})
		case len(parts) == 3 && parts[2] == "locations":
			write(w, map[string]any{"locations": []map[string]any{
				{"name": p + "/europe-west1", "locationId": "europe-west1",
					"labels": map[string]string{"cloud.googleapis.com/region": "europe-west1"}},
			}})
		case len(parts) == 5 && parts[4] == "jobs":
			write(w, map[string]any{"jobs": []map[string]any{
				{"name": p + "/nightly", "schedule": "0 3 * * *", "timeZone": "Etc/UTC",
					"httpTarget": map[string]any{"uri": "https://example.com"}, "state": "ENABLED"},
			}})
		case len(parts) == 6 && parts[4] == "jobs":
			write(w, map[string]any{"name": p, "schedule": "*/5 * * * *", "state": "PAUSED"})
		default:
			http.NotFound(w, r)
		}
	})
}

func TestSchedulerJobsListInProcess(t *testing.T) {
	inv, stdout, stderr := newFakeAPIInvocation(t, fakeScheduler(t))
	// No --location: falls back to the App Engine app's location.
	if err := runInvocation(inv, "scheduler", "jobs", "list", "--project=p1"); err != nil {
		t.Fatal(err)
	}
	want := "ID       LOCATION      SCHEDULE (TZ)        TARGET_TYPE  STATE\n" +
		"nightly  europe-west1  0 3 * * * (Etc/UTC)  HTTP         ENABLED\n"
	if got := stdout.String(); got != want {
		t.Errorf("scheduler jobs list:\n%q\nwant:\n%q", got, want)
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty", stderr.String())
	}

	stdout.Reset()
	if err := runInvocation(inv, "scheduler", "jobs", "list", "--project=p1", "--location=us-central1", "--uri"); err != nil {
		t.Fatal(err)
	}
	if want := "projects/p1/locations/us-central1/jobs/nightly\n"; stdout.String() != want {
		t.Errorf("scheduler jobs list --uri = %q, want %q", stdout.String(), want)
	}
}

func TestSchedulerConcurrentInProcess(t *testing.T) {
	var wg sync.WaitGroup
	for i := range 10 {
		wg.Go(func() {
			inv, stdout, _ := newFakeAPIInvocation(t, fakeScheduler(t))
			project := fmt.Sprintf("p%d", i)
			if err := runInvocation(inv, "scheduler", "jobs", "describe", "job", "--project="+project, "--location=l1", "--format=json"); err != nil {
				t.Error(err)
				return
			}
			var job struct{ Name string }
			if err := json.Unmarshal(stdout.Bytes(), &job); err != nil {
				t.Errorf("describe output is not JSON: %v\n%s", err, stdout.String())
				return
			}
			if want := "projects/" + project + "/locations/l1/jobs/job"; job.Name != want {
				t.Errorf("describe name = %q, want %q", job.Name, want)
			}
		})
	}
	wg.Wait()
}

func TestSchedulerJobsDescribeDefaultsToYAML(t *testing.T) {
	inv, stdout, _ := newFakeAPIInvocation(t, fakeScheduler(t))
	if err := runInvocation(inv, "scheduler", "jobs", "describe", "job", "--project=p1", "--location=l1"); err != nil {
		t.Fatal(err)
	}
	want := "name: projects/p1/locations/l1/jobs/job\nschedule: '*/5 * * * *'\nstate: PAUSED\n"
	if got := stdout.String(); got != want {
		t.Errorf("describe = %q, want %q", got, want)
	}
}

// Guards #2081: create sends the --time-zone and --http-method defaults even
// though update binds the same flags with empty defaults.
func TestSchedulerJobsCreateDefaults(t *testing.T) {
	var mu sync.Mutex
	var got map[string]any
	api := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/jobs") {
			http.NotFound(w, r)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(got)
	})
	inv, stdout, _ := newFakeAPIInvocation(t, api)
	if err := runInvocation(inv, "scheduler", "jobs", "create", "http", "j1", "--project=p1", "--location=l1",
		"--schedule=0 * * * *", "--uri=https://example.com"); err != nil {
		t.Fatal(err)
	}
	if got["timeZone"] != "Etc/UTC" {
		t.Errorf("timeZone = %v, want Etc/UTC", got["timeZone"])
	}
	if target, _ := got["httpTarget"].(map[string]any); target["httpMethod"] != "POST" {
		t.Errorf("httpTarget = %v, want httpMethod POST", got["httpTarget"])
	}
	if want := "Created scheduler job [j1].\n"; stdout.String() != want {
		t.Errorf("stdout = %q, want %q", stdout.String(), want)
	}
}
