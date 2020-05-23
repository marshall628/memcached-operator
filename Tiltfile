allow_k8s_contexts('sysflow/api-marshall-os-fyre-ibm-com:6443/maryang')
k8s_yaml(['deploy/operator.yaml', 'deploy/role_binding.yaml', 'deploy/role.yaml', 'deploy/service_account.yaml'])
k8s_yaml('deploy/crds/cache.example.com_memcacheds_crd.yaml')
k8s_yaml('deploy/crds/cache.example.com_v1alpha1_memcached_cr.yaml')

custom_build(
  'quay.io/marshall628/memcached-operator',
  'operator-sdk build $EXPECTED_REF',
  ['cmd/manager', 'pkg', 'go.mod', 'go.sum', 'version', 'tools.go'],
  'v0.0.1',
  ignore=['apis/cache/v1alpha1/zz_generated.deepcopy.go'],
)

local_resource('Regenerate CRDS', 'operator-sdk generate k8s && operator-sdk generate crds', deps='deploy/crds/cache.example.com_memcacheds_crd.yaml', trigger_mode=TRIGGER_MODE_MANUAL)
local_resource('Run unit test', ["go", "test", "github.com/marshall628/memcached-operator/pkg/controller/memcached"], deps="pkg/controller/memcached/")
local_resource('Run end to end test', 'operator-sdk test local $(pwd)/test/e2e --operator-namespace operator-test --up-local --verbose --skip-cleanup-error', deps="test/e2e")


