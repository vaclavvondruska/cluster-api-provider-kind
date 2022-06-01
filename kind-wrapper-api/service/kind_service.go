package service

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"kind-wrapper-api/kind"
	"kind-wrapper-api/kubernetes"
	"log"
	"time"
)

const (
	kindClusterContextPrefix = "kind-"
	kindClusterCreationResultSuccess = 0
	kindClusterCreationResultFailure = 1
)

// KindClusterNotFoundError is returned by the GetClusterState method in case the cluster does not exist
var KindClusterNotFoundError = errors.New("cluster not found in kind")

// KindService provides information about Kind clusters based on data read from Kind CLI
type KindService struct {
	kindClient kind.Client
	kubeConfigPath string
}

// NewKindService creates a new instance of KindService
func NewKindService(kindClient kind.Client, kubeConfigPath string) *KindService {
	return &KindService{kindClient: kindClient, kubeConfigPath: kubeConfigPath}
}

// CreateCluster creates a new Kind cluster from the provided specifications
// The method waits for the cluster to appear in the output of "kind get clusters"
// or for the creation of the cluster to complete - whichever comes first
// An error is returned in case the cluster could not be created
func (s *KindService) CreateCluster(spec ClusterConfig) error {
	chanCreate := make(chan int, 1)
	chanWait := make(chan int, 1)
	waitTicker := time.NewTicker(500 * time.Millisecond)

	go awaitCluster(s.kindClient, spec.Name, waitTicker, chanCreate, chanWait)
	go executeAndNotifyCreateCluster(s.kindClient, spec, chanCreate)

	result := <-chanWait
	waitTicker.Stop()

	if result == kindClusterCreationResultFailure {
		return fmt.Errorf("failed to create cluster %s", spec.Name)
	}
	return nil
}

// DeleteCluster calls Kind CLI to delete an existing cluster
// The deletion is asynchronous, the output is ignored
// Successful deletion can be verified by calling the GetClusterState method
func (s *KindService) DeleteCluster(name string) {
	go func(name string) {
		_ = s.kindClient.DeleteCluster(name)
	}(name)
}

// GetClusterState checks if a cluster with a specified name exists and returns its state:
// Running state is returned in case the cluster exists and is ready to be used
// Pending state is returned in case the cluster exists but is not ready
// Failed state is returned in case the cluster exists but Kind does not know about it
// KindClusterNotFoundError is returned in case the cluster does not exist
// A generic error is returned in case the cluster info could not be retrieved
func (s *KindService) GetClusterState(clusterName string) (KindClusterStatus, error) {
	clusterHasNodes, err := s.kindClient.ClusterHasNodes(clusterName)
	if err != nil {
		return NewKindClusterStatus(KindClusterStateUnknown, ""), err
	}
	_, clusterConfigs, err := kubernetes.GetKubeContexts(s.kubeConfigPath)
	if err != nil {
		return NewKindClusterStatus(KindClusterStateUnknown, ""), err
	}

	clusterConfig, clusterHasConfig := clusterConfigs[kindClusterContextName(clusterName)]

	if !clusterHasNodes && !clusterHasConfig {
		return NewKindClusterStatus(KindClusterStateUnknown, ""), KindClusterNotFoundError
	} else if !clusterHasNodes && clusterHasConfig {
		return NewKindClusterStatus(KindClusterStateFailed, ""), nil
	} else if clusterHasNodes && !clusterHasConfig {
		return NewKindClusterStatus(KindClusterStatePending, ""), nil
	}
	return NewKindClusterStatus(KindClusterStateRunning, clusterConfig.Server), nil
}

func executeAndNotifyCreateCluster(client kind.Client, spec ClusterConfig, chanCreate chan <- int) {
	specBytes, err := yaml.Marshal(spec)
	log.Printf("Creating cluster from %s\n", specBytes)
	if err != nil {
		log.Printf("Creation of cluster %s failed\n", spec.Name)
		chanCreate <- kindClusterCreationResultFailure
	} else if err = client.CreateCluster(spec.Name, specBytes); err != nil {
		log.Printf("Creation of cluster %s failed\n", spec.Name)
		chanCreate <- kindClusterCreationResultFailure
	} else {
		log.Printf("Creation of cluster %s succeeded\n", spec.Name)
		chanCreate <- kindClusterCreationResultSuccess
	}
	close(chanCreate)
}

func awaitCluster(client kind.Client, name string, ticker *time.Ticker, chanCreate <- chan int, chanWait chan <- int) {
	for {
		select {
		case result := <-chanCreate:
			log.Printf("Creation of cluster %s finished with status %d\n", name, result)
			chanWait <- result
			close(chanWait)
			return
		case <-ticker.C:
			clusterHasNodes, err := client.ClusterHasNodes(name)
			if err != nil {
				chanWait <- kindClusterCreationResultFailure
				close(chanWait)
				return
			} else if clusterHasNodes {
				log.Printf("Cluster %s ready\n", name)
				chanWait <- kindClusterCreationResultSuccess
				close(chanWait)
				return
			}
			log.Printf("Cluster %s not ready yet\n", name)
		}
	}
}

func kindClusterContextName(clusterName string) string {
	return kindClusterContextPrefix + clusterName
}
