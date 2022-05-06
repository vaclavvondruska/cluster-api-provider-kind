package controllers

import (
	"cluster-api-provider-kind/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"net/http"
)

var _ = Describe("KindClient", func() {

	var namespace string
	var name string
	var kindClient *KindClient
	var spec v1alpha1.KindClusterSpec

	BeforeEach(func() {
		namespace = "default"
		name = "kind-cluster"
		spec = v1alpha1.KindClusterSpec{Nodes: []v1alpha1.KindClusterNode{{Role: "control-plane"}}}
		kindClient = &KindClient{host: mockKindApiServer.server.URL, client: http.DefaultClient}
	})

	AfterEach(func() {
		mockKindApiServer.Reset()
	})

	It("should handle cluster creation", func() {
		mockKindApiServer.SetDefaultCreateResponse(SimpleSuccessMockApiResponse)
		Expect(kindClient.CreateCluster(namespace, name, spec)).To(Succeed())

		mockKindApiServer.SetDefaultCreateResponse(InternalServerErrorResponse)
		Expect(kindClient.CreateCluster(namespace, name, spec)).NotTo(Succeed())
	})

	It("should handle cluster deletion", func() {
		mockKindApiServer.SetDefaultDeleteResponse(SimpleSuccessMockApiResponse)
		Expect(kindClient.DeleteCluster(namespace, name)).To(Succeed())

		mockKindApiServer.SetDefaultDeleteResponse(InternalServerErrorResponse)
		Expect(kindClient.DeleteCluster(namespace, name)).NotTo(Succeed())
	})

	It("should retrieve cluster status", func() {
		mockKindApiServer.SetDefaultStatusResponse(NotFoundMockApiResponse)
		_, err := kindClient.GetClusterStatus(namespace, name)
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(KindClusterNotFoundError))

		mockKindApiServer.SetDefaultStatusResponse(PendingStatusMockApiResponse)
		status, err := kindClient.GetClusterStatus(namespace, name)
		Expect(err).NotTo(HaveOccurred())
		Expect(status.State).To(Equal(KindStatePending))

		mockKindApiServer.SetDefaultStatusResponse(RunningStatusMockApiResponse)
		status, err = kindClient.GetClusterStatus(namespace, name)
		Expect(err).NotTo(HaveOccurred())
		Expect(status.State).To(Equal(KindStateRunning))
		Expect(status.Host).To(Equal("127.0.0.1"))
		Expect(status.Port).To(Equal(6443))

		mockKindApiServer.SetDefaultStatusResponse(InternalServerErrorResponse)
		_, err = kindClient.GetClusterStatus(namespace, name)
		Expect(err).To(HaveOccurred())
		Expect(err).NotTo(Equal(KindClusterNotFoundError))
	})

})
