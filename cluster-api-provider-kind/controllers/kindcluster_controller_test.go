package controllers

import (
	"cluster-api-provider-kind/api/v1alpha1"
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Describe("KindCluster Controller", func() {

	var spec v1alpha1.KindClusterSpec

	BeforeEach(func() {
		spec = v1alpha1.KindClusterSpec{
			Nodes: []v1alpha1.KindClusterNode{{
				Role: "control-plane",
			}},
		}

		mockKindApiServer.Reset()
		//mockKindApiServer.SetPendingCycles(2)
	})

	AfterEach(func() {

	})

	It("should create Kind Cluster when a new KindCluster CR is received", func() {

		mockKindApiServer.SetDefaultStatusResponse(PendingStatusMockApiResponse)
		mockKindApiServer.AddStatusResponse(NotFoundMockApiResponse)

		key := types.NamespacedName{
			Name:      "kind-cluster1",
			Namespace: "default",
		}

		kindCluster := &v1alpha1.KindCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: spec,
		}

		Expect(k8sClient.Create(context.Background(), kindCluster)).Should(Succeed())

		fetched := &v1alpha1.KindCluster{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			g.Expect(fetched.Status.State).To(Equal(v1alpha1.KindClusterStatePending))
		}, 20*time.Second, 2*time.Second).Should(Succeed())
	})

	It("should update existing KindCluster CR spec and state based on observed state", func() {
		mockKindApiServer.SetDefaultStatusResponse(RunningStatusMockApiResponse)
		mockKindApiServer.AddStatusResponse(NotFoundMockApiResponse)
		mockKindApiServer.AddStatusResponse(PendingStatusMockApiResponse)
		mockKindApiServer.AddStatusResponse(PendingStatusMockApiResponse)

		key := types.NamespacedName{
			Name:      "kind-cluster2",
			Namespace: "default",
		}

		kindCluster := &v1alpha1.KindCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: spec,
		}

		Expect(k8sClient.Create(context.Background(), kindCluster)).Should(Succeed())

		fetched := &v1alpha1.KindCluster{}
		Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			g.Expect(fetched.Status.State).To(Equal(v1alpha1.KindClusterStateRunning))
		}, 20*time.Second, 2*time.Second).Should(Succeed())

		//Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
		Expect(fetched.HasFinalizer(v1alpha1.KindClusterFinalizerName)).To(BeTrue())
		Expect(fetched.HasControlPlaneEndpoint()).To(BeTrue())
	})

	It("should delete all related resources when KindCluster CR is deleted", func() {
		mockKindApiServer.SetDefaultStatusResponse(NotFoundMockApiResponse)
		mockKindApiServer.AddStatusResponse(PendingStatusMockApiResponse)
		mockKindApiServer.AddStatusResponse(PendingStatusMockApiResponse)
		mockKindApiServer.AddStatusResponse(PendingStatusMockApiResponse)

		key := types.NamespacedName{
			Name:      "kind-cluster3",
			Namespace: "default",
		}

		kindCluster := &v1alpha1.KindCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec: spec,
		}

		Expect(k8sClient.Create(context.Background(), kindCluster)).Should(Succeed())
		Expect(k8sClient.Delete(context.Background(), kindCluster)).Should(Succeed())
		fetched := &v1alpha1.KindCluster{}
		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(context.Background(), key, fetched)).NotTo(Succeed())
		}, 20*time.Second, 2*time.Second).Should(Succeed())

	})

})
