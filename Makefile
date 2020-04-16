GO ?= go
GOLANGCI_LINT ?= golangci-lint

LDFLAGS ?= -s -w
ifdef COMMIT
LDFLAGS += -X github.com/ethersphere/bee.commit="$(COMMIT)"
endif

.PHONY: all
all: build lint vet test binary

.PHONY: binary
binary: export CGO_ENABLED=0
binary: dist FORCE
	$(GO) version
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" -o dist/bee ./cmd/bee

dist:
	mkdir $@

.PHONY: lint
lint:
	$(GOLANGCI_LINT) run

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: test
test:
	$(GO) test -v ./...

.PHONY: build
build: export CGO_ENABLED=0
build:
	$(GO) build -trimpath -ldflags "$(LDFLAGS)" ./...

.PHONY: clean
clean:
	$(GO) clean
	rm -rf dist/

check-swagger:
	which swagger || (GO111MODULE=off go get -u github.com/go-swagger/go-swagger/cmd/swagger)

swagger-api: check-swagger
	swagger generate spec -w ./pkg/api -o ./docs/api/swagger.yaml --scan-models

swagger-debugapi: check-swagger
	swagger generate spec -w ./pkg/debugapi -o ./docs/debugapi/swagger.yaml --scan-models

serve-swagger-api: check-swagger
	swagger serve docs/api/swagger.yaml

serve-swagger-debugapi: check-swagger
	swagger serve docs/debugapi/swagger.yaml

FORCE:
