GIT_HASH=$(shell git rev-parse --short HEAD)
IMAGE=ghcr.io/nenorrell/xrai
TAG=latest

.PHONY: all build run tag push deploy clean help

help:
	@echo "xrai - LLM-optimized database schema introspection"
	@echo ""
	@echo "Usage:"
	@echo "  make build      Build Docker image with git hash tag"
	@echo "  make tag        Tag image as latest"
	@echo "  make push       Push latest tag to registry"
	@echo "  make deploy     Build, tag, and push"
	@echo "  make run        Run xrai locally (go run)"
	@echo "  make test       Run tests"
	@echo "  make clean      Remove built binaries"
	@echo ""
	@echo "Variables:"
	@echo "  IMAGE=$(IMAGE)"
	@echo "  TAG=$(TAG)"
	@echo "  GIT_HASH=$(GIT_HASH)"

build:
	docker buildx build -t $(IMAGE):$(GIT_HASH) .

tag: build
	docker tag $(IMAGE):$(GIT_HASH) $(IMAGE):$(TAG)

push:
	docker push $(IMAGE):$(TAG)

deploy: tag push

run:
	go run ./cmd/xrai $(ARGS)

test:
	go test ./...

clean:
	rm -f xrai
	go clean

# Local binary build
bin:
	go build -o xrai ./cmd/xrai
