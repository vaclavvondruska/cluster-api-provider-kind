package service

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"kind-wrapper-api/kindclient"
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
	kubeConfigPath string
}

// NewKindService creates a new instance of KindService
func NewKindService(kubeConfigPath string) *KindService {
	return &KindService{kubeConfigPath: kubeConfigPath}
}

// CreateCluster creates a new Kind cluster from the provided specifications
// The method waits for the cluster to appear in the output of "kind get clusters"
// or for the creation of the cluster to complete - whichever comes first
// An error is returned in case the cluster could not be created
func (s *KindService) CreateCluster(spec ClusterConfig) error {
	createResult := make(chan int, 1)
	waitResult := make(chan int, 1)
	waitTicker := time.NewTicker(500 * time.Millisecond)

	go waitForCluster(spec.Name, waitTicker, createResult, waitResult)
	go executeAndNotifyCreateCluster(spec, createResult)

	result := <-waitResult
	waitTicker.Stop()

	if result == kindClusterCreationResultFailure {
		return fmt.Errorf("failed to create cluster %s", spec.Name)
	}
	return nil
}

// DeleteCluster calls Kind CLI to delete an existing cluster
// The deletion is asynchronous, an output is ignored
// Successful deletion can be verified by calling the GetClusterState method
func (s *KindService) DeleteCluster(name string) {
	go func(name string) {
		_ = kindclient.DeleteCluster(name)
	}(name)
}

// GetClusterState checks if a cluster with a specified name exists and returns its state:
// Running state is returned in case the cluster exists and is ready to be used
// Pending state is returned in case the cluster exists but is not ready
// Failed state is returned in case the cluster exists but Kind does not know about it
// KindClusterNotFoundError is returned in case the cluster does not exist
// A generic error is returned in case the cluster info could not be retrieved
func (s *KindService) GetClusterState(clusterName string) (KindClusterStatus, error) {
	contexts, clusterConfigs, err := kubernetes.GetKubeContexts(s.kubeConfigPath)
	if err != nil {
		return NewKindClusterStatus(KindClusterStateUnknown, ""), err
	}

	clusters, err := kindclient.GetClusters()
	if err != nil {
		return NewKindClusterStatus(KindClusterStateUnknown, ""), err
	}

	kindClusterName := kindClusterContextName(clusterName)

	_, clusterNameExists := clusters[clusterName]
	_, clusterContextExists := contexts[kindClusterName]
	clusterConfig, clusterConfigExists := clusterConfigs[kindClusterName]

	if clusterNameExists && clusterContextExists && clusterConfigExists {
		return NewKindClusterStatus(KindClusterStateRunning, clusterConfig.Server), nil
	} else if clusterNameExists {
		return NewKindClusterStatus(KindClusterStatePending, ""), err
	} else if clusterContextExists && clusterConfigExists {
		return NewKindClusterStatus(KindClusterStateFailed, ""), err
	}

	return NewKindClusterStatus(KindClusterStateUnknown, ""), KindClusterNotFoundError
}

func executeAndNotifyCreateCluster(spec ClusterConfig, createResult chan <- int) {
	specStr, err := yaml.Marshal(spec)
	log.Printf("Creating cluster from %s\n", specStr)
	if err != nil {
		log.Printf("Creation of cluster %s failed\n", spec.Name)
		createResult <- kindClusterCreationResultFailure
	} else if err = kindclient.CreateCluster(string(specStr)); err != nil {
		log.Printf("Creation of cluster %s failed\n", spec.Name)
		createResult <- kindClusterCreationResultFailure
	} else {
		log.Printf("Creation of cluster %s succeeded\n", spec.Name)
		createResult <- kindClusterCreationResultSuccess
	}
	close(createResult)
}
func waitForCluster(clusterName string, waitTicker *time.Ticker, createResult <- chan int, waitResult chan <- int) {
	for {
		select {
		case result := <-createResult:
			log.Printf("Creation of cluster %s finished with status %d\n", clusterName, result)
			waitResult <- result
			close(waitResult)
			return
		case <-waitTicker.C:
			if clusters, err := kindclient.GetClusters(); err == nil {
				_, ok := clusters[clusterName]
				if ok {
					log.Printf("Cluster %s ready\n", clusterName)
					waitResult <- kindClusterCreationResultSuccess
					close(waitResult)
					return
				}
			}
			log.Printf("Cluster %s not ready yet\n", clusterName)
		}
	}
}

func kindClusterContextName(clusterName string) string {
	return kindClusterContextPrefix + clusterName
}