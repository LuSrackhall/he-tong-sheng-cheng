# Stage 1: Build frontend
FROM node:22-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json ./frontend/
RUN cd frontend && npm ci
COPY frontend/ ./frontend/
RUN cd frontend && npm run build

# Stage 2: Build backend
FROM golang:1.26.1-alpine AS backend-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend-builder /app/frontend/dist ./cmd/server/dist
RUN CGO_ENABLED=1 go build -ldflags="-s -w" -o server ./cmd/server

# Stage 3: Production
FROM alpine:3.21
RUN apk add --no-cache ca-certificates && adduser -D -u 1001 appuser
WORKDIR /app
COPY --from=backend-builder --chown=appuser:appuser /app/server .
RUN mkdir -p uploads backups data && chown -R appuser:appuser uploads backups data
USER appuser
EXPOSE 8080
CMD ["sh", "-c", "./server"]
