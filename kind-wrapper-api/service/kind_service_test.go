package service

import (
	"errors"
	"github.com/stretchr/testify/require"
	"kind-wrapper-api/test"
	"os"
	"testing"
	"time"
)

func TestKindService(t *testing.T) {
	tempDir := os.TempDir()
	kubeConfigDir, err := os.MkdirTemp(tempDir, "kube")
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(kubeConfigDir)
	}()

	kubeConfigPath, err := test.SetupKubeConfig(kubeConfigDir, test.KubeConfig)
	require.NoError(t, err)

	t.Run("test get cluster state", func(t *testing.T) {
		mockKindClient := test.NewMockKindClient()
		mockKindClient.SetDefaultGet(func() (map[string]bool, error) {
			return map[string]bool{"kind": true, "kind-2": true}, nil
		})
		kindService := NewKindService(mockKindClient, kubeConfigPath)

		state, err := kindService.GetClusterState("kind")
		require.NoError(t, err)
		require.Equal(t, KindClusterStateRunning, state.State)

		state, err = kindService.GetClusterState("kind-2")
		require.NoError(t, err)
		require.Equal(t, KindClusterStatePending, state.State)

		_, err = kindService.GetClusterState("kind-3")
		require.Error(t, err)
	})

	t.Run("test get cluster state failure", func(t *testing.T) {
		mockKindClient := test.NewMockKindClient()
		mockKindClient.SetDefaultGet(func() (map[string]bool, error) {
			return make(map[string]bool), errors.New("failed to get clusters")
		})
		kindService := NewKindService(mockKindClient, kubeConfigPath)
		_, err = kindService.GetClusterState("kind")
		require.Error(t, err)
	})

	t.Run("test cluster creation", func(t *testing.T) {
		mockKindClient := test.NewMockKindClient()
		mockKindClient.AddGet(func() (map[string]bool, error) {
			return make(map[string]bool), nil
		})
		mockKindClient.AddGet(func() (map[string]bool, error) {
			return make(map[string]bool), nil
		})
		mockKindClient.AddGet(func() (map[string]bool, error) {
			return make(map[string]bool), nil
		})
		mockKindClient.SetDefaultGet(func() (map[string]bool, error) {
			return map[string]bool{"kind": true}, nil
		})
		mockKindClient.SetCreate(func() error {
			time.Sleep(3 * time.Second)
			return nil
		})

		spec := ClusterConfig{Name: "kind"}
		kindService := NewKindService(mockKindClient, kubeConfigPath)
		err = kindService.CreateCluster(spec)
		require.NoError(t, err)
	})

	t.Run("test get cluster creation failure", func(t *testing.T) {
		mockKindClient := test.NewMockKindClient()
		mockKindClient.AddGet(func() (map[string]bool, error) {
			return make(map[string]bool), nil
		})
		mockKindClient.AddGet(func() (map[string]bool, error) {
			return make(map[string]bool), nil
		})
		mockKindClient.AddGet(func() (map[string]bool, error) {
			return make(map[string]bool), nil
		})
		mockKindClient.SetDefaultGet(func() (map[string]bool, error) {
			return map[string]bool{"kind": true}, nil
		})
		mockKindClient.SetCreate(func() error {
			time.Sleep(time.Second)
			return errors.New("failed to create cluster")
		})

		spec := ClusterConfig{Name: "kind"}
		kindService := NewKindService(mockKindClient, kubeConfigPath)
		err = kindService.CreateCluster(spec)
		require.Error(t, err)

		mockKindClient.AddGet(func() (map[string]bool, error) {
			return make(map[string]bool), nil
		})
		mockKindClient.AddGet(func() (map[string]bool, error) {
			return make(map[string]bool), errors.New("failed to get clusters")
		})
		mockKindClient.SetCreate(func() error {
			time.Sleep(3 * time.Second)
			return nil
		})
		err = kindService.CreateCluster(spec)
		require.Error(t, err)
	})
}
