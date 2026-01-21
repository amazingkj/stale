# Stage 1: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app/ui
COPY ui/package*.json ./
RUN npm ci
COPY ui/ ./
RUN npm run build

# Stage 2: Build backend
FROM golang:1.23-alpine AS backend-builder
ENV GOTOOLCHAIN=auto
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/ui/dist ./ui/dist
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /stale ./cmd/stale

# Stage 3: Final image
FROM alpine:3.19

# Install dependencies (ca-certificates for HTTPS, tzdata for timezones, curl for healthcheck)
RUN apk --no-cache add ca-certificates tzdata curl

# Create non-root user for security
RUN addgroup -g 1000 stale && adduser -D -u 1000 -G stale stale

WORKDIR /app
COPY --from=backend-builder /stale .

# Create data directory and set ownership
RUN mkdir -p /data && chown -R stale:stale /data /app

ENV STALE_DB_PATH=/data/stale.db
ENV STALE_PORT=3000

# Switch to non-root user
USER stale

EXPOSE 3000
VOLUME ["/data"]

# Health check - verify the API is responding
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:3000/api/v1/health || exit 1

ENTRYPOINT ["/app/stale"]
