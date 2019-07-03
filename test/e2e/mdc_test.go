package e2e

import (
	"testing"
	"time"

	apis "github.com/aerogear/mobile-developer-console-operator/pkg/apis"
	mdcv1alpha1 "github.com/aerogear/mobile-developer-console-operator/pkg/apis/mdc/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
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

	err := ctx.InitializeClusterResources(&framework.CleanupOptions{
		TestContext:   ctx,
		Timeout:       cleanupTimeout,
		RetryInterval: cleanupRetryInterval,
	})

	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Successfully initialized cluster resources")

	namespace, err := ctx.GetNamespace()

	f := framework.Global
	if err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "mobile-developer-console-operator", 1, retryInterval, timeout); err != nil {
		t.Fatal(err)
	}

}
