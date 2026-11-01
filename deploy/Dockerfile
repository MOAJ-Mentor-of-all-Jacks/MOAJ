# MOAJ (Mentor of all Jacks) - Multi-Stage Single-Binary Container Build
# Implementation Target: Phase 5 (Dev Compose) & Phase 8 (Production-ready Community release)
# Reference: docs/adr/ADR-01-go-sqlc-single-binary.md & docs/IMPLEMENTATION_PLAN.md §7 (Phase 8)

FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/moaj ./cmd/moaj

FROM alpine:3.20
RUN adduser -D -u 10001 moaj
USER moaj
WORKDIR /app
COPY --from=builder /out/moaj /usr/local/bin/moaj
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/moaj"]
CMD ["serve"]
