# 部署指南

> 本文档面向系统管理员，介绍如何在服务器上部署和维护租赁管理系统。

---

## 目录

- [环境要求](#环境要求)
- [快速部署（SQLite 模式）](#快速部署sqlite-模式)
- [Docker 部署](#docker-部署)
- [PostgreSQL 部署](#postgresql-部署)
- [环境变量说明](#环境变量说明)
- [数据备份与恢复](#数据备份与恢复)
- [HTTPS 配置](#https-配置)

---

## 环境要求

| 组件 | 最低版本 | 说明 |
|------|----------|------|
| Go | 1.21+ | 编译后端 |
| Node.js | 18+ | 编译前端 |
| Docker（可选） | 20.10+ | Docker 部署时需要 |
| SQLite | 内置 | 默认数据库，无需额外安装 |
| PostgreSQL（可选） | 14+ | 生产环境推荐 |

---

## 快速部署（SQLite 模式）

SQLite 模式是最简单的部署方式，无需额外数据库，适合小型团队或试用。

### 1. 获取代码

```bash
git clone <仓库地址>
cd he-tong-sheng-cheng
```

### 2. 编译

```bash
make build
```

此命令会依次：
- 编译前端（`cd frontend && npm ci && npm run build`）
- 编译后端 Go 二进制（将前端 dist 嵌入二进制文件）

编译完成后生成单个 `server` 文件。

### 3. 启动

```bash
JWT_SECRET=your-random-secret-string ./server
```

> `JWT_SECRET` 是**必填**环境变量，请设置为一个随机字符串（如 `openssl rand -hex 32` 生成）。

### 4. 访问

打开浏览器访问：`http://服务器IP:8080`

默认管理员账号：`admin` / `admin123`

### 开发模式

如果需要在开发时自动重载：

```bash
# 终端 1：前端热重载
make dev-frontend

# 终端 2：后端
make dev
```

---

## Docker 部署

使用 Docker Compose 一键部署，自动包含 PostgreSQL 数据库。

### 1. 配置环境变量

创建 `.env` 文件（项目根目录）：

```env
JWT_SECRET=your-random-secret-string
```

### 2. 启动

```bash
docker-compose up -d
```

该命令会启动两个容器：
- **app**：应用主程序，监听 8080 端口
- **postgres**：PostgreSQL 17 数据库

### 3. 访问

打开浏览器访问：`http://服务器IP:8080`

### 4. 查看日志

```bash
docker-compose logs -f app
```

### 5. 停止和重启

```bash
# 停止
docker-compose down

# 重启
docker-compose restart

# 重新编译并启动
docker-compose up -d --build
```

### 数据持久化

Docker Compose 使用名为 `pgdata` 的 volume 持久化 PostgreSQL 数据。即使容器被删除，数据也不会丢失。

如需完全清除数据（谨慎操作）：

```bash
docker-compose down -v
```

---

## PostgreSQL 部署

适用于正式生产环境，数据量较大或需要多用户并发访问时推荐使用。

### 1. 数据库准备

```sql
-- 以 postgres 用户登录
CREATE DATABASE asset_leasing;
```

### 2. 启动参数

**使用命令行参数：**

```bash
./server -mode postgres \
  -db-host localhost \
  -db-port 5432 \
  -db-user postgres \
  -db-pass your-db-password \
  -db-name asset_leasing
```

**使用环境变量（推荐）：**

```bash
export MODE=postgres
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASS=your-db-password
export DB_NAME=asset_leasing
export JWT_SECRET=your-random-secret-string
export PORT=8080

./server
```

---

## 环境变量说明

| 变量名 | 必填 | 默认值 | 说明 |
|--------|------|--------|------|
| `JWT_SECRET` | **是** | 无 | JWT 认证密钥，必须设置 |
| `MODE` | 否 | `sqlite` | 数据库模式：`sqlite` 或 `postgres` |
| `PORT` | 否 | `8080` | 服务监听端口 |
| `DB_HOST` | 否 | `localhost` | PostgreSQL 主机地址 |
| `DB_PORT` | 否 | `5432` | PostgreSQL 端口 |
| `DB_USER` | 否 | `postgres` | PostgreSQL 用户名 |
| `DB_PASS` | 否 | 空 | PostgreSQL 密码 |
| `DB_NAME` | 否 | `asset_leasing` | PostgreSQL 数据库名 |

> 命令行参数和环境变量同时设置时，**环境变量优先**。

---

## 数据备份与恢复

### SQLite 备份

**备份：**

```bash
# 停止服务后复制数据库文件
cp data.db data.db.bak.$(date +%Y%m%d)
```

或使用 SQLite 的在线备份命令（不需要停止服务）：

```bash
sqlite3 data.db ".backup 'data.db.bak.$(date +%Y%m%d)'"
```

**恢复：**

```bash
# 停止服务
cp data.db.bak.20240101 data.db
# 重启服务
```

### PostgreSQL 备份

**备份：**

```bash
# 使用 pg_dump
pg_dump -U postgres -h localhost asset_leasing > backup_$(date +%Y%m%d).sql
```

**恢复：**

```bash
# 创建空数据库（如需要）
createdb -U postgres asset_leasing_restore

# 恢复数据
psql -U postgres -h localhost asset_leasing_restore < backup_20240101.sql
```

**定时备份（crontab）：**

```bash
# 编辑 crontab
crontab -e

# 每天凌晨 2 点自动备份
0 2 * * * pg_dump -U postgres asset_leasing > /data/backup/db_$(date +\%Y\%m\%d).sql
```

---

## HTTPS 配置

建议使用 Nginx 作为反向代理，配置 SSL 证书实现 HTTPS 访问。

### Nginx 配置示例

```nginx
server {
    listen 80;
    server_name your-domain.com;
    return 301 https://$host$request_uri;
}

server {
    listen 443 ssl http2;
    server_name your-domain.com;

    ssl_certificate     /etc/nginx/ssl/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/key.pem;

    # 安全头
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### 使用 Let's Encrypt 免费证书

```bash
# 安装 certbot（Ubuntu/Debian）
sudo apt install certbot python3-certbot-nginx

# 获取证书
sudo certbot --nginx -d your-domain.com

# 自动续期（certbot 通常会自动配置）
sudo certbot renew --dry-run
```
