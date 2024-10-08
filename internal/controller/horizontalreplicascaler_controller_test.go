package controller

import (
	"context"
	"fmt"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	rrethyv1 "github.com/RRethy/horizontalreplicascaler/api/v1"
	"github.com/RRethy/horizontalreplicascaler/internal/stabilization"
)

const (
	consistentlyTimeout    = 1 * time.Second
	eventuallyTimeout      = 2 * time.Second
	interval               = 250 * time.Millisecond
	scalerName             = "test-scaler"
	namespace              = "default"
	deploymentName         = "test-deployment"
	initialDeploymentScale = 10
	initialMinReplicas     = 3
	initialMaxReplicas     = 20
)

var (
	defaultScaleTargetRef           = rrethyv1.ScaleTargetRef{Group: "apps", Kind: "Deployment", Name: deploymentName}
	defaultDeploymentNamespacedName = types.NamespacedName{Name: deploymentName, Namespace: namespace}
	defaultStabilizationKey         = stabilization.KeyFor(namespace, scalerName, deploymentName, defaultScaleTargetRef.Kind, defaultScaleTargetRef.Group)
	defaultScalerNamespacedName     = types.NamespacedName{Name: scalerName, Namespace: namespace}
	defaultDeployment               = &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: deploymentName, Namespace: namespace},
		Spec: appsv1.DeploymentSpec{
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
			ScaleTargetRef: rrethyv1.ScaleTargetRef{Group: "apps", Kind: "Deployment", Name: deploymentName},
			MinReplicas:    initialMinReplicas,
			MaxReplicas:    initialMaxReplicas,
			Metrics:        []rrethyv1.MetricSpec{{Type: "static", Target: rrethyv1.TargetSec{Type: "value", Value: fmt.Sprintf("%d", initialDeploymentScale)}}},
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

			By("Waiting for the deployment to scale based on the default scaler")
			Eventually(func() int32 {
				var deployment appsv1.Deployment
				Expect(k8sClient.Get(ctx, defaultDeploymentNamespacedName, &deployment)).To(Succeed())
				return *deployment.Spec.Replicas
			}, eventuallyTimeout, interval).Should(Equal(int32(initialDeploymentScale)))
		})

		AfterEach(func() {
			By("Cleaning up the scaler")
			Expect(k8sClient.Delete(ctx, defaultHorizontalReplicaScaler)).To(Succeed())

			By("Getting the existing deployment")
			var deployment appsv1.Deployment
			err := k8sClient.Get(ctx, defaultDeploymentNamespacedName, &deployment)
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
				Expect(k8sClient.Get(ctx, defaultDeploymentNamespacedName, &deployment)).To(Succeed())
				return *deployment.Spec.Replicas
			}, eventuallyTimeout, interval).Should(Equal(int32(5)))
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
				Expect(k8sClient.Get(ctx, defaultDeploymentNamespacedName, &deployment)).To(Succeed())
				return *deployment.Spec.Replicas
			}, eventuallyTimeout, interval).Should(Equal(int32(9)))
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
				Expect(k8sClient.Get(ctx, defaultDeploymentNamespacedName, &deployment)).To(Succeed())
				return *deployment.Spec.Replicas
			}, eventuallyTimeout, interval).Should(Equal(int32(5)))
		})

		It("Should respect max replicas", func() {
			By("Getting the existing scaler")
			var horizontalreplicascaler rrethyv1.HorizontalReplicaScaler
			Expect(k8sClient.Get(ctx, defaultScalerNamespacedName, &horizontalreplicascaler)).To(Succeed())

			By("Changing the static metric value to more than max replicas")
			horizontalreplicascaler.Spec.Metrics[0].Target.Value = "20"
			horizontalreplicascaler.Spec.MaxReplicas = 10
			Expect(k8sClient.Update(ctx, &horizontalreplicascaler)).To(Succeed())

			By("Getting the deployment to check the replica count")
			Eventually(func() int32 {
				var deployment appsv1.Deployment
				Expect(k8sClient.Get(ctx, defaultDeploymentNamespacedName, &deployment)).To(Succeed())
				return *deployment.Spec.Replicas
			}, eventuallyTimeout, interval).Should(Equal(int32(10)))
		})

		It("Should not scale if in dry run mode", func() {
			By("Getting the existing scaler")
			var horizontalreplicascaler rrethyv1.HorizontalReplicaScaler
			Expect(k8sClient.Get(ctx, defaultScalerNamespacedName, &horizontalreplicascaler)).To(Succeed())

			By("Setting the scaler to dry run mode")
			horizontalreplicascaler.Spec.DryRun = true

			By("Changing the static metric value to trigger a reconcile")
			horizontalreplicascaler.Spec.Metrics[0].Target.Value = "9"
			Expect(k8sClient.Update(ctx, &horizontalreplicascaler)).To(Succeed())

			By("Getting the status of the scaler to check if the status was updated")
			Eventually(func() int32 {
				var horizontalreplicascaler rrethyv1.HorizontalReplicaScaler
				Expect(k8sClient.Get(ctx, defaultScalerNamespacedName, &horizontalreplicascaler)).To(Succeed())
				return horizontalreplicascaler.Status.DesiredReplicas
			}, eventuallyTimeout, interval).Should(Equal(int32(9)))

			By("Getting the deployment to check the replica count")
			Consistently(func() int32 {
				var deployment appsv1.Deployment
				Expect(k8sClient.Get(ctx, defaultDeploymentNamespacedName, &deployment)).To(Succeed())
				return *deployment.Spec.Replicas
			}, consistentlyTimeout, interval).Should(Equal(int32(initialDeploymentScale)))
		})

		It("Should scale down according to the stabilization window", func() {
			stabilizationWindowDuration := 1 * time.Second

			By("Setting up a fake stabilization window")
			scaleDownStabilizationWindow.Stabilize(defaultStabilizationKey, initialDeploymentScale-2, stabilizationWindowDuration+10*time.Second)
			fakeclock.SetTime(fakeclock.Now().Add(600 * time.Millisecond))
			scaleDownStabilizationWindow.Stabilize(defaultStabilizationKey, initialDeploymentScale-1, stabilizationWindowDuration+10*time.Second)
			scaleDownStabilizationWindow.Stabilize(defaultStabilizationKey, initialDeploymentScale-5, stabilizationWindowDuration+10*time.Second)
			fakeclock.SetTime(fakeclock.Now().Add(500 * time.Millisecond))

			By("Getting the existing scaler")
			var horizontalreplicascaler rrethyv1.HorizontalReplicaScaler
			Expect(k8sClient.Get(ctx, defaultScalerNamespacedName, &horizontalreplicascaler)).To(Succeed())

			By("Changing the scale down stabilization window to 1 second and metric value to less than current")
			horizontalreplicascaler.Spec.ScalingBehavior.ScaleDown.StabilizationWindow = metav1.Duration{Duration: stabilizationWindowDuration}
			horizontalreplicascaler.Spec.Metrics[0].Target.Value = strconv.Itoa(initialDeploymentScale - 4)
			Expect(k8sClient.Update(ctx, &horizontalreplicascaler)).To(Succeed())

			By("Getting the deployment to check the replica count")
			Eventually(func() int32 {
				var deployment appsv1.Deployment
				Expect(k8sClient.Get(ctx, defaultDeploymentNamespacedName, &deployment)).To(Succeed())
				return *deployment.Spec.Replicas
			}, eventuallyTimeout, interval).Should(Equal(int32(initialDeploymentScale - 1)))
		})

		It("Should scale up according to the stabilization window", func() {
			stabilizationWindowDuration := 1 * time.Second

			By("Setting up a fake stabilization window")
			scaleUpStabilizationWindow.Stabilize(defaultStabilizationKey, initialDeploymentScale+2, stabilizationWindowDuration+10*time.Second)
			fakeclock.SetTime(fakeclock.Now().Add(600 * time.Millisecond))
			scaleUpStabilizationWindow.Stabilize(defaultStabilizationKey, initialDeploymentScale+1, stabilizationWindowDuration+10*time.Second)
			scaleUpStabilizationWindow.Stabilize(defaultStabilizationKey, initialDeploymentScale+5, stabilizationWindowDuration+10*time.Second)
			fakeclock.SetTime(fakeclock.Now().Add(500 * time.Millisecond))

			By("Getting the existing scaler")
			var horizontalreplicascaler rrethyv1.HorizontalReplicaScaler
			Expect(k8sClient.Get(ctx, defaultScalerNamespacedName, &horizontalreplicascaler)).To(Succeed())

			By("Changing the scale up stabilization window to 1 second and metric value to greater than current")
			horizontalreplicascaler.Spec.ScalingBehavior.ScaleUp.StabilizationWindow = metav1.Duration{Duration: stabilizationWindowDuration}
			horizontalreplicascaler.Spec.Metrics[0].Target.Value = strconv.Itoa(initialDeploymentScale + 4)
			Expect(k8sClient.Update(ctx, &horizontalreplicascaler)).To(Succeed())

			By("Getting the deployment to check the replica count")
			Eventually(func() int32 {
				var deployment appsv1.Deployment
				Expect(k8sClient.Get(ctx, defaultDeploymentNamespacedName, &deployment)).To(Succeed())
				return *deployment.Spec.Replicas
			}, eventuallyTimeout, interval).Should(Equal(int32(initialDeploymentScale + 1)))
		})

		It("Should not scale down if stabilization window is not met", func() {
			stabilizationWindowDuration := 1 * time.Second

			By("Setting up a fake stabilization window")
			scaleDownStabilizationWindow.Stabilize(defaultStabilizationKey, initialDeploymentScale-2, stabilizationWindowDuration+10*time.Second)
			scaleDownStabilizationWindow.Stabilize(defaultStabilizationKey, initialDeploymentScale-1, stabilizationWindowDuration+10*time.Second)
			scaleDownStabilizationWindow.Stabilize(defaultStabilizationKey, initialDeploymentScale-5, stabilizationWindowDuration+10*time.Second)
			fakeclock.SetTime(fakeclock.Now().Add(500 * time.Millisecond))

			By("Getting the existing scaler")
			var horizontalreplicascaler rrethyv1.HorizontalReplicaScaler
			Expect(k8sClient.Get(ctx, defaultScalerNamespacedName, &horizontalreplicascaler)).To(Succeed())

			By("Changing the scale down stabilization window to 1 second and metric value to less than current")
			horizontalreplicascaler.Spec.ScalingBehavior.ScaleDown.StabilizationWindow = metav1.Duration{Duration: stabilizationWindowDuration}
			horizontalreplicascaler.Spec.Metrics[0].Target.Value = strconv.Itoa(initialDeploymentScale - 4)
			Expect(k8sClient.Update(ctx, &horizontalreplicascaler)).To(Succeed())

			By("Getting the deployment to check the replica count")
			Consistently(func() int32 {
				var deployment appsv1.Deployment
				Expect(k8sClient.Get(ctx, defaultDeploymentNamespacedName, &deployment)).To(Succeed())
				return *deployment.Spec.Replicas
			}, consistentlyTimeout, interval).Should(Equal(int32(initialDeploymentScale)))
		})

		It("Should not scale up if stabilization window is not met", func() {
			stabilizationWindowDuration := 1 * time.Second

			By("Setting up a fake stabilization window")
			scaleUpStabilizationWindow.Stabilize(defaultStabilizationKey, initialDeploymentScale+2, stabilizationWindowDuration+10*time.Second)
			scaleUpStabilizationWindow.Stabilize(defaultStabilizationKey, initialDeploymentScale+1, stabilizationWindowDuration+10*time.Second)
			scaleUpStabilizationWindow.Stabilize(defaultStabilizationKey, initialDeploymentScale+5, stabilizationWindowDuration+10*time.Second)
			fakeclock.SetTime(fakeclock.Now().Add(500 * time.Millisecond))

			By("Getting the existing scaler")
			var horizontalreplicascaler rrethyv1.HorizontalReplicaScaler
			Expect(k8sClient.Get(ctx, defaultScalerNamespacedName, &horizontalreplicascaler)).To(Succeed())

			By("Changing the scale up stabilization window to 1 second and metric value to greater than current")
			horizontalreplicascaler.Spec.ScalingBehavior.ScaleUp.StabilizationWindow = metav1.Duration{Duration: stabilizationWindowDuration}
			horizontalreplicascaler.Spec.Metrics[0].Target.Value = strconv.Itoa(initialDeploymentScale + 4)
			Expect(k8sClient.Update(ctx, &horizontalreplicascaler)).To(Succeed())

			By("Getting the deployment to check the replica count")
			Consistently(func() int32 {
				var deployment appsv1.Deployment
				Expect(k8sClient.Get(ctx, defaultDeploymentNamespacedName, &deployment)).To(Succeed())
				return *deployment.Spec.Replicas
			}, consistentlyTimeout, interval).Should(Equal(int32(initialDeploymentScale)))
		})

		It("Should not scale if metric is oscillating and we have up and down stabilization", func() {
			stabilizationWindowDuration := 1 * time.Second

			By("Setting up a fake stabilization window")
			for _, value := range []int32{initialDeploymentScale - 2, initialDeploymentScale + 2} {
				scaleDownStabilizationWindow.Stabilize(defaultStabilizationKey, value, stabilizationWindowDuration+10*time.Second)
				scaleUpStabilizationWindow.Stabilize(defaultStabilizationKey, value, stabilizationWindowDuration+10*time.Second)
			}
			fakeclock.SetTime(fakeclock.Now().Add(600 * time.Millisecond))
			for _, value := range []int32{initialDeploymentScale - 1, initialDeploymentScale + 1, initialDeploymentScale - 5, initialDeploymentScale + 5} {
				scaleDownStabilizationWindow.Stabilize(defaultStabilizationKey, value, stabilizationWindowDuration+10*time.Second)
				scaleUpStabilizationWindow.Stabilize(defaultStabilizationKey, value, stabilizationWindowDuration+10*time.Second)
			}
			fakeclock.SetTime(fakeclock.Now().Add(500 * time.Millisecond))

			By("Getting the existing scaler")
			var horizontalreplicascaler rrethyv1.HorizontalReplicaScaler
			Expect(k8sClient.Get(ctx, defaultScalerNamespacedName, &horizontalreplicascaler)).To(Succeed())

			By("Changing the scale down stabilization window to 1 second and metric value to greater than current")
			horizontalreplicascaler.Spec.ScalingBehavior.ScaleDown.StabilizationWindow = metav1.Duration{Duration: stabilizationWindowDuration}
			horizontalreplicascaler.Spec.ScalingBehavior.ScaleUp.StabilizationWindow = metav1.Duration{Duration: stabilizationWindowDuration}
			horizontalreplicascaler.Spec.Metrics[0].Target.Value = strconv.Itoa(initialDeploymentScale + 4)
			Expect(k8sClient.Update(ctx, &horizontalreplicascaler)).To(Succeed())

			By("Getting the deployment to check the replica count")
			Consistently(func() int32 {
				var deployment appsv1.Deployment
				Expect(k8sClient.Get(ctx, defaultDeploymentNamespacedName, &deployment)).To(Succeed())
				return *deployment.Spec.Replicas
			}, consistentlyTimeout, interval).Should(Equal(int32(initialDeploymentScale)))

			By("Changing the metric value to less than current")
			horizontalreplicascaler.Spec.Metrics[0].Target.Value = strconv.Itoa(initialDeploymentScale - 4)
			Expect(k8sClient.Update(ctx, &horizontalreplicascaler)).To(Succeed())

			By("Getting the deployment to check the replica count")
			Consistently(func() int32 {
				var deployment appsv1.Deployment
				Expect(k8sClient.Get(ctx, defaultDeploymentNamespacedName, &deployment)).To(Succeed())
				return *deployment.Spec.Replicas
			}, consistentlyTimeout, interval).Should(Equal(int32(initialDeploymentScale)))
		})

		It("Should not scale if desired replicas are equal to target", func() {
			stabilizationWindowDuration := 1 * time.Second

			By("Setting up a fake stabilization window")
			scaleUpStabilizationWindow.Stabilize(defaultStabilizationKey, initialDeploymentScale+2, stabilizationWindowDuration+10*time.Second)
			fakeclock.SetTime(fakeclock.Now().Add(600 * time.Millisecond))
			scaleUpStabilizationWindow.Stabilize(defaultStabilizationKey, initialDeploymentScale+1, stabilizationWindowDuration+10*time.Second)
			scaleUpStabilizationWindow.Stabilize(defaultStabilizationKey, initialDeploymentScale+5, stabilizationWindowDuration+10*time.Second)
			fakeclock.SetTime(fakeclock.Now().Add(500 * time.Millisecond))

			By("Getting the existing scaler")
			var horizontalreplicascaler rrethyv1.HorizontalReplicaScaler
			Expect(k8sClient.Get(ctx, defaultScalerNamespacedName, &horizontalreplicascaler)).To(Succeed())

			By("Changing the scale up stabilization window to 1 second and metric value to equal to current")
			horizontalreplicascaler.Spec.ScalingBehavior.ScaleUp.StabilizationWindow = metav1.Duration{Duration: stabilizationWindowDuration}
			horizontalreplicascaler.Spec.Metrics[0].Target.Value = strconv.Itoa(initialDeploymentScale)
			Expect(k8sClient.Update(ctx, &horizontalreplicascaler)).To(Succeed())

			By("Getting the deployment to check the replica count")
			Consistently(func() int32 {
				var deployment appsv1.Deployment
				Expect(k8sClient.Get(ctx, defaultDeploymentNamespacedName, &deployment)).To(Succeed())
				return *deployment.Spec.Replicas
			}, consistentlyTimeout, interval).Should(Equal(int32(initialDeploymentScale)))
		})
	})
})
