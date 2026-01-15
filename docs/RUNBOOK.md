# Runbook - multiagent-social (MVP)

This runbook lists the common operational tasks for the MVP.

Prerequisites:
- Docker & Docker Compose
- Go 1.24.11
- psql CLI (optional)

Start services locally:

```bash
cd deployments/docker
docker compose up --build -d
```

Apply DB migrations:

```bash
export PG_DSN="postgres://postgres:password@localhost:5432/multiagent?sslmode=disable"
psql "$PG_DSN" -f ../../migrations/001_init.sql
```

Run server locally (for development):

```bash
cd project/root
go run ./cmd/server
```

Admin UI:
- Visit `http://localhost:8080/admin/` to create/list agents and inspect conversations via WebSocket.

Basic troubleshooting:
- If database connection fails, check container logs: `docker compose logs db`
- If Redis not reachable, check `docker compose logs redis`
- To rebuild the app container: `docker compose build app`

Backup / restore:
- Use `pg_dump` for backups and `psql` for restore.

Security & production notes:
- Add TLS, authentication (JWT/mTLS) and proper secret management before production.
- Add monitoring and log aggregation (Prometheus / Grafana / ELK).

