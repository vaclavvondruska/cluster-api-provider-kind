package test

import (
	"errors"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestMockClientDefaultBehaviour(t *testing.T) {
	mockKindClient := NewMockKindClient()

	require.NoError(t, mockKindClient.CreateCluster("---"), "Default Create should succeed")

	require.NoError(t, mockKindClient.DeleteCluster("kind"), "Default Delete should succeed")

	clusters, err := mockKindClient.GetClusters()
	require.NoError(t, err, "Default Get should succeed")
	require.Len(t, clusters, 0, "Default Get should return 0 clusters")
}

func TestMockClientCustomFailures(t *testing.T) {
	mockKindClient := NewMockKindClient()

	mockKindClient.SetCreate(func() error {
		return errors.New("failed to create cluster")
	})
	require.Error(t, mockKindClient.CreateCluster("---"), "Custom Create should fail")

	mockKindClient.SetDelete(func() error {
		return errors.New("failed to delete cluster")
	})
	require.Error(t, mockKindClient.DeleteCluster("kind"), "Custom Delete should fail")

	mockKindClient.SetDefaultGet(func() (map[string]bool, error) {
		return make(map[string]bool), errors.New("failed to get clusters")
	})
	_, err := mockKindClient.GetClusters()
	require.Error(t, err, "Custom Get should fail")
}

func TestMockClusterCustomGetQueue(t *testing.T) {
	mockKindClient := NewMockKindClient()
	mockKindClient.SetDefaultGet(func() (map[string]bool, error) {
		return make(map[string]bool), errors.New("failed to get clusters")
	})
	mockKindClient.AddGet(func() (map[string]bool, error) {
		return make(map[string]bool), nil
	})
	mockKindClient.AddGet(func() (map[string]bool, error) {
		return map[string]bool{"kind": true}, nil
	})
	mockKindClient.AddGet(func() (map[string]bool, error) {
		return map[string]bool{"kind": true, "kind-2": true}, nil
	})
	clusters, err := mockKindClient.GetClusters()
	require.NoError(t, err, "First custom Get should not fail")
	require.Len(t, clusters, 0, "First custom Get should return 0 clusters")

	clusters, err = mockKindClient.GetClusters()
	require.NoError(t, err, "Second custom Get should not fail")
	require.Len(t, clusters, 1, "Second custom Get should return 1 cluster")

	clusters, err = mockKindClient.GetClusters()
	require.NoError(t, err, "Third custom Get should not fail")
	require.Len(t, clusters, 2, "Third custom Get should return 2 clusters")

	_, err = mockKindClient.GetClusters()
	require.Error(t, err, "Fourth custom Get should fall back to default and fail")
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
