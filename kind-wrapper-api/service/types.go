package service

import (
	"net/url"
	"strconv"
)

type KindClusterState string

const (
	KindClusterStatePending = KindClusterState("pending")
	KindClusterStateRunning = KindClusterState("running")
	KindClusterStateUnknown = KindClusterState("unknown")
	KindClusterStateFailed = KindClusterState("failed")
)

// ExtraPortMappingConfig defines configuration options fpr extra port mappings in NodeConfig
type ExtraPortMappingConfig struct {
	ContainerPort int    `yaml:"containerPort,omitempty"`
	HostPort      int    `yaml:"hostPort,omitempty"`
	ListenAddress string `yaml:"listenAddress,omitempty"` // optional: set the bind address on the host; default 0.0.0.0
	Protocol      string `yaml:"protocol,omitempty"` // optional: set the protocol to one of TCP, UDP, SCTP; default TCP
}

// ExtraMountConfig defines configuration options for extra mounts in NodeConfig
type ExtraMountConfig struct {
	HostPath       string `yaml:"hostPath,omitempty"`
	ContainerPath  string `yaml:"containerPath,omitempty"`
	ReadOnly       bool   `yaml:"readOnly,omitempty"`       // optional: if set, the mount is read-only; default false
	SELinuxRelabel bool   `yaml:"selinuxRelabel,omitempty"`
	Propagation    string `yaml:"propagation,omitempty"`    // optional: set propagation mode (None, HostToContainer or Bidirectional); default None
	// see https://kubernetes.io/docs/concepts/storage/volumes/#mount-propagation
}

// NodeConfig defines configuration options for nodes in ClusterConfig
type NodeConfig struct {
	Role                 string                   `yaml:"role"`
	Image                string                   `yaml:"image,omitempty"`
	KubeAdmConfigPatches []string                 `yaml:"kubeadmConfigPatches,omitempty"`
	ExtraMounts          []ExtraMountConfig       `yaml:"extraMounts,omitempty"`
	ExtraPortMappings    []ExtraPortMappingConfig `yaml:"extraPortMappings,omitempty"`
}

// NetworkingConfig defines configuration options for networking in ClusterConfig
type NetworkingConfig struct {
	IPFamily          string `yaml:"ipFamily,omitempty"`
	APIServerAddress  string `yaml:"apiServerAddress,omitempty"`  // By default the API server listens on a random open port.
	// You may choose a specific port but probably don't need to in most cases.
	// Using a random port makes it easier to spin up multiple clusters.
	APIServerPort     int    `yaml:"apiServerPort,omitempty"`
	PodSubnet         string `yaml:"podSubnet,omitempty"`
	ServiceSubnet     string `yaml:"serviceSubnet,omitempty"`
	DisableDefaultCNI bool   `yaml:"disableDefaultCNI,omitempty"` // the default CNI will not be installed
	KubeProxyMode     string `yaml:"kubeProxyMode,omitempty"`     // iptables (default), ipvs, none
}

// ClusterConfig defines configuration options for Kind Cluster
type ClusterConfig struct {
	Kind          string            `yaml:"kind"`
	APIVersion    string            `yaml:"apiVersion"`
	Name          string            `yaml:"name"`
	FeatureGates  map[string]bool   `yaml:"featureGates,omitempty"`
	RuntimeConfig map[string]string `yaml:"runtimeConfig,omitempty"`
	Networking    NetworkingConfig  `yaml:"networking,omitempty"`
	Nodes         []NodeConfig      `yaml:"nodes,omitempty"`
}

// KindClusterStatus contains information about Kind cluster state and clontrol plane endpoint if available
type KindClusterStatus struct {
	State KindClusterState `json:"state"`
	Host  string           `json:"host,omitempty"`
	Port int               `json:"port,omitempty"`
}

// NewKindClusterStatus creates a new instance of KindClusterStatus
func NewKindClusterStatus(state KindClusterState, rawServerUrl string) KindClusterStatus {
	host := ""
	port := 0
	serverUrl, err := url.Parse(rawServerUrl)
	if err == nil {
		host = serverUrl.Hostname()
		portStr := serverUrl.Port()
		port, _ = strconv.Atoi(portStr)
	}
	return KindClusterStatus{State: state, Host: host, Port: port}
}