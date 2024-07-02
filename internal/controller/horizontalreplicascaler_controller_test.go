package controller

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	rrethyv1 "github.com/RRethy/horizontalrpelicascaler/api/v1"
)

var _ = Describe("HorizontalReplicaScaler Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"
		const namespace = "default"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: namespace,
		}
		horizontalreplicascaler := &rrethyv1.HorizontalReplicaScaler{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind HorizontalReplicaScaler")
			err := k8sClient.Get(ctx, typeNamespacedName, horizontalreplicascaler)
			if err != nil && errors.IsNotFound(err) {
				resource := &rrethyv1.HorizontalReplicaScaler{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			resource := &rrethyv1.HorizontalReplicaScaler{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance HorizontalReplicaScaler")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
	})
})
