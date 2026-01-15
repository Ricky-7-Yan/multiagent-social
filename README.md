# multiagent-social (MVP)

This repository is an MVP implementation of a multi-Agent social ecosystem server written in Go (Go 1.24.11).

Quick start (Windows PowerShell):

1. Ensure Go 1.24.11 is installed:

```powershell
go version
# expected: go version go1.24.11 windows/amd64
```

2. Start dependencies with Docker Compose:

```powershell
cd deployments/docker
docker compose up --build
```

3. Run migrations (example using psql):

```powershell
$env:PG_DSN = "postgres://postgres:password@localhost:5432/multiagent?sslmode=disable"
psql $env:PG_DSN -f ../..\\migrations\\001_init.sql
```

4. Run the server:

```powershell
cd ../..
go run ./cmd/server
```

Health check: `http://localhost:8080/health`

Endpoints (MVP):
- `GET /api/v1/agents` - list agents (stub)
- `POST /api/v1/agents` - create agent (stub)
- `POST /api/v1/conversations` - create conversation (returns id)
- `POST /api/v1/conversations/{id}/messages` - post a user message (body raw text)
 - `POST /api/v1/conversations/{id}/debate` - start structured debate among participants
 - `GET /api/v1/conversations` - list conversations (id + title)
 - `GET /metrics` - Prometheus metrics endpoint

This README contains minimal instructions for local development. See `Makefile` and `deployments/docker/docker-compose.yml`.

Embedding & PGVector:
- Set `OPENAI_API_KEY` in environment to enable OpenAI embeddings.
- Ensure Postgres has `pgvector` extension: the migration uses `vector(1536)` column. If your Postgres image doesn't include `pgvector`, install the extension or use a Postgres image with pgvector (e.g., `ankane/pgvector`).
- After starting DB, run migrations as before. The embeddings will be generated via OpenAI and stored in `embeddings.vector`.

Front-end (React + TypeScript):
- Requirements: Node 18+ and npm.
- Development:
  - cd web/react-app
  - npm install
  - npm run dev
- Build (produces static files served at `/admin` by the Go server):
  - make frontend-build
  - or: cd web/react-app && npm ci && npm run build

CI / Release:
- The repository contains GitHub Actions workflows:
  - `.github/workflows/build-and-push.yml` — runs on pushes to `main`: builds frontend, runs `go test ./...`, runs `golangci-lint`, then builds & pushes Docker image (requires `DOCKER_*` secrets).
  - `.github/workflows/release.yml` — runs on tag pushes (vX.Y.Z): builds frontend and uses `docker/build-push-action` to push tagged image and create a GitHub Release.
- To perform a local release build:
  - make frontend-build
  - make docker-build IMAGE_NAME=multiagent-social TAG=latest
  - make docker-push REGISTRY=your.registry.com IMAGE_NAME=multiagent-social TAG=latest
  - or use `scripts/release.sh` (set REGISTRY/IMAGE_NAME/TAG env vars).

Kubernetes / Helm:
- Chart is in `deployments/k8s/`. Configure `values.yaml` (image.repository, tag, ingress, serviceMonitor, env) and install with:
  - helm install my-release deployments/k8s -f deployments/k8s/values.yaml
- There is a `migration` Job template to apply DB migrations; enable and run it as part of deployment (it expects `PG_DSN` in env).

Local validation helper:
- `scripts/local-validate.sh` will spin up Docker Compose, run migrations, and execute unit tests.


