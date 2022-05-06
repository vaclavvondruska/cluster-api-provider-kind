/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	infrastructurev1alpha1 "cluster-api-provider-kind/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

type MockKindApiServerResponse struct {
	Status  int
	Payload string
}

type MockKindApiServer struct {
	/*desiredState  KindState
	pendingCycles int*/
	server                *httptest.Server
	createResponses       []MockKindApiServerResponse
	deleteResponses       []MockKindApiServerResponse
	statusResponses       []MockKindApiServerResponse
	defaultCreateResponse MockKindApiServerResponse
	defaultDeleteResponse MockKindApiServerResponse
	defaultStatusResponse MockKindApiServerResponse
}

var SimpleSuccessMockApiResponse = MockKindApiServerResponse{Status: http.StatusOK, Payload: "OK"}
var PendingStatusMockApiResponse = MockKindApiServerResponse{Status: http.StatusOK, Payload: "{\"state\":\"pending\"}"}
var RunningStatusMockApiResponse = MockKindApiServerResponse{Status: http.StatusOK, Payload: "{\"state\":\"running\",\"host\":\"127.0.0.1\",\"port\":6443}"}
var NotFoundMockApiResponse = MockKindApiServerResponse{Status: http.StatusNotFound, Payload: "Not Found"}
var InternalServerErrorResponse = MockKindApiServerResponse{Status: http.StatusInternalServerError, Payload: "Internal Server Error"}

func (m *MockKindApiServer) Init() {
	m.defaultCreateResponse = SimpleSuccessMockApiResponse
	m.defaultDeleteResponse = SimpleSuccessMockApiResponse
	m.defaultStatusResponse = NotFoundMockApiResponse
	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, kindApiPathCluster) {
			if len(m.createResponses) > 0 {
				response := m.createResponses[0]
				m.createResponses = m.createResponses[1:]
				m.writeResponse(w, response)
			} else {
				m.writeResponse(w, m.defaultCreateResponse)
			}
		} else if r.Method == http.MethodDelete {
			if len(m.deleteResponses) > 0 {
				response := m.deleteResponses[0]
				m.deleteResponses = m.deleteResponses[1:]
				m.writeResponse(w, response)
			} else {
				m.writeResponse(w, m.defaultDeleteResponse)
			}
		} else if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, kindApiPathCluster) {
			if len(m.statusResponses) > 0 {
				response := m.statusResponses[0]
				m.statusResponses = m.statusResponses[1:]
				m.writeResponse(w, response)
			} else {
				m.writeResponse(w, m.defaultStatusResponse)
			}
		} else {
			m.writeResponse(w, NotFoundMockApiResponse)
		}
	}))
}

func (m *MockKindApiServer) Reset() {
	m.createResponses = []MockKindApiServerResponse{}
	m.deleteResponses = []MockKindApiServerResponse{}
	m.statusResponses = []MockKindApiServerResponse{}
	m.defaultCreateResponse = SimpleSuccessMockApiResponse
	m.defaultDeleteResponse = SimpleSuccessMockApiResponse
	m.defaultStatusResponse = NotFoundMockApiResponse
}

func (m *MockKindApiServer) AddCreateResponse(response MockKindApiServerResponse) {
	m.createResponses = append(m.createResponses, response)
}

func (m *MockKindApiServer) AddDeleteResponse(response MockKindApiServerResponse) {
	m.deleteResponses = append(m.createResponses, response)
}

func (m *MockKindApiServer) AddStatusResponse(response MockKindApiServerResponse) {
	m.statusResponses = append(m.createResponses, response)
}

func (m *MockKindApiServer) SetDefaultCreateResponse(response MockKindApiServerResponse) {
	m.defaultCreateResponse = response
}

func (m *MockKindApiServer) SetDefaultDeleteResponse(response MockKindApiServerResponse) {
	m.defaultDeleteResponse = response
}

func (m *MockKindApiServer) SetDefaultStatusResponse(response MockKindApiServerResponse) {
	m.defaultStatusResponse = response
}

func (m *MockKindApiServer) writeResponse(w http.ResponseWriter, response MockKindApiServerResponse) {
	w.WriteHeader(response.Status)
	if _, err := fmt.Fprint(w, response.Payload); err != nil {
		fmt.Printf("Failed to write response: %s\n", err)
	}
}

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc
var mockKindApiServer *MockKindApiServer

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {

	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	mockKindApiServer = &MockKindApiServer{}
	mockKindApiServer.Init()

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = infrastructurev1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).NotTo(HaveOccurred())

	kindClient := &KindClient{mockKindApiServer.server.URL, http.DefaultClient}

	err = (&KindClusterReconciler{
		Client:     k8sManager.GetClient(),
		KindClient: kindClient,
		Scheme:     k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).NotTo(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
	}()

}, 60)

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
