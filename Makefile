.PHONY: build build-frontend build-backend run clean test test-cover dev-frontend dev lint typecheck docker-build docker-up docker-down

# ---- 构建 ----
build: build-frontend build-backend

build-frontend:
	@echo "==> 构建前端..."
	cd frontend && npm ci && npm run build

build-backend:
	@echo "==> 构建后端二进制..."
	CGO_ENABLED=1 go build -ldflags="-s -w" -o server ./cmd/server

# ---- 运行 ----
run: build
	JWT_SECRET=$${JWT_SECRET:?JWT_SECRET is required} ADMIN_PASSWORD=$${ADMIN_PASSWORD:?ADMIN_PASSWORD is required} ./server

dev-frontend:
	cd frontend && npm run dev

dev: build-frontend
	JWT_SECRET=dev-secret-key ADMIN_PASSWORD=admin123 go run ./cmd/server

# ---- 测试 ----
test:
	go test ./... -v -count=1

test-cover:
	go test ./... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

# ---- 代码质量 ----
lint:
	@echo "==> 运行 go vet..."
	go vet ./...
	@echo "==> 运行前端 lint..."
	cd frontend && npx vue-tsc --noEmit

typecheck:
	@echo "==> Go 类型检查..."
	go build ./...
	@echo "==> TypeScript 类型检查..."
	cd frontend && npx vue-tsc --noEmit

# ---- Docker ----
docker-build:
	docker build -t asset-leasing-system .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

# ---- 数据库 ----
migrate:
	@echo "数据库迁移由 GORM AutoMigrate 在启动时自动执行"

# ---- 清理 ----
clean:
	rm -f server
	rm -rf cmd/server/dist
	rm -f coverage.out coverage.html
