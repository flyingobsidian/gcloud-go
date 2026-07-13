package cmd

import (
	"strings"
	"testing"
)

func TestCheatSheetCommandRegistered(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "cheat-sheet" {
			return
		}
	}
	t.Fatal("cheat-sheet command not registered on rootCmd")
}

func TestCheatSheetTextEmbedded(t *testing.T) {
	if len(cheatSheetText) == 0 {
		t.Fatal("embedded cheat_sheet.txt is empty")
	}
	if !strings.Contains(cheatSheetText, "gcloud cheat-sheet") {
		t.Errorf("cheat sheet text does not mention gcloud cheat-sheet")
	}
}
