package test

import (
	"os"
)

const EmptyKubeConfig = `---
apiVersion: v1
kind: Config
clusters: []
contexts: []
preferences: {}
users: []
`

const KubeConfig = `---
apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: ""
    server: https://127.0.0.1:8080
  name: kind-kind
contexts:
- context:
    cluster: kind-kind
    user: kind-kind
  name: kind-kind
preferences: {}
users:
- name: kind-kind
  user:
    client-certificate-data: ""
    client-key-data: ""`

type MockKindClient struct {
	hasNodesQueue []func() (bool, error)
	defaultHasNodes func() (bool, error)
	create func() error
	delete func() error
}

func NewMockKindClient() *MockKindClient {
	return &MockKindClient{
		hasNodesQueue: []func() (bool, error){},
		defaultHasNodes: func() (bool, error) {
			return false, nil
		},
		create: func() error {
			return nil
		},
		delete: func() error {
			return nil
		},
	}
}

func (m *MockKindClient) AddHasNodes(hasNodes func() (bool, error)) {
	m.hasNodesQueue = append(m.hasNodesQueue, hasNodes)
}

func (m *MockKindClient) ClearHasNodesQueue() {
	m.hasNodesQueue = []func() (bool, error){}
}

func (m *MockKindClient) SetDefaultHasNodes(hasNodes func() (bool, error)) {
	m.defaultHasNodes = hasNodes
}

func (m *MockKindClient) SetCreate(create func() error) {
	m.create = create
}

func (m *MockKindClient) SetDelete(delete func() error) {
	m.delete = delete
}

func (m *MockKindClient) CreateCluster(_ string, _ []byte) error {
	return m.create()
}

func (m *MockKindClient) DeleteCluster(_ string) error {
	return m.delete()
}

func (m *MockKindClient) ClusterHasNodes(_ string) (bool, error) {
	if len(m.hasNodesQueue) > 0 {
		hasNodes := m.hasNodesQueue[0]
		m.hasNodesQueue = m.hasNodesQueue[1:]
		return hasNodes()
	}
	return m.defaultHasNodes()
}

func SetupKubeConfig(dir string, content string) (string, error) {
	kubeConfigFile, err := os.CreateTemp(dir, "config")
	if err != nil {
		return "", err
	}

	defer func() {
		_ = kubeConfigFile.Close()
	}()

	_, err = kubeConfigFile.WriteString(content)
	if err != nil {
		return "", err
	}

	err = kubeConfigFile.Sync()
	if err != nil {
		return "", err
	}

	return kubeConfigFile.Name(), nil
}