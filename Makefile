.PHONY: build build-frontend build-backend run clean test dev-frontend dev

# ---- 构建 ----
build: build-frontend build-backend

build-frontend:
	@echo "==> Building frontend..."
	cd frontend && npm ci && npm run build

build-backend:
	@echo "==> Building Go binary..."
	CGO_ENABLED=1 go build -ldflags="-s -w" -o server ./cmd/server

# ---- 运行 ----
run: build
	JWT_SECRET=$${JWT_SECRET:?JWT_SECRET is required} ./server

dev-frontend:
	cd frontend && npm run dev

dev: build-frontend
	JWT_SECRET=dev-secret-key go run ./cmd/server

# ---- 测试 ----
test:
	go test ./... -v -count=1

test-cover:
	go test ./... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# ---- 清理 ----
clean:
	rm -f server
	rm -rf cmd/server/dist
	rm -f coverage.out coverage.html
