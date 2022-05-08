package kindclient

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

const newLineSeparator = "\n"

// CreateCluster executes the Kind CLI command to create a new cluster
func CreateCluster(spec string) error {
	if len(spec) == 0 {
		return errors.New("failed to create cluster - empty spec provided")
	}
	cmdStr := fmt.Sprintf("kind create cluster --config=- <<EOF\n%s\nEOF\n", spec)
	cmd := exec.Command("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Create Cluster command output: %s\n", output)
		return err
	}
	return nil
}

// DeleteCluster executes the Kind CLI command to delete a cluster
func DeleteCluster(name string) error {
	if len(name) == 0 {
		return errors.New("failed to delete cluster - empty name provided")
	}
	cmd := exec.Command("kind", "delete", "cluster", "--name", name)
	return cmd.Run()
}

// GetClusters executes the Kind CLI command to get a list of cluster names
// and converts the output to an array of strings
func GetClusters() (map[string]bool, error) {
	cmd := exec.Command("kind", "get", "clusters")
	output, err := cmd.Output()
	if err != nil {
		return make(map[string]bool), err
	}

	return parseClusterNamesFromCommandOutput(output), nil
}

func parseClusterNamesFromCommandOutput(output []byte) map[string]bool {
	clusters := make(map[string]bool)
	outputStr := strings.Trim(string(output), newLineSeparator)
	clusterNames := strings.Split(outputStr, newLineSeparator)
	for _, clusterName := range clusterNames {
		clusters[clusterName] = true
	}
	return clusters
}
