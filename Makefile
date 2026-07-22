BINARY      := truerepublicd
VERSION     := $(shell git describe --tags --always 2>/dev/null || echo "dev")
LDFLAGS     := -s -w -X main.version=$(VERSION)
BUILD_DIR   := ./build

.PHONY: build install verify test lint clean docker-build docker-up docker-down proto-gen

build:
	@echo "Building $(BINARY)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) ./

install:
	@echo "Installing $(BINARY)..."
	go install -ldflags="$(LDFLAGS)" ./

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
	@echo "Running staticcheck (if installed)..."
	-./scripts/go-packages.sh staticcheck

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
