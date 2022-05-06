package controllers

import (
	"bytes"
	"cluster-api-provider-kind/api/v1alpha1"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"net/http"
	"os"
)

// KindState defines Kind cluster states, which can be obtained from the Kind Wrapper API
type KindState string

const (
	kindAPIHostEnvName = "KIND_API_HOST"
	kindAPIDefaultHost = "http://127.0.0.1:8888"

	kindApiPathCluster = "/api/v1/cluster"

	kindClusterKind       = "Cluster"
	kindClusterAPIVersion = "kind.x-k8s.io/v1alpha4"

	KindStatePending = KindState("pending")
	KindStateRunning = KindState("running")
	KindStateFailed  = KindState("failed")
)

// KindClusterNotFoundError is returned when a Kind cluster does not exist
var KindClusterNotFoundError = errors.New("kind cluster not found")

// KindClusterStatus defines a structure of Kind cluster status retrieved from the Kind Wrapper API
type KindClusterStatus struct {
	State KindState `json:"state"`
	Host  string    `json:"host,omitempty"`
	Port  int       `json:"port,omitempty"`
}

// KindClusterConfig defines the structure of a config YAML file for Kind clusters
type KindClusterConfig struct {
	v1alpha1.KindClusterSpec `yaml:",inline"`
	Kind                     string `yaml:"kind"`
	APIVersion               string `yaml:"apiVersion"`
	Name                     string `yaml:"name"`
}

// KindClient communicates with Kind Wrapper API external service via HTTP
type KindClient struct {
	host   string
	client *http.Client
}

// NewKindClient creates a new instance of KindClient with host read from an environment variable if it exists
func NewKindClient() *KindClient {
	host := os.Getenv(kindAPIHostEnvName)
	if host == "" {
		host = kindAPIDefaultHost
	}
	return &KindClient{host: host, client: &http.Client{}}
}

// CreateCluster sends a POST request with a Kind cluster configuration YAML to create a new Kind cluster
// Namespaced name must be unused, otherwise an error is returned
// A response is received when the Kind cluster gets created, not when it gets ready
func (u *KindClient) CreateCluster(namespace, name string, spec v1alpha1.KindClusterSpec) error {
	clusterName := compositeClusterName(namespace, name)
	url := fmt.Sprintf("%s%s", u.host, kindApiPathCluster)
	payload := KindClusterConfig{
		KindClusterSpec: spec,
		Kind:            kindClusterKind,
		APIVersion:      kindClusterAPIVersion,
		Name:            clusterName,
	}

	yamlBytes, err := yaml.Marshal(payload)
	if err != nil {
		return err
	}

	response, err := u.client.Post(url, "text/plain; charset=utf8", bytes.NewBuffer(yamlBytes))
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("received error status %d from kind api", response.StatusCode)
	}

	return nil
}

// DeleteCluster sends a DELETE request to delete a Kind cluster with a specified name
// A response is received when the Kind Wrapper API starts the deletion process, not when delete is finished
func (u *KindClient) DeleteCluster(namespace, name string) error {
	clusterName := compositeClusterName(namespace, name)
	url := fmt.Sprintf("%s%s/%s", u.host, kindApiPathCluster, clusterName)
	request, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	response, err := u.client.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("received error status %d from kind api", response.StatusCode)
	}

	return nil
}

// GetClusterStatus sends a GET request to get status of a specific Kind cluster
// `pending` status is returned when the cluster exists but is not ready
// `running` status is returned when the cluster is ready.
// Control plane host and port are included in the response for running clusters
// KindClusterNotFoundError is returhen when the cluster does not exist
// any other error means that the clent was unable to retrieve the cluster status
func (u *KindClient) GetClusterStatus(namespace, name string) (KindClusterStatus, error) {
	clusterName := compositeClusterName(namespace, name)
	url := fmt.Sprintf("%s%s/%s", u.host, kindApiPathCluster, clusterName)
	response, err := u.client.Get(url)

	var clusterStatus KindClusterStatus

	if err != nil {
		return clusterStatus, err
	}

	if response.StatusCode == http.StatusNotFound {
		return clusterStatus, KindClusterNotFoundError
	}

	err = json.NewDecoder(response.Body).Decode(&clusterStatus)
	if err != nil {
		return KindClusterStatus{}, err
	}
	return clusterStatus, nil
}

func compositeClusterName(namespace, name string) string {
	return fmt.Sprintf("%s-%s", namespace, name)
}
