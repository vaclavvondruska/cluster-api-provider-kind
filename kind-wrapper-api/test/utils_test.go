package test

import (
	"errors"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestMockClientDefaultBehaviour(t *testing.T) {
	mockKindClient := NewMockKindClient()

	require.NoError(t, mockKindClient.CreateCluster("kind", []byte{}), "Default Create should succeed")

	require.NoError(t, mockKindClient.DeleteCluster("kind"), "Default Delete should succeed")

	hasNodes, err := mockKindClient.ClusterHasNodes("kind")
	require.NoError(t, err, "Default Get should succeed")
	require.False(t, hasNodes, "Default result of ClusterHasNodes should be false")
}

func TestMockClientCustomFailures(t *testing.T) {
	mockKindClient := NewMockKindClient()

	mockKindClient.SetCreate(func() error {
		return errors.New("failed to create cluster")
	})
	require.Error(t, mockKindClient.CreateCluster("kind", []byte{}), "Custom Create should fail")

	mockKindClient.SetDelete(func() error {
		return errors.New("failed to delete cluster")
	})
	require.Error(t, mockKindClient.DeleteCluster("kind"), "Custom Delete should fail")

	mockKindClient.SetDefaultHasNodes(func() (bool, error) {
		return false, errors.New("failed to get clusters")
	})
	_, err := mockKindClient.ClusterHasNodes("kind")
	require.Error(t, err, "Custom ClusterHasNodes should fail")
}

func TestMockClientCustomHasNodesQueue(t *testing.T) {
	mockKindClient := NewMockKindClient()
	mockKindClient.SetDefaultHasNodes(func() (bool, error) {
		return false, errors.New("failed to get clusters")
	})
	mockKindClient.AddHasNodes(func() (bool, error) {
		return false, nil
	})
	mockKindClient.AddHasNodes(func() (bool, error) {
		return true, nil
	})
	hasNodes, err := mockKindClient.ClusterHasNodes("kind")
	require.NoError(t, err, "First custom Get should not fail")
	require.False(t, hasNodes, "First custom ClusterHasNodes return false")

	hasNodes, err = mockKindClient.ClusterHasNodes("kind")
	require.NoError(t, err, "Second custom Get should not fail")
	require.True(t, hasNodes, "Second custom ClusterHasNodes return true")

	_, err = mockKindClient.ClusterHasNodes("kind")
	require.Error(t, err, "Third custom ClusterHasNodes should fall back to default and fail")
}

func TestMockClientClearCustomHasNodesQueue(t *testing.T) {
	mockKindClient := NewMockKindClient()
	require.Len(t, mockKindClient.hasNodesQueue, 0, "Initial length of HasNodes queue should be 0")

	mockKindClient.AddHasNodes(func() (bool, error) {
		return false, nil
	})
	require.Len(t, mockKindClient.hasNodesQueue, 1, "Length of updated HasNodes queue should be 1")

	mockKindClient.ClearHasNodesQueue()
	require.Len(t, mockKindClient.hasNodesQueue, 0, "Length of cleared HasNodes queue should be 0")
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
