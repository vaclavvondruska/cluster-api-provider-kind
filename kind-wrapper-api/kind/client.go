package kind

import (
	"sigs.k8s.io/kind/pkg/cluster"
)

// Client defines methods required to interact with Kind
type Client interface {
	CreateCluster(name string, spec []byte) error
	DeleteCluster(name string) error
	ClusterHasNodes(name string) (bool, error)
}

// ProviderClient implements interaction with Kind Provider
type ProviderClient struct {
	kubeConfigPath string
	provider *cluster.Provider
}

// NewProviderClient creates a new instance of ProviderClient
func NewProviderClient(kubeConfigPath string) *ProviderClient {
	provider := cluster.NewProvider(cluster.ProviderWithDocker())
	return &ProviderClient{kubeConfigPath: kubeConfigPath, provider: provider}
}

// CreateCluster executes the Kind Provider command to create a new cluster
func (c *ProviderClient) CreateCluster(name string, spec []byte) error {
	return c.provider.Create(name, cluster.CreateWithRawConfig(spec))
}

// DeleteCluster executes the Kind Provider command to delete a cluster
func (c *ProviderClient) DeleteCluster(name string) error {
	return c.provider.Delete(name, c.kubeConfigPath)
}

// ClusterHasNodes checks if the specified cluster has any existing nodes
func (c *ProviderClient) ClusterHasNodes(name string) (bool, error) {
	clusterNodes, err := c.provider.ListNodes(name)
	if err != nil {
		return false, err
	}
	return len(clusterNodes) > 0, err
}

func parseClusterNamesFromCommandOutput(clusterNames []string) map[string]bool {
	parsedClusterNames := make(map[string]bool)
	for _, clusterName := range clusterNames {
		parsedClusterNames[clusterName] = true
	}
	return parsedClusterNames
}
