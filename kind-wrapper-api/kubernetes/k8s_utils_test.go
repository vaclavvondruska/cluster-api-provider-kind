package kubernetes

import (
	"github.com/stretchr/testify/require"
	"kind-wrapper-api/test"
	"os"
	"testing"
)

func TestKubernetesUtils(t *testing.T) {
	tempDir := os.TempDir()
	kubeConfigDir, err := os.MkdirTemp(tempDir, "kube")
	require.NoError(t, err)

	defer func() {
		_ = os.RemoveAll(kubeConfigDir)
	}()

	kubeConfigPath, err := test.SetupKubeConfig(kubeConfigDir, test.KubeConfig)
	require.NoError(t, err)

	emptyConfigPath, err := test.SetupKubeConfig(kubeConfigDir, test.EmptyKubeConfig)
	require.NoError(t, err)

	t.Run("test get kubeconfig path", func(t *testing.T) {
		retrievedKubeConfigPath := GetKubeConfigPath()
		require.NotEqual(t, kubeConfigPath, retrievedKubeConfigPath)

		t.Setenv("KUBECONFIG", kubeConfigPath)
		retrievedKubeConfigPath = GetKubeConfigPath()
		require.Equal(t, kubeConfigPath, retrievedKubeConfigPath)
	})

	t.Run("test get kube contexts", func(t *testing.T) {
		_, _, err := GetKubeContexts("")
		require.Error(t, err)

		contexts, clusters, err := GetKubeContexts(emptyConfigPath)
		require.NoError(t, err)
		require.Len(t, contexts, 0)
		require.Len(t, clusters, 0)

		contexts, clusters, err = GetKubeContexts(kubeConfigPath)
		require.NoError(t, err)
		require.Len(t, contexts, 1)
		require.Len(t, clusters, 1)
		_, ok := contexts["kind-kind"]
		require.True(t, ok)
		_, ok = clusters["kind-kind"]
		require.True(t, ok)
	})
}
