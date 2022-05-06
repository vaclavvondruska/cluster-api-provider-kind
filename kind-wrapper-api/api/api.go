package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"gopkg.in/yaml.v2"
	"kind-wrapper-api/service"
	"log"
	"net/http"
)

// API implements HTTP server and routing of incoming requests
type API struct {
	host string
	port int
	kindService *service.KindService
}

// NewAPI creates a new instance of API with the specified host, port and service (as a source of data)
func NewAPI(host string, port int, kindService *service.KindService) *API {
	return &API{host: host, port: port, kindService: kindService}
}

// Start binds all available routes and starts the server
func (api *API) Start() error {
	addr := fmt.Sprintf("%s:%d", api.host, api.port)
	router := &httprouter.Router{
		NotFound: http.NotFoundHandler(),
	}
	router.GET("/health", api.handleHealth)
	router.GET("/api/v1/cluster/:name", api.handleGetClusterStatus)
	router.POST("/api/v1/cluster", api.handleCreateClusterAsync)
	router.DELETE("/api/v1/cluster/:name", api.handleDeleteClusterAsync)
	log.Printf("Listening on %s", addr)
	return http.ListenAndServe(addr, router)
}

func (api *API) handleCreateClusterAsync(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	var clusterConfig service.ClusterConfig
	if err := yaml.NewDecoder(req.Body).Decode(&clusterConfig); err != nil {
		writeResponse(w, http.StatusBadRequest, fmt.Sprintf("Bad request!\nFailed to parse request payload: %s", err))
	} else {
		if _, err = api.kindService.GetClusterState(clusterConfig.Name); err == nil {
			writeResponse(w, http.StatusConflict, fmt.Sprintf("Conflict!\nCluster with the same name already exists: %s", err))
		} else if err := api.kindService.CreateCluster(clusterConfig); err != nil {
			writeResponse(w, http.StatusInternalServerError, "Internal Server Error")
		} else {
			writeResponse(w, http.StatusOK, "OK")
		}
	}
}

func (api *API) handleDeleteClusterAsync(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	name := params.ByName("name")
	if name == "" {
		writeResponse(w, http.StatusBadRequest, "Bad Request!\nInvalid name provided")
	} else {
		api.kindService.DeleteCluster(name)
		writeResponse(w, http.StatusOK, "OK")
	}
}

func (api *API) handleGetClusterStatus(w http.ResponseWriter, _ *http.Request, params httprouter.Params) {
	name := params.ByName("name")
	if name == "" {
		writeResponse(w, http.StatusBadRequest, "Bad Request!\nInvalid name provided")
	} else {
		clusterStatus, err := api.kindService.GetClusterState(name)
		if err != nil && errors.Is(err, service.KindClusterNotFoundError) {
			writeResponse(w, http.StatusNotFound, "Not Found")
		} else if err != nil {
			writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal Server Error!\n%s", err))
		} else {
			data, err := json.Marshal(clusterStatus)
			if err != nil {
				writeResponse(w, http.StatusInternalServerError, fmt.Sprintf("Internal Server Error!\n%s", err))
			} else {
				writeResponse(w, http.StatusOK, string(data))
			}
		}
	}
}

func (api *API) handleHealth(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
	writeResponse(w, http.StatusOK, "OK")
}

func writeResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.WriteHeader(statusCode)
	if _, err := fmt.Fprint(w, payload); err != nil {
		fmt.Printf("Failed to write response: %s\n", err)
	}
}