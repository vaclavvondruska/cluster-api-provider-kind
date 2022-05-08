package test

import (
	"errors"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestMockKindClient(t *testing.T) {
	mockKindClient := NewMockKindClient()

	require.NoError(t, mockKindClient.CreateCluster("---"))
	require.NoError(t, mockKindClient.DeleteCluster("kind"))
	clusters, err := mockKindClient.GetClusters()
	require.NoError(t, err)
	require.Len(t, clusters, 0)

	mockKindClient.SetCreate(func() error {
		return errors.New("failed to create cluster")
	})
	require.Error(t, mockKindClient.CreateCluster("---"))

	mockKindClient.SetDelete(func() error {
		return errors.New("failed to delete cluster")
	})
	require.Error(t, mockKindClient.DeleteCluster("kind"))

	mockKindClient.SetDefaultGet(func() (map[string]bool, error) {
		return make(map[string]bool), errors.New("failed to get clusters")
	})
	_, err = mockKindClient.GetClusters()
	require.Error(t, err)

	mockKindClient.AddGet(func() (map[string]bool, error) {
		return make(map[string]bool), nil
	})
	mockKindClient.AddGet(func() (map[string]bool, error) {
		return map[string]bool{"kind": true}, nil
	})
	mockKindClient.AddGet(func() (map[string]bool, error) {
		return map[string]bool{"kind": true, "kind-2": true}, nil
	})
	clusters, err = mockKindClient.GetClusters()
	require.NoError(t, err)
	require.Len(t, clusters, 0)

	clusters, err = mockKindClient.GetClusters()
	require.NoError(t, err)
	require.Len(t, clusters, 1)

	clusters, err = mockKindClient.GetClusters()
	require.NoError(t, err)
	require.Len(t, clusters, 2)

	_, err = mockKindClient.GetClusters()
	require.Error(t, err)

}

func TestCreatingKubeConfigs(t *testing.T) {
	tempDir := os.TempDir()
	kubeConfigDir, err := os.MkdirTemp(tempDir, "kube")
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(kubeConfigDir)
	}()

	kubeConfigPath, err := SetupKubeConfig(kubeConfigDir, KubeConfig)
	require.NoError(t, err)

	_, err = os.Stat(kubeConfigPath)
	require.NoError(t, err)



}
