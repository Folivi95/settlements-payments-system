load("./local/postgres/Tiltfile", "POSTGRES_RESOURCE")
load("./local/localstack/Tiltfile", "LOCALSTACK_RESOURCE")
load("./local/setup-localstack/Tiltfile", "LOCALSTACK_JOB_RESOURCES")
load("./local/wait-for-kafka/Tiltfile", "KAFKA_JOB_RESOURCES")
load("./local/integration-tests/Tiltfile", "INT_JOB_RESOURCES")
load("./local/service-tests/Tiltfile", "SERV_JOB_RESOURCES")
load("./local/kafka-client/Tiltfile", "KAFKA_CLIENT_RESOURCE")

v1alpha1.extension_repo(name='settlements-fake-providers', url='https://github.com/saltpay/settlements-fake-providers')
v1alpha1.extension(name='settlements-fake-providers', repo_name='settlements-fake-providers')
load('ext://settlements-fake-providers', 'FAKE_RESOURCES')

# generate vendor folder
local('make mod', quiet=True, echo_off=True)

SPS = ["settlements-payments-system"]
DEPENDENCIES = LOCALSTACK_RESOURCE + POSTGRES_RESOURCE + LOCALSTACK_JOB_RESOURCES + FAKE_RESOURCES + KAFKA_JOB_RESOURCES + KAFKA_CLIENT_RESOURCE

groups = {
    'dependencies': DEPENDENCIES,
    'ps': SPS + DEPENDENCIES,
    'integration-tests': INT_JOB_RESOURCES,
    'service-tests': SERV_JOB_RESOURCES
}
config.define_string_list("to-run", args=True)
cfg = config.parse()

resources = []
for arg in cfg.get('to-run', []):
  if arg in groups:
    resources += groups[arg]
config.set_enabled_resources(resources)

k8s_yaml("./local/k8s/namespace.yaml")
k8s_yaml("./local/k8s/alertnamespace.yaml")
secret_settings(disable_scrub=True)
k8s_yaml("./local/k8s/secrets.yaml")


docker_build(
    "localhost:35000/settlements-payments-system",
    ".",
)
docker_build(
    "localhost:35000/tests-payments-system",
    ".",
    dockerfile="Tests.Dockerfile",
)

k8s_yaml(
    helm(
        "./chart",
        namespace="settlements-payments-system",
        name="settlements-payments-system",
        values=["local/k8s/values.yaml"],
        set=[
                "serviceMonitor.enabled=false",
                "serviceAccount.create=true",
        ],
    )
)

k8s_resource(
    "settlements-payments-system",
    labels=["settlements-payments-system"],
    resource_deps=DEPENDENCIES
)
