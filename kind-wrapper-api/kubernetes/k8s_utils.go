package kubernetes

import (
	"errors"
	"flag"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
	"log"
	"os"
	"path/filepath"
)

const kubeConfigAbsPathEnvKey = "KUBECONFIG"

// GetKubeConfigPath attempts to get a path to an existing kubeconfig file
// It checks several locations
func GetKubeConfigPath() string {
	kubeConfigPath := os.Getenv(kubeConfigAbsPathEnvKey)
	if kubeConfigPath == "" {
		var kubeConfigPathPointer *string
		if home := homedir.HomeDir(); home != "" {
			kubeConfigPathPointer = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "absolute path to kubeconfig file")
		} else {
			kubeConfigPathPointer = flag.String("kubeconfig", "", "absolute path to kubeconfig file")
		}
		flag.Parse()
		if kubeConfigPathPointer == nil {
			return ""
		}
		kubeConfigPath = *kubeConfigPathPointer
	}
	if _, err := os.Stat(kubeConfigPath); os.IsNotExist(err) {
		return ""
	}
	return kubeConfigPath
}

// GetKubeContexts retrieves an information about existing clusters and contexts from a kubeconfig file
func GetKubeContexts(kubeConfigPath string) (map[string]*api.Context, map[string]*api.Cluster, error) {
	if kubeConfigPath == "" {
		log.Printf("Failed to load Kubeconfig from %s\n", kubeConfigPath)
		return make(map[string]*api.Context), make(map[string]*api.Cluster), errors.New("failed to retrieve contexts from kubeconfig")
	}

	log.Printf("Loading Kubeconfig from %s\n", kubeConfigPath)

	data, err := ioutil.ReadFile(kubeConfigPath)
	if err != nil {
		log.Printf("Failed to read Kubeconfig from %s: %s\n", kubeConfigPath, err)
		return make(map[string]*api.Context), make(map[string]*api.Cluster), err
	}

	clientConfig, err := clientcmd.NewClientConfigFromBytes(data)
	if err != nil {
		log.Printf("Failed to parse Kubeconfig from %s: %s\n", kubeConfigPath, err)
		return make(map[string]*api.Context), make(map[string]*api.Cluster), err
	}

	rawConfig, err := clientConfig.RawConfig()

	if err != nil {
		log.Printf("Failed to extract Kubeconfig from %s: %s\n", kubeConfigPath, err)
		return make(map[string]*api.Context), make(map[string]*api.Cluster), err
	}

	return rawConfig.Contexts, rawConfig.Clusters, nil
}

