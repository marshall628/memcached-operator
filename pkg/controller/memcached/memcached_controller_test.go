package memcached

import (
	"context"
	cachev1alpha1 "github.com/marshall628/memcached-operator/pkg/apis/cache/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"math/rand"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strconv"
	"testing"
)

func TestMemcachedController(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	var (
		name = "memcached-operator"
		namespace = "memcached"
		replicas int32 = 3
	)

	memcached := &cachev1alpha1.Memcached{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Namespace: namespace,
		},
		Spec: cachev1alpha1.MemcachedSpec{
			Size: replicas,
		},
	}

	objs := []runtime.Object{
		memcached,
	}

	s := scheme.Scheme
	s.AddKnownTypes(cachev1alpha1.SchemeGroupVersion, memcached)
	cl := fake.NewFakeClient(objs...)
	r := &ReconcileMemcached{client: cl, scheme: s}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName {
			Name: name,
			Namespace: namespace,
		},
	}
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	// Check if deployment has been created and with correct size
	dep := &appsv1.Deployment{}
	err = cl.Get(context.TODO(), req.NamespacedName, dep)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}

	dsize := *dep.Spec.Replicas
	if dsize != replicas {
		t.Errorf("dep size (%d) is not the expected size (%d)", dsize, replicas)
	}

	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	// Check the result of reconciliation to make sure it is under correct state
	if res.Requeue {
		t.Error("reconcile requeue request which is not expected")
	}

	// Check if service has been created
	ser := &corev1.Service{}
	err = cl.Get(context.TODO(), req.NamespacedName, ser)
	if err != nil {
		t.Fatalf("get service: (%v)", err)
	}

	// Create 3 expected pods in namespace and get their names
	podLabels := labelsForMemcached(name)
	pod := corev1.Pod {
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Labels: podLabels,
		},
	}
	podNames := make([]string, 3)
	for i := 0; i < 3; i++ {
		pod.ObjectMeta.Name = name + ".pod." + strconv.Itoa(rand.Int())
		podNames[i] = pod.ObjectMeta.Name
		if err = cl.Create(context.TODO(), pod.DeepCopy()); err != nil {
			t.Fatalf("Create pod %d: (%v)", i, err)
		}
	}

	// Reconcile again so checks pods and update memcached resources status
	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if res != (reconcile.Result{}) {
		t.Error("reconcile did not requeue request as expected")
	}

	// Get the updated Memcached object
	memcached = &cachev1alpha1.Memcached{}
	err = r.client.Get(context.TODO(), req.NamespacedName, memcached)
	if err != nil {
		t.Errorf("get memcached: (%v)", err)
	}

	nodes := memcached.Status.Nodes
	if !reflect.DeepEqual(podNames, nodes) {
		t.Errorf("pod names %v did not match expected %v", nodes, podNames)
	}
}