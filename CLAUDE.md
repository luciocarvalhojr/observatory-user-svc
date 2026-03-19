# observatory-user-svc

Go-based user management microservice for the Observatory platform. Provides user CRUD via REST, persists to PostgreSQL, and publishes `user.created` / `user.deleted` events to NATS.

---

## Current State

- **Version**: unreleased (new service)
- **Port**: 8082
- **Pipeline**: `.github/workflows/devsecops.yml` — mirrors auth-svc pipeline: lint → security-scan → docker-scan → release → build-and-push → update-helm-chart
- **Coverage gate**: Set to 0% (not enforced — no tests written yet)
- **Deployed**: No — not yet added to helm-charts or k8s-home-lab

---

## In Progress

- Initial scaffold complete. Needs: GitHub repo creation, go mod tidy, first commit.

---

## Known Issues

- **No tests**: Coverage gate is set to `THRESHOLD=0`. No unit or integration tests exist.
- **No startup validation beyond DATABASE_URL**: NATS_URL is not validated at startup — service will fail at runtime if missing.
- **No migration tooling**: Schema is applied via `dev/init.sql` on first PostgreSQL start. No migration runner (e.g. goose) for production schema changes.

---

## Next Steps

- Create GitHub repo `observatory-user-svc`
- Run `go mod tidy` to resolve dependencies
- Run `swag init -g cmd/api/main.go` to generate Swagger docs
- `lefthook install` to activate pre-commit hooks
- Write unit tests and raise coverage gate to 80%
- Add a migration runner (goose) for schema versioning
- Add Helm chart in `luciocarvalhojr/helm-charts`
- Add ArgoCD Application in `k8s-home-lab`

---

## Key Decisions

- **PostgreSQL via pgx/v5**: Direct SQL, no ORM. pgxpool for connection pooling.
- **NATS for events**: Publishes `user.created` and `user.deleted`; consumers are notify-svc and incident-svc.
- **Non-fatal publish errors**: If NATS publish fails after a successful DB write, the error is logged but not returned to the caller. The write is the source of truth.
- **No Redis**: Unlike auth-svc, user-svc has no session state — PostgreSQL only.
- **Same DevSecOps pipeline as auth-svc**: Identical CI/CD structure for consistency across services.

---

## API Endpoints

| Method   | Path        | Description              | Auth |
|----------|-------------|--------------------------|------|
| `POST`   | `/users`    | Create user              | No (called by api-gateway after auth) |
| `GET`    | `/users/:id`| Get user by ID           | No |
| `PUT`    | `/users/:id`| Update user name         | No |
| `DELETE` | `/users/:id`| Delete user              | No |
| `GET`    | `/healthz`  | Liveness probe           | No |
| `GET`    | `/readyz`   | Readiness probe (PG ping)| No |

## Environment Variables

| Variable       | Description              | Default    |
|----------------|--------------------------|------------|
| `PORT`         | HTTP listen port         | `8082`     |
| `DATABASE_URL` | PostgreSQL connection URL | required  |
| `NATS_URL`     | NATS connection URL      | required   |
| `OTLP_ENDPOINT`| OpenTelemetry collector  | `http://jaeger:4318` |
| `ENV`          | `development`/`production` | `production` |

## Local Dev

```bash
docker compose up        # starts postgres + nats + user-svc
curl localhost:8082/healthz
curl -X POST localhost:8082/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","name":"Alice"}'
```
