package controller

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/scale"
	"k8s.io/client-go/tools/record"
	clock "k8s.io/utils/clock/testing"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	rrethyv1 "github.com/RRethy/horizontalreplicascaler/api/v1"
	"github.com/RRethy/horizontalreplicascaler/internal/stabilization"
	// +kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var (
	cfg                          *rest.Config
	k8sClient                    client.Client
	scaleClient                  scale.ScalesGetter
	eventRecorder                *record.FakeRecorder
	testEnv                      *envtest.Environment
	ctx                          context.Context
	cancel                       context.CancelFunc
	fakeclock                    *clock.FakeClock
	scaleDownStabilizationWindow *stabilization.Window
	scaleUpStabilizationWindow   *stabilization.Window
)

func TestControllers(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("Bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd")},
		ErrorIfCRDPathMissing: true,
		BinaryAssetsDirectory: filepath.Join("..", "..", "bin", "k8s",
			fmt.Sprintf("1.30.0-%s-%s", runtime.GOOS, runtime.GOARCH)),
	}

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = rrethyv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	// +kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{Scheme: scheme.Scheme})
	Expect(err).ToNot(HaveOccurred())

	eventRecorder = record.NewFakeRecorder(10)

	clientset, err := kubernetes.NewForConfig(k8sManager.GetConfig())
	Expect(err).ToNot(HaveOccurred())

	scaleKindResolver := scale.NewDiscoveryScaleKindResolver(clientset.Discovery())
	scaleClient, err = scale.NewForConfig(k8sManager.GetConfig(), k8sManager.GetRESTMapper(), dynamic.LegacyAPIPathResolverFunc, scaleKindResolver)
	Expect(err).ToNot(HaveOccurred())

	fakeclock = clock.NewFakeClock(time.Date(1997, time.November, 7, 0, 0, 0, 0, time.UTC))

	scaleDownStabilizationWindow = stabilization.NewWindow(stabilization.MaxRollingWindow, stabilization.WithClock(fakeclock))
	scaleUpStabilizationWindow = stabilization.NewWindow(stabilization.MinRollingWindow, stabilization.WithClock(fakeclock))

	err = (&HorizontalReplicaScalerReconciler{
		Client:                       k8sManager.GetClient(),
		Scheme:                       k8sManager.GetScheme(),
		Recorder:                     eventRecorder,
		ScaleClient:                  scaleClient,
		ScaleDownStabilizationWindow: scaleDownStabilizationWindow,
		ScaleUpStabilizationWindow:   scaleUpStabilizationWindow,
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
