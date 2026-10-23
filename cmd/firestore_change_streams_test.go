package cmd

import "testing"

func TestFirestoreChangeStreamsSubcommands(t *testing.T) {
	g := firestoreSubgroup("change-streams")
	if g == nil {
		t.Fatal("firestore change-streams missing")
	}
	assertSubcommands(t, g, []string{"create", "delete", "describe", "list"})
}

func TestFSChangeStreamBody(t *testing.T) {
	cs, err := fsChangeStreamBody("7d", true, false, "")
	if err != nil {
		t.Fatal(err)
	}
	if cs.RetentionPeriod != "604800s" || cs.DatabaseScope == nil || cs.CollectionGroupScope != nil {
		t.Errorf("database scope: %+v", cs)
	}

	cs, err = fsChangeStreamBody("1d", false, true, "my-cg")
	if err != nil {
		t.Fatal(err)
	}
	if cs.RetentionPeriod != "86400s" || cs.DatabaseScope != nil ||
		cs.CollectionGroupScope == nil || cs.CollectionGroupScope.CollectionGroupId != "my-cg" {
		t.Errorf("collection group scope: %+v", cs)
	}

	for _, tc := range []struct {
		retention string
		dbScope   bool
		cgSet     bool
		cg        string
	}{
		{"12h", true, false, ""},
		{"8d", true, false, ""},
		{"36h", true, false, ""},
		{"bogus", true, false, ""},
		{"1d", false, true, ""},
		{"1d", false, false, ""},
	} {
		if _, err := fsChangeStreamBody(tc.retention, tc.dbScope, tc.cgSet, tc.cg); err == nil {
			t.Errorf("fsChangeStreamBody(%+v): expected error", tc)
		}
	}
}
