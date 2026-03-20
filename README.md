# observatory-user-svc

User management microservice for the [Observatory](https://github.com/luciocarvalhojr/observatory) platform. Provides user CRUD via REST, persists to PostgreSQL, and publishes `user.created` / `user.deleted` events to NATS.

## Endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/users` | Create user |
| `GET` | `/users/:id` | Get user by ID |
| `PUT` | `/users/:id` | Update user name |
| `DELETE` | `/users/:id` | Delete user |
| `GET` | `/healthz` | Liveness probe |
| `GET` | `/readyz` | Readiness probe (DB ping) |

## Local Development

The easiest way to get all dependencies (PostgreSQL and NATS) up and running is using Docker Compose.

```bash
docker compose up --build
```

Alternatively, you can run the service manually:

1. **Start dependencies:** `docker compose up postgres nats -d`
2. **Run:** `go run cmd/api/main.go`

## Local Validation (Docker Compose)

The `docker-compose.yml` provides a complete environment including **PostgreSQL 17** and **NATS 2** with JetStream enabled. The database schema is applied automatically via `dev/init.sql` on first start.

### 1. Start the Environment

```bash
docker compose up --build
```

### 2. Create a User

```bash
curl -X POST http://localhost:8082/users \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","name":"Alice"}'
```

### 3. Fetch a User

```bash
curl http://localhost:8082/users/<id>
```

### 4. Verify Health

```bash
curl http://localhost:8082/healthz
curl http://localhost:8082/readyz
```

## Environment Variables

| Variable | Description | Default |
|---|---|---|
| `PORT` | Listen port | `8082` |
| `DATABASE_URL` | PostgreSQL connection URL | required |
| `NATS_URL` | NATS connection URL | required |
| `OTLP_ENDPOINT` | Jaeger OTLP endpoint | `http://jaeger:4318` |
| `ENV` | `development` or `production` | `production` |

## Pre-commit Hooks (lefthook)

The project uses [lefthook](https://github.com/evilmartians/lefthook) to run the same checks as the CI pipeline locally on every `git commit`.

### Install

```bash
brew install lefthook
lefthook install
```

### What runs on each commit

| Check | Tool | Mirrors pipeline step |
|---|---|---|
| Lint | `golangci-lint run` | `test-and-lint` / Lint |
| Tests + coverage gate | `go test -race -coverprofile` | `test-and-lint` / Test & Coverage |
| Swagger docs up to date | `swag init` + git diff | `test-and-lint` / Verify Swagger docs |
| SAST | `gosec ./...` | `security-scan` / Gosec |
| SCA | `govulncheck ./...` | `security-scan` / Govulncheck |
| Secrets scan | `gitleaks protect --staged --redact` | `security-scan` / Gitleaks |
| No direct commits to `main` | branch name check | branch protection |
| Merge conflict markers | grep on staged files | code quality |
| Trailing whitespace | grep on staged files | code quality |
| Missing newline at EOF | tail check on staged files | code quality |
| Large files (>512 KB) | file size check | repo hygiene |
| YAML syntax | `yamllint -d relaxed` | code quality |

All checks run in parallel. If any fail, the commit is blocked.

### Required tools

```bash
brew install golangci-lint gitleaks yamllint
go install github.com/swaggo/swag/cmd/swag@latest
go install github.com/securego/gosec/v2/cmd/gosec@v2.23.0
go install golang.org/x/vuln/cmd/govulncheck@latest
```

### Skip hooks (emergency only)

```bash
git commit --no-verify
```

## Generate Swagger Docs

```bash
swag init -g cmd/api/main.go
```

## Run Tests

```bash
go test -v -race -cover ./...
```

## Build Validation (Docker)

```bash
docker build -t user-svc .
```

## CI/CD Pipeline

The GitHub Actions pipeline (`.github/workflows/devsecops.yml`) runs on every push and pull request to `main`.

| Job | What it does |
|---|---|
| `test-and-lint` | Lint (`golangci-lint`), tests with race detector, coverage gate, Swagger doc verification |
| `security-scan` | Secret scanning (Gitleaks), SAST (Gosec), SCA (Govulncheck) |
| `docker-scan` | Builds the Docker image and runs a Trivy scan (blocks on `CRITICAL`/`HIGH` CVEs) |
| `release` | Creates a GitHub release via semantic-release (conventional commits) |
| `build-and-push` | Builds and pushes the image to GHCR, generates an SPDX SBOM, and signs the image with Cosign (keyless via GitHub OIDC) |
| `update-helm-chart` | Opens a PR on [`luciocarvalhojr/helm-charts`](https://github.com/luciocarvalhojr/helm-charts) to bump the image tag/digest — merging that PR triggers ArgoCD to deploy |

The `release`, `build-and-push`, and `update-helm-chart` jobs only run on pushes to `main` when a new version is published.
