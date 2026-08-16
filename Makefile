BINARY      := truerepublicd
VERSION     := $(shell git describe --tags --always 2>/dev/null || echo "dev")
LDFLAGS     := -s -w -X main.version=$(VERSION)
BUILD_DIR   := ./build
DETERMINISTIC_TARGET ?= linux-amd64
SOURCE_REF           ?= $(shell git rev-parse HEAD)
DETERMINISTIC_OUT    ?= $(BUILD_DIR)/deterministic/$(DETERMINISTIC_TARGET)

.PHONY: build critical-coverage quality-depth concurrency-replay ibc-two-chain governed-upgrade security-contract go-vuln static-analysis secret-scan deterministic-linux-daemon deterministic-build-contract-test install-lifecycle-contract-test verify test lint clean docker-build docker-up docker-down proto-gen

build:
	@echo "Building $(BINARY)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) ./

critical-coverage:
	./scripts/check-critical-coverage.sh

quality-depth:
	./scripts/check-generative-quality.sh

concurrency-replay:
	TRUEREPUBLIC_CONCURRENCY_REPLAY_SMOKE=1 go test . \
		-run '^TestConcurrentSharedStateReplayRestart$$' \
		-count=1 -timeout=900s -v

ibc-two-chain:
	TRUEREPUBLIC_IBC_TWO_CHAIN_SMOKE=1 ./scripts/go-packages.sh go test \
		-run '^TestIBCTwoChain(TransferAcknowledgementTimeoutReplayRecovery|ChannelCloseTimeoutRecoveryReplacement|CompatibleBinaryRestartRecovery)$$' \
		-count=1 -timeout=900s -v

governed-upgrade:
	TRUEREPUBLIC_MULTI_VALIDATOR_SMOKE=1 go test . \
		-run '^TestGovernedUpgradeMultiValidatorHaltFailureRecovery$$' \
		-count=1 -timeout=720s -v

security-contract:
	go test . -run '^TestSecurityGateRepositoryContract$$' -count=1

go-vuln:
	./scripts/check-go-vulnerabilities.sh
	./scripts/test-go-vulnerability-scan.sh

static-analysis:
	./scripts/check-static-analysis.sh

secret-scan:
	./scripts/check-secret-scan.sh

deterministic-linux-daemon:
	./scripts/build-deterministic-daemon.sh \
		--contract configs/build/deterministic-linux-daemon.json \
		--target "$(DETERMINISTIC_TARGET)" \
		--source-ref "$(SOURCE_REF)" \
		--output-dir "$(DETERMINISTIC_OUT)"

deterministic-build-contract-test:
	./scripts/test-deterministic-daemon.sh

install-lifecycle-contract-test:
	go test ./installlifecycle ./cmd/install-lifecycle

verify:
	@echo "Verifying repository Go packages..."
	./scripts/test-go-packages.sh
	CGO_ENABLED=1 ./scripts/go-packages.sh go build
	./scripts/go-packages.sh go vet
	CGO_ENABLED=1 ./scripts/go-packages.sh go test -race -cover -count=1 -timeout=600s

test:
	@echo "Running tests..."
	./scripts/go-packages.sh go test -race -cover -count=1

lint:
	@echo "Running vet..."
	./scripts/go-packages.sh go vet
	@echo "Running blocking staticcheck..."
	./scripts/check-static-analysis.sh

clean:
	rm -rf $(BUILD_DIR)

docker-build:
	docker build -t truerepublic/node:$(VERSION) -t truerepublic/node:latest .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

proto-gen:
	@echo "Proto generation stub -- add protoc commands when proto files are added"
