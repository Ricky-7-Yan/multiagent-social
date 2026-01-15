BINARY = multiagent

.PHONY: build run docker-up migrate test lint

build:
	go build -v -o $(BINARY) ./cmd/server

run: build
	./$(BINARY)

docker-up:
	docker compose -f deployments/docker/docker-compose.yml up --build

migrate:
	@echo "Run migrations (manual step):"
	@echo "  psql \"$(PG_DSN)\" -f migrations/001_init.sql"

test:
	go test ./...

lint:
	golangci-lint run ./...

frontend-build:
	cd web/react-app && npm ci && npm run build

docker-build:
	# Build Docker image locally. Set IMAGE_NAME and TAG env vars to customize.
	docker build -t ${IMAGE_NAME:-multiagent-social}:${TAG:-latest} .

docker-push:
	# Push image to registry. Set REGISTRY and IMAGE_NAME and TAG.
ifdef REGISTRY
	docker tag ${IMAGE_NAME:-multiagent-social}:${TAG:-latest} ${REGISTRY}/${IMAGE_NAME:-multiagent-social}:${TAG:-latest}
	docker push ${REGISTRY}/${IMAGE_NAME:-multiagent-social}:${TAG:-latest}
else
	@echo "REGISTRY not set. Set REGISTRY to enable push."
endif

release: frontend-build docker-build docker-push
	@echo "release done"

