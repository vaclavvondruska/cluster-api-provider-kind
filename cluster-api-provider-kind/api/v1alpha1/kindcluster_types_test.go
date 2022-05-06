package v1alpha1

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
)

var _ = Describe("KindCluster", func() {

	var (
		key              types.NamespacedName
		created, fetched *KindCluster
	)

	Context("Create API", func() {

		It("should create a resource successfully", func() {
			key = types.NamespacedName{
				Name:      "kind-cluster",
				Namespace: "default",
			}
			created = &KindCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      key.Name,
					Namespace: key.Namespace,
				},
				Spec: KindClusterSpec{
					Nodes: []KindClusterNode{{Role: "control-plane"}},
				},
			}

			By("creating a resource")
			Expect(k8sClient.Create(context.Background(), created)).To(Succeed())

			fetched = &KindCluster{}
			Expect(k8sClient.Get(context.Background(), key, fetched)).To(Succeed())
			Expect(fetched).To(Equal(created))

			By("deleting the resource")
			Expect(k8sClient.Delete(context.Background(), created)).To(Succeed())
			Expect(k8sClient.Get(context.Background(), key, created)).NotTo(Succeed())
		})

		It("should correctly handle controlPlaneEndpoint", func() {
			kindCluster := &KindCluster{}
			Expect(kindCluster.HasControlPlaneEndpoint()).To(BeFalse())

			kindCluster.AddControlPlaneEndpoint("127.0.0.1", 6334)
			Expect(kindCluster.HasControlPlaneEndpoint()).To(BeTrue())
		})

		It("should correctly handle finalizers", func() {
			kindCluster := &KindCluster{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{
						Time: time.Now(),
					},
				},
			}
			Expect(kindCluster.IsBeingDeleted()).To(BeTrue())
			Expect(kindCluster.HasFinalizer(KindClusterFinalizerName)).To(BeFalse())

			kindCluster.AddFinalizer(KindClusterFinalizerName)
			Expect(kindCluster.HasFinalizer(KindClusterFinalizerName)).To(BeTrue())

			kindCluster.RemoveFinalizer(KindClusterFinalizerName)
			Expect(kindCluster.HasFinalizer(KindClusterFinalizerName)).To(BeFalse())
		})

		It("should serialize and deserialize specscorrectly", func() {
			kindCluster := &KindCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kind-cluster",
					Namespace: "default",
				},
				Spec: KindClusterSpec{
					Nodes: []KindClusterNode{{
						Role:  "control-plane",
						Image: "http://localhost/image:latest",
					}},
					Networking: KindClusterNetworking{
						IPFamily:         KindCLusterIPFamilyV4,
						APIServerAddress: "10.10.10.1",
						APIServerPort:    6334,
					},
					FeatureGates:  map[string]bool{"key": true},
					RuntimeConfig: map[string]string{"key": "value"},
					ControlPlaneEndpoint: KindClusterControlPlaneEndpoint{
						Host: "10.10.10.1",
						Port: 6334,
					},
				},
			}
			bytes, err := yaml.Marshal(kindCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(bytes).NotTo(BeEmpty())

			deserialized := &KindCluster{}
			err = yaml.Unmarshal(bytes, deserialized)
			Expect(err).NotTo(HaveOccurred())
			Expect(deserialized.Namespace).To(Equal(kindCluster.Namespace))
			Expect(deserialized.Name).To(Equal(kindCluster.Name))
			Expect(deserialized.Spec).To(Equal(kindCluster.Spec))
		})

	})

})
