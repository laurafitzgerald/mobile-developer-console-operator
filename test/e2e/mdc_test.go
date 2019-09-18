package e2e

import (
	goctx "context"
	"testing"
	"time"

	apis "github.com/aerogear/mobile-developer-console-operator/pkg/apis"
	mdcv1alpha1 "github.com/aerogear/mobile-developer-console-operator/pkg/apis/mdc/v1alpha1"
	operator "github.com/aerogear/mobile-developer-console-operator/pkg/apis/mdc/v1alpha1"
	dcv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 200
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestMdc(t *testing.T) {
	mdcList := &mdcv1alpha1.MobileDeveloperConsoleList{}
	if err := framework.AddToFrameworkScheme(apis.AddToScheme, mdcList); err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	t.Run("mdc-e2e", MdcTest)
}

func MdcTest(t *testing.T) {
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	f := framework.Global
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatalf("failed to get namespace: %v", err)
	}

	mdcAppName := "mdc"
	mdcTestCR := &mdcv1alpha1.MobileDeveloperConsole{
		TypeMeta: metav1.TypeMeta{
			Kind:       "MobileDeveloperConsole",
			APIVersion: "mdc.aerogear.org/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      mdcAppName,
			Namespace: namespace,
		},
		Spec: operator.MobileDeveloperConsoleSpec{
			OAuthClientId:     "mobile-developer-console",
			OAuthClientSecret: "dummy-client-secret",
		},
	}

	if err := initializeMDCResources(t, f, ctx, namespace); err != nil {
		t.Fatal(err)
	}

	// Create MDC CR
	if err := createMDCCustomResource(t, f, ctx, mdcTestCR); err != nil {
		t.Fatal(err)
	}

	// Additional client needed for retrieving deploymentConfigs
	dcV1Client, err := dcv1.NewForConfig(f.KubeConfig)
	if err != nil {
		t.Fatalf("Failed to initialize DeploymentConfig Client : %v", err)
	}

	// Ensure MDC was deployed successfully
	if err := waitForDeploymentConfig(t, *dcV1Client, namespace, mdcAppName, 1); err != nil {
		t.Fatal(err)
	}
	t.Log("UPS deployment was successful")

	// Delete MDC CR
	if err := deleteMDCCustomResource(t, f, ctx, mdcTestCR); err != nil {
		t.Fatal(err)
	}

	// Ensure MDC was deleted successfully
	if err := waitForDeploymentConfig(t, *dcV1Client, namespace, mdcAppName, 0); err != nil {
		t.Fatal(err)
	}
	t.Log("UPS was deleted successfully")

}

func initializeMDCResources(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, namespace string) error {
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       cleanupTimeout,
		RetryInterval: cleanupRetryInterval,
	})

	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Successfully initialized cluster resources")

	if err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "mobile-developer-console-operator", 1, retryInterval, timeout); err != nil {
		t.Fatal(err)
	}

	t.Log("MDC Operator successfully deployed")

	return nil
}

func createMDCCustomResource(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, testCr *mdcv1alpha1.MobileDeveloperConsole) error {

	err := f.Client.Create(goctx.TODO(), testCr, &framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       cleanupTimeout,
		RetryInterval: cleanupRetryInterval,
	})
	if err != nil {
		return err
	}
	t.Log("Successfully created MDC Custom Resource")

	return nil
}

func deleteMDCCustomResource(t *testing.T, f *framework.Framework, ctx *framework.TestCtx, testCr *mdcv1alpha1.MobileDeveloperConsole) error {

	err := f.Client.Delete(goctx.TODO(), testCr)
	if err != nil {
		return err
	}
	t.Log("Successfully deleted MDC Custom Resource")

	return nil
}

// Helper function for checking whether specified DeploymentConfig has a certain number of available replicas
// Copied & edited from https://github.com/operator-framework/operator-sdk/blob/f6d83791dd8880f0e33d549343642aabadc9d3a0/pkg/test/e2eutil/wait_util.go#L46
func waitForDeploymentConfig(t *testing.T, dcV1Client dcv1.AppsV1Client, namespace, name string, replicas int) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		dc, err := dcV1Client.DeploymentConfigs(namespace).Get(name, metav1.GetOptions{IncludeUninitialized: true})

		if err != nil {
			if apierrors.IsNotFound(err) && replicas == 0 {
				return true, nil
			}
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s Deployment Config\n", name)
				return false, nil
			}
			return false, err
		}

		if int(dc.Status.AvailableReplicas) == replicas {
			return true, nil
		}
		t.Logf("Waiting for full availability of %s Deployment Config (%d/%d)\n", name, dc.Status.AvailableReplicas, replicas)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("Deployment Config has now requested number of replicas: (%d/%d)\n", replicas, replicas)
	return nil
}
