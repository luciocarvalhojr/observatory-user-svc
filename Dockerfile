# ── Build stage ──────────────────────────────────────────────────────
FROM golang:1.26.2-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /user-svc ./cmd/api

# ── Final stage — distroless ─────────────────────────────────────────
FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=builder /user-svc /user-svc

EXPOSE 8082

USER nonroot:nonroot

ENTRYPOINT ["/user-svc"]
