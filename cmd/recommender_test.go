package cmd

import (
	"testing"

	"github.com/spf13/cobra"
)

func recommenderSubgroup(name string) *cobra.Command {
	for _, c := range recommenderCmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

func TestRecommenderInsightsSubcommands(t *testing.T) {
	g := recommenderSubgroup("insights")
	if g == nil {
		t.Fatal("recommender insights missing")
	}
	assertSubcommands(t, g, []string{"describe", "list", "mark-accepted", "mark-active", "mark-dismissed"})
}

func TestRecommenderRecommendationsSubcommands(t *testing.T) {
	g := recommenderSubgroup("recommendations")
	if g == nil {
		t.Fatal("recommender recommendations missing")
	}
	assertSubcommands(t, g, []string{
		"describe", "list", "mark-active", "mark-claimed", "mark-dismissed",
		"mark-failed", "mark-succeeded",
	})
}

func TestRecommenderInsightTypeConfigSubcommands(t *testing.T) {
	g := recommenderSubgroup("insight-type-config")
	if g == nil {
		t.Fatal("recommender insight-type-config missing")
	}
	assertSubcommands(t, g, []string{"describe", "update"})
}

func TestRecommenderRecommenderConfigSubcommands(t *testing.T) {
	g := recommenderSubgroup("recommender-config")
	if g == nil {
		t.Fatal("recommender recommender-config missing")
	}
	assertSubcommands(t, g, []string{"describe", "update"})
}
