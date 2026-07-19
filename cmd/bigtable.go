package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// --- gcloud bigtable (#308) ---

var bigtableCmd = &cobra.Command{Use: "bigtable", Short: "Manage Cloud Bigtable"}

func init() {
	rootCmd.AddCommand(bigtableCmd)
}

// btProjectName returns "projects/PROJECT" for the resolved project.
func btProjectName() (string, error) {
	project, err := resolveProject()
	if err != nil {
		return "", err
	}
	return "projects/" + project, nil
}

// btInstanceName returns projects/.../instances/INSTANCE.
func btInstanceName(instance string) (string, error) {
	if instance == "" {
		return "", fmt.Errorf("--instance is required")
	}
	if strings.HasPrefix(instance, "projects/") {
		return instance, nil
	}
	project, err := btProjectName()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/instances/%s", project, instance), nil
}

// btClusterName returns projects/.../instances/INSTANCE/clusters/CLUSTER.
func btClusterName(instance, cluster string) (string, error) {
	if cluster == "" {
		return "", fmt.Errorf("--cluster is required")
	}
	if strings.HasPrefix(cluster, "projects/") {
		return cluster, nil
	}
	inst, err := btInstanceName(instance)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/clusters/%s", inst, cluster), nil
}

// btTableName returns projects/.../instances/INSTANCE/tables/TABLE.
func btTableName(instance, table string) (string, error) {
	if table == "" {
		return "", fmt.Errorf("--table is required")
	}
	if strings.HasPrefix(table, "projects/") {
		return table, nil
	}
	inst, err := btInstanceName(instance)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/tables/%s", inst, table), nil
}
