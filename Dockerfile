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
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=backend-builder /stale .

# Create data directory for SQLite
RUN mkdir -p /data
ENV STALE_DB_PATH=/data/stale.db
ENV STALE_PORT=3000

EXPOSE 3000
VOLUME ["/data"]

ENTRYPOINT ["/app/stale"]
