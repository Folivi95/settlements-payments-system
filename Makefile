.PHONY: build

LINTER_VERSION=v1.45.2

default: build

get-linter:
	command -v golangci-lint || curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b ${GOPATH}/bin $(LINTER_VERSION)

## lint: download/install golangci-lint and analyse the source code with the configuration in .golangci.yml
lint: get-linter
	golangci-lint run --timeout=10m

unit-tests:
	go fmt ./...
	go test -vet all -shuffle=on --tags=unit ./...

## test-race: run tests with race detection
race-condition-tests:
	go test -v -race ./...

build: mod lint unit-tests race-condition-tests tilt-tests

mod:
	go mod vendor -v

tidy:
	go mod tidy -v

## ==== Tilt make commands ====
cluster:
	@ctlptl apply -f ./local/k8s/kind-cluster.yaml

tilt-deps: cluster
	-@kubectl delete job setup-local-aws -n settlements-payments-system > /dev/null 2>&1 	# we want ensure localstack is healthy, localstack resources are ready by rerunning the setup-local-aws job
	-@kubectl delete job wait-for-kafka > /dev/null 2>&1 									# we want ensure kafka is healthy, topics are ready by rerunning the wait-for-kafka job
	@make remove-ps																			# we only want to run the dependencies therefore remove PS deployment

	@echo " 🎬 Starting dependencies..."
	@echo " 👀 Watch progress here: http://localhost:10350/"
	@tilt ci dependencies > /dev/null

	@echo " ✨  Successfully started dependencies"

tilt-up: cluster
	@echo " 💸 Deploying Payment System..."
	@echo " 👀 Watch progress here: http://localhost:10350/"
	-@kubectl delete job setup-local-aws -n settlements-payments-system > /dev/null 2>&1
	-@kubectl delete job wait-for-kafka > /dev/null 2>&1 # we want ensure kafka is healthy by rerunning the wait-for-kafka job

	@tilt ci ps > /dev/null
	@echo " 🍻 All services running"

tilt-down:
	@echo " 🛑️ Stopping services..."
	@tilt down > /dev/null

tilt-integration-tests:
	-@kubectl delete job integration-tests -n settlements-payments-system > /dev/null 2>&1
	@make remove-ps

	@tilt ci integration-tests --port 0

tilt-service-tests:
	-@kubectl delete job service-tests -n settlements-payments-system > /dev/null 2>&1
	@make tilt-purge-topic

	@make tilt-up

	@tilt ci service-tests --port 0

tilt-tests:
	@make tilt-deps

	@echo " ⌛  Starting Tests"
	@make tilt-integration-tests
	@make tilt-service-tests

	@echo " ✅ ✅ ✅ ✅ ✅ ✅ ✅ ✅ ✅ ✅ ✅ ✅ ✅ "

tilt-purge-topic:
	@echo " 🔖 Moving offsets to end of topics..."
	-@kubectl exec -it deploy/kafka-client -- kafka-consumer-groups.sh --group settlements-payments-system-payment-state-updates-groupId --reset-offsets --to-latest --topic settlements-payments-system-payment-state-updates --bootstrap-server  kafka-headless.default.svc.cluster.local:9092 --execute > /dev/null
	-@kubectl exec -it deploy/kafka-client -- kafka-consumer-groups.sh --group settlements-payments-system-transactions-updates-groupId --reset-offsets --to-latest --topic settlements-payments-system-transactions-updates --bootstrap-server  kafka-headless.default.svc.cluster.local:9092 --execute > /dev/null

remove-ps:
	@if kubectl get deployment settlements-payments-system -n settlements-payments-system  > /dev/null 2>&1; then \
  		echo " 💀 Removing Payment System..." ;\
        kubectl delete deployment settlements-payments-system -n settlements-payments-system  > /dev/null ;\
     	while kubectl rollout status deployment settlements-payments-system -n settlements-payments-system  --timeout=30s > /dev/null 2>&1; do sleep 1; done ;\
    fi

upload-ufx:
	@aws --endpoint-url=http://localhost:4566 s3api put-object --bucket my-bucket --key my-object --body $(abspath)

expose-ps:
	@kubectl port-forward service/settlements-payments-system 8080:8080 --namespace=settlements-payments-system

expose-localstack:
	@kubectl port-forward service/localstack 4566:4566 --namespace=settlements-payments-system

expose-kafka:
	@kubectl port-forward service/kafka-headless  9092 9093

expose-pg:
	@kubectl port-forward service/postgresql-hl 5432 5432 --namespace=settlements-payments-system

expose-fakes:
	@kubectl port-forward service/settlements-fake-providers 8083:8080 --namespace=settlements-fake-providers

expose-services:
	@make expose-ps & \
	make expose-localstack & \
	make expose-kafka & \
	make expose-pg &\
	make expose-fakes && fg
