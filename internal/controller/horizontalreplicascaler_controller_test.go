package controller

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"

	rrethyv1 "github.com/RRethy/horizontalreplicascaler/api/v1"
)

const (
	timeout        = time.Second * 20
	interval       = time.Millisecond * 250
	scalerName     = "test-scaler"
	namespace      = "default"
	deploymentName = "test-deployment"
)

var (
	defaultScalerNamespacedName = types.NamespacedName{
		Name:      scalerName,
		Namespace: namespace,
	}
	defaultDeployment = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: namespace},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To[int32](3),
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "test"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "busybox", Image: "busybox", Command: []string{"sleep", "3600"}}},
				},
			},
		},
	}
	defaultHorizontalReplicaScaler = &rrethyv1.HorizontalReplicaScaler{
		ObjectMeta: metav1.ObjectMeta{Name: scalerName, Namespace: namespace},
		Spec: rrethyv1.HorizontalReplicaScalerSpec{
			ScaleTargetRef: &rrethyv1.ScaleTargetRef{Group: "apps", Kind: "Deployment", Name: deploymentName},
			MinReplicas:    3,
			MaxReplicas:    10,
			Metrics:        []rrethyv1.MetricSpec{{Type: "static", Target: rrethyv1.TargetSec{Type: "value", Value: "10"}}},
		},
	}
)

var _ = Describe("HorizontalReplicaScaler Controller", func() {
	Context("When scaling a Deployment", func() {
		ctx := context.Background()

		BeforeEach(func() {
			By("Creating a default deployment to scale")
			Expect(k8sClient.Create(ctx, defaultDeployment.DeepCopy())).To(Succeed())

			By("Creating a new custom resource for the Kind HorizontalReplicaScaler")
			Expect(k8sClient.Create(ctx, defaultHorizontalReplicaScaler.DeepCopy())).To(Succeed())
		})

		AfterEach(func() {
			By("Cleaning up the scaler")
			Expect(k8sClient.Delete(ctx, defaultHorizontalReplicaScaler)).To(Succeed())

			By("Getting the existing deployment")
			var deployment appsv1.Deployment
			err := k8sClient.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: namespace}, &deployment)
			Expect(err).To(SatisfyAny(BeNil(), WithTransform(errors.IsNotFound, BeTrue())))

			if err == nil {
				By("Cleaning up the deployment")
				Expect(k8sClient.Delete(ctx, &deployment)).To(Succeed())
			}
		})

		It("Should change the replica count to the static value", func() {
			By("Getting the existing scaler")
			var horizontalreplicascaler rrethyv1.HorizontalReplicaScaler
			Expect(k8sClient.Get(ctx, defaultScalerNamespacedName, &horizontalreplicascaler)).To(Succeed())

			By("Changing the static metric value in the scaler to trigger a reconcile")
			horizontalreplicascaler.Spec.Metrics[0].Target.Value = "5"
			Expect(k8sClient.Update(ctx, &horizontalreplicascaler)).To(Succeed())

			By("Getting the deployment to check the replica count")
			Eventually(func() int32 {
				var deployment appsv1.Deployment
				err := k8sClient.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: namespace}, &deployment)
				Expect(err).ToNot(HaveOccurred())
				return *deployment.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(5)))
		})

		It("Should create an event if the scale subresource does not exist", func() {
			By("Changing the target name to a non-existent deployment")
			var horizontalreplicascaler rrethyv1.HorizontalReplicaScaler
			Expect(k8sClient.Get(ctx, defaultScalerNamespacedName, &horizontalreplicascaler)).To(Succeed())
			nonExistentDeploymentName := "non-existent-deployment"
			horizontalreplicascaler.Spec.ScaleTargetRef.Name = nonExistentDeploymentName
			Expect(k8sClient.Update(ctx, &horizontalreplicascaler)).To(Succeed())

			By("Checking if the event was recorded")
			Eventually(eventRecorder.Events).Should(Receive(ContainSubstring(fmt.Sprintf("\"%s\" not found", nonExistentDeploymentName))))
		})

		It("Should take the max of the metrics", func() {
			By("Getting the existing scaler")
			var horizontalreplicascaler rrethyv1.HorizontalReplicaScaler
			Expect(k8sClient.Get(ctx, defaultScalerNamespacedName, &horizontalreplicascaler)).To(Succeed())

			By("Adding a new metric to the scaler")
			horizontalreplicascaler.Spec.Metrics = []rrethyv1.MetricSpec{
				{Type: "static", Target: rrethyv1.TargetSec{Type: "value", Value: "9"}},
				{Type: "static", Target: rrethyv1.TargetSec{Type: "value", Value: "7"}},
			}
			Expect(k8sClient.Update(ctx, &horizontalreplicascaler)).To(Succeed())

			By("Getting the deployment to check the replica count")
			Eventually(func() int32 {
				var deployment appsv1.Deployment
				err := k8sClient.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: namespace}, &deployment)
				Expect(err).ToNot(HaveOccurred())
				return *deployment.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(9)))
		})

		It("Should respect min replicas", func() {
			By("Getting the existing scaler")
			var horizontalreplicascaler rrethyv1.HorizontalReplicaScaler
			Expect(k8sClient.Get(ctx, defaultScalerNamespacedName, &horizontalreplicascaler)).To(Succeed())

			By("Changing the static metric value to less than min replicas")
			horizontalreplicascaler.Spec.Metrics[0].Target.Value = "2"
			horizontalreplicascaler.Spec.MinReplicas = 5
			Expect(k8sClient.Update(ctx, &horizontalreplicascaler)).To(Succeed())

			By("Getting the deployment to check the replica count")
			Eventually(func() int32 {
				var deployment appsv1.Deployment
				err := k8sClient.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: namespace}, &deployment)
				Expect(err).ToNot(HaveOccurred())
				return *deployment.Spec.Replicas
			}, timeout, interval).Should(Equal(int32(5)))
		})
	})
})
