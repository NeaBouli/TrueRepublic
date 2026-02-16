BINARY      := truerepublicd
VERSION     := $(shell git describe --tags --always 2>/dev/null || echo "dev")
LDFLAGS     := -s -w -X main.version=$(VERSION)
BUILD_DIR   := ./build

.PHONY: build install test lint clean docker-build docker-up docker-down proto-gen

build:
	@echo "Building $(BINARY)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) ./

install:
	@echo "Installing $(BINARY)..."
	go install -ldflags="$(LDFLAGS)" ./

test:
	@echo "Running tests..."
	go test ./... -race -cover -count=1

lint:
	@echo "Running vet..."
	go vet ./...
	@echo "Running staticcheck (if installed)..."
	-staticcheck ./...

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
