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
	getQueue []func() (map[string]bool, error)
	defaultGet func() (map[string]bool, error)
	create func() error
	delete func() error
}

func NewMockKindClient() *MockKindClient {
	return &MockKindClient{
		getQueue: []func() (map[string]bool, error){},
		defaultGet: func() (map[string]bool, error) {
			return make(map[string]bool), nil
		},
		create: func() error {
			return nil
		},
		delete: func() error {
			return nil
		},
	}
}

func (m *MockKindClient) AddGet(get func() (map[string]bool, error)) {
	m.getQueue = append(m.getQueue, get)
}

func (m *MockKindClient) SetDefaultGet(get func() (map[string]bool, error)) {
	m.defaultGet = get
}

func (m *MockKindClient) SetCreate(create func() error) {
	m.create = create
}

func (m *MockKindClient) SetDelete(delete func() error) {
	m.delete = delete
}

func (m *MockKindClient) CreateCluster(_ string) error {
	return m.create()
}

func (m *MockKindClient) DeleteCluster(_ string) error {
	return m.delete()
}

func (m *MockKindClient) GetClusters() (map[string]bool, error) {
	if len(m.getQueue) > 0 {
		get := m.getQueue[0]
		m.getQueue = m.getQueue[1:]
		return get()
	}
	return m.defaultGet()
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