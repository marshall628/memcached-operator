package e2e

import (
	"context"
	"fmt"
	"github.com/marshall628/memcached-operator/pkg/apis"
	operator "github.com/marshall628/memcached-operator/pkg/apis/cache/v1alpha1"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
	"time"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestMemcached(t *testing.T) {
	memcachedList := &operator.MemcachedList{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, memcachedList)
	if err != nil {
		t.Fatalf("Failed to add custom resource scheme to framework: %v", err)
	}

	// run subtest
	t.Run("memcached-group", func(t *testing.T) {
		t.Run("Cluster", MemcachedCluster)
		t.Run("Cluster2", MemcachedCluster)
	})
}

func memcachedScaleTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("Couldn't get namespace: %v", err)
	}

	exampleMemcached := &operator.Memcached{
		ObjectMeta: metav1.ObjectMeta{
			Name: "example-memcached",
			Namespace: namespace,
		},
		Spec: operator.MemcachedSpec{
			Size: 3,
		},
	}

	err = f.Client.Create(context.TODO(), exampleMemcached, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	// wait for example-memcached to reach 3 replicas
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-memcached", 3, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = f.Client.Get(context.TODO(), types.NamespacedName{Name: "example-memcached", Namespace: namespace}, exampleMemcached)
	if err != nil {
		return err
	}

	exampleMemcached.Spec.Size = 4
	err = f.Client.Update(context.TODO(), exampleMemcached)
	if err != nil {
		return err
	}

	// wait for example-memcached to reach 4 replicas
	return e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "example-memcached", 4, retryInterval, timeout)
}

func MemcachedCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()

	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})

	if err != nil {
		t.Fatalf("Failed to initialize cluster resources :%v", err)
	}

	t.Log("Initialize cluster resources")
	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}

	f := framework.Global
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "memcached-operator", 1, retryInterval, timeout)
}