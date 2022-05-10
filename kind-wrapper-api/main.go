package main

import (
	"fmt"
	"kind-wrapper-api/api"
	"kind-wrapper-api/kind"
	"kind-wrapper-api/kubernetes"
	"kind-wrapper-api/service"
	"os"
	"strconv"
)

const (
	apiHostEnvKey = "API_HOST"
	apiPortEnvKey = "API_PORT"

	defaultApiHost = "0.0.0.0"
	defaultApiPort = 8888
)

func main() {
	kubeConfigPath := kubernetes.GetKubeConfigPath()
	kindClient := kind.NewProviderClient(kubeConfigPath)
	kindService := service.NewKindService(kindClient, kubeConfigPath)

	host := os.Getenv(apiHostEnvKey)
	if host == "" {
		host = defaultApiHost
	}

	portStr := os.Getenv(apiPortEnvKey)
	port, err := strconv.Atoi(portStr)
	if port == 0 || err != nil {
		port = defaultApiPort
	}

	if err := api.NewAPI(host, port, kindService).Start(); err != nil {
		fmt.Println(fmt.Sprintf("Failed to start API: %s", err))
	}
}
