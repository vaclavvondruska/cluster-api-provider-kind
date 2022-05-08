package kind

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

const newLineSeparator = "\n"

type execContext = func(name string, arg ...string) *exec.Cmd

// Client defines methods required to interact with Kind
type Client interface {
	CreateCluster(spec string) error
	DeleteCluster(name string) error
	GetClusters() (map[string]bool, error)
}

// CLIClient implements interaction with Kind CLI
type CLIClient struct {
	cmdContext execContext
}

// NewCLIClient creates a new instance of CLIClient
func NewCLIClient() *CLIClient {
	return &CLIClient{cmdContext: exec.Command}
}

// CreateCluster executes the Kind CLI command to create a new cluster
func (c *CLIClient) CreateCluster(spec string) error {
	if spec == "" {
		return errors.New("failed to create cluster - empty spec provided")
	}
	cmdStr := fmt.Sprintf("kind create cluster --config=- <<EOF\n%s\nEOF\n", spec)
	cmd := c.cmdContext("bash", "-c", cmdStr)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Create Cluster command output: %s\n", output)
		return err
	} else {
		log.Printf("Create Cluster command output: %s\n", output)
	}
	return nil
}

// DeleteCluster executes the Kind CLI command to delete a cluster
func (c *CLIClient) DeleteCluster(name string) error {
	if len(name) == 0 {
		return errors.New("failed to delete cluster - empty name provided")
	}
	cmd := c.cmdContext("kind", "delete", "cluster", "--name", name)
	return cmd.Run()
}

// GetClusters executes the Kind CLI command to get a list of cluster names
// and converts the output to an array of strings
func (c *CLIClient) GetClusters() (map[string]bool, error) {
	cmd := c.cmdContext("kind", "get", "clusters")
	output, err := cmd.Output()
	if err != nil {
		return make(map[string]bool), err
	}
	return parseClusterNamesFromCommandOutput(output), nil
}

func parseClusterNamesFromCommandOutput(output []byte) map[string]bool {
	clusters := make(map[string]bool)
	outputStr := strings.Trim(string(output), newLineSeparator)
	if outputStr != "" {
		clusterNames := strings.Split(outputStr, newLineSeparator)
		for _, clusterName := range clusterNames {
			clusters[clusterName] = true
		}
	}
	return clusters
}
