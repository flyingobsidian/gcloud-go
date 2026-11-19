package cli

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	storage "google.golang.org/api/storage/v1"
)

func TestCRC32CCombine(t *testing.T) {
	table := crc32.MakeTable(crc32.Castagnoli)
	for _, tc := range []struct{ a, b string }{
		{"hello ", "world"},
		{"", "abc"},
		{"abc", ""},
		{strings.Repeat("x", 1000), strings.Repeat("y", 4097)},
	} {
		want := crc32.Checksum([]byte(tc.a+tc.b), table)
		got := crc32cCombine(crc32.Checksum([]byte(tc.a), table), crc32.Checksum([]byte(tc.b), table), int64(len(tc.b)))
		if got != want {
			t.Errorf("combine(%q, %q) = %08x, want %08x", tc.a, tc.b, got, want)
		}
	}
}

func crc32cB64(s string) string {
	return encodeCRC32C(crc32.Checksum([]byte(s), crc32.MakeTable(crc32.Castagnoli)))
}

// composeAPI serves source objects a and b, a compose that returns
// composedCRC, and records deletes.
func composeAPI(composedCRC string, deleted *[]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/o/a"):
			fmt.Fprintf(w, `{"bucket":"bk","name":"a","size":"6","crc32c":%q}`, crc32cB64("hello "))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/o/b"):
			fmt.Fprintf(w, `{"bucket":"bk","name":"b","size":"5","crc32c":%q}`, crc32cB64("world"))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/compose"):
			fmt.Fprintf(w, `{"bucket":"bk","name":"c","generation":"7","size":"11","crc32c":%q}`, composedCRC)
		case r.Method == http.MethodDelete:
			*deleted = append(*deleted, r.URL.Path+"?generation="+r.URL.Query().Get("generation"))
			w.WriteHeader(http.StatusNoContent)
		default:
			http.Error(w, r.Method+" "+r.URL.Path, http.StatusNotFound)
		}
	})
}

func writeComposeRequest(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "compose.json")
	if err := os.WriteFile(path, []byte(`{"sourceObjects":[{"name":"a"},{"name":"b"}]}`), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestStorageObjectsComposeValidatesCRC32C(t *testing.T) {
	var deleted []string
	inv, stdout, _ := newFakeAPIInvocation(t, composeAPI(crc32cB64("hello world"), &deleted))
	if err := runInvocation(inv, "storage", "objects", "compose", "c", "--bucket", "bk", "--config-file", writeComposeRequest(t), "--format", "value(name)"); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(stdout.String(), "Composed object [c] in bucket [bk].\n") {
		t.Errorf("stdout = %q", stdout)
	}
	if len(deleted) != 0 {
		t.Errorf("deleted = %v, want nothing", deleted)
	}
}

func TestStorageObjectsComposeHashMismatchDeletes(t *testing.T) {
	var deleted []string
	bad := crc32cB64("something else")
	inv, stdout, _ := newFakeAPIInvocation(t, composeAPI(bad, &deleted))
	err := runInvocation(inv, "storage", "objects", "compose", "c", "--bucket", "bk", "--config-file", writeComposeRequest(t))
	want := fmt.Sprintf("Source hash %s does not match destination hash %s for object gs://bk/c.", crc32cB64("hello world"), bad)
	if err == nil || err.Error() != want {
		t.Fatalf("err = %v, want %q", err, want)
	}
	if len(deleted) != 1 || !strings.HasSuffix(deleted[0], "/b/bk/o/c?generation=7") {
		t.Errorf("deleted = %v, want the composed object", deleted)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want nothing", stdout)
	}
}

func TestStorageObjectsComposeCheckHashesNever(t *testing.T) {
	var deleted []string
	inv, _, _ := newFakeAPIInvocation(t, composeAPI(crc32cB64("something else"), &deleted))
	if err := inv.env.SetProperty("storage", "check_hashes", "never"); err != nil {
		t.Fatal(err)
	}
	if err := runInvocation(inv, "storage", "objects", "compose", "c", "--bucket", "bk", "--config-file", writeComposeRequest(t)); err != nil {
		t.Fatal(err)
	}
	if len(deleted) != 0 {
		t.Errorf("deleted = %v, want nothing", deleted)
	}
}

func TestPrepareServerSideHashValidation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "f")
	if err := os.WriteFile(path, []byte("data"), 0600); err != nil {
		t.Fatal(err)
	}
	sum := md5.Sum([]byte("data"))
	wantMD5 := base64.StdEncoding.EncodeToString(sum[:])
	large := int64(64 << 20) // resumable; also differs from the real size

	for _, tc := range []struct {
		name    string
		vars    map[string]string
		size    int64
		wantMD5 string
		wantErr string
	}{
		{name: "small upload is not validated", size: 4},
		{name: "resumable upload", size: large, wantMD5: wantMD5, wantErr: "Source file size changed post-initiation."},
		{name: "disabled", vars: map[string]string{"CLOUDSDK_STORAGE_ENABLE_SERVER_SIDE_HASH_VALIDATION": "False"}, size: large},
		{name: "check_hashes never", vars: map[string]string{"CLOUDSDK_STORAGE_CHECK_HASHES": "never"}, size: large, wantErr: "Source file size changed post-initiation."},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inv, _, _ := newTestInvocation(t, tc.vars, "")
			s := &storageCmds{inv: inv}
			f, err := os.Open(path)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()
			obj := &storage.Object{}
			r, err := s.prepareServerSideHashValidation(f, tc.size, obj)
			if err != nil {
				t.Fatal(err)
			}
			if obj.Md5Hash != tc.wantMD5 {
				t.Errorf("Md5Hash = %q, want %q", obj.Md5Hash, tc.wantMD5)
			}
			data, err := io.ReadAll(r)
			if gotErr := fmt.Sprint(err); tc.wantErr != "" && gotErr != tc.wantErr || tc.wantErr == "" && err != nil {
				t.Errorf("read err = %v, want %q", err, tc.wantErr)
			}
			if string(data) != "data" {
				t.Errorf("read %q, want the whole file", data)
			}
		})
	}
}
