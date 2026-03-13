# ── Dev stage (used by docker compose) ────────────────────
FROM golang:1.24-alpine AS dev
RUN go install github.com/air-verse/air@v1.64.4
WORKDIR /app
COPY api/go.mod api/go.sum ./api/
RUN cd api && go mod download
COPY api/ ./api/
COPY .air.toml .
CMD ["air", "-c", ".air.toml"]

# ── Build Go API ──────────────────────────────────────────
FROM golang:1.24-alpine AS api-build
WORKDIR /app
COPY api/go.mod api/go.sum ./api/
RUN cd api && go mod download
COPY api/ ./api/
RUN cd api && CGO_ENABLED=0 go build -o /bin/server ./cmd/server

# ── Build SPA ─────────────────────────────────────────────
FROM node:20-alpine AS spa-build
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm ci
COPY frontend/ .
RUN npm run build

# ── Production runtime ────────────────────────────────────
FROM alpine:3.19 AS production
RUN apk add --no-cache ca-certificates tzdata
COPY --from=api-build /bin/server /bin/server
COPY --from=spa-build /app/dist /app/dist
WORKDIR /app
EXPOSE 8080
CMD ["/bin/server"]
