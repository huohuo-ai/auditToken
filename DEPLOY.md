# AI Gateway Docker Compose 部署指南

本文档介绍如何使用 Docker Compose 部署 AI Gateway 系统。

## 📋 环境要求

- Docker 20.10+
- Docker Compose 2.0+
- Git

## 🚀 快速开始

### 1. 克隆仓库

```bash
git clone https://github.com/huohuo-ai/auditToken.git
cd auditToken
```

### 2. 初始化配置

```bash
# Linux/Mac
./deploy.sh setup

# Windows PowerShell
.\deploy.ps1 setup
```

### 3. 访问服务

- **前端界面**: http://localhost
- **后端 API**: http://localhost:8080
- **健康检查**: http://localhost:8080/health

默认账号：
- 用户名: `admin`
- 密码: `admin123`

> ⚠️ **安全提醒**: 生产环境请立即修改默认密码！

## 📁 目录结构

```
ai-gateway/
├── docker-compose.yml          # Docker Compose 主配置
├── .env.example                # 环境变量模板
├── deploy.sh                   # Linux/Mac 部署脚本
├── deploy.ps1                  # Windows 部署脚本
├── backend/
│   ├── config.docker.yaml      # Docker 环境后端配置
│   └── deployments/
│       └── Dockerfile          # 后端镜像构建
├── frontend/
│   ├── Dockerfile              # 前端镜像构建
│   └── nginx.conf              # Nginx 配置
├── nginx/
│   └── nginx.conf              # 反向代理配置
└── init/                       # 数据库初始化脚本
    ├── mysql/
    └── clickhouse/
```

## ⚙️ 配置说明

### 环境变量 (.env)

复制 `.env.example` 为 `.env` 并修改配置：

```bash
cp .env.example .env
```

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| MYSQL_ROOT_PASSWORD | password | MySQL root 密码 |
| MYSQL_DATABASE | ai_gateway | 数据库名 |
| MYSQL_USER | gateway | 数据库用户名 |
| MYSQL_PASSWORD | gateway123 | 数据库密码 |
| BACKEND_PORT | 8080 | 后端服务端口 |
| FRONTEND_PORT | 80 | 前端服务端口 |

### 后端配置 (backend/config.yaml)

复制 Docker 配置模板：

```bash
cp backend/config.docker.yaml backend/config.yaml
```

**重要**: 修改以下配置项：
- `jwt.secret`: JWT 密钥（生产环境必须修改）
- 数据库连接信息（如果使用非 Docker 部署的数据库）

## 🛠️ 常用命令

### Linux/Mac

```bash
# 初始化并启动
./deploy.sh setup

# 启动服务
./deploy.sh start

# 停止服务
./deploy.sh stop

# 重启服务
./deploy.sh restart

# 重新构建镜像
./deploy.sh build

# 更新服务
./deploy.sh update

# 查看日志
./deploy.sh logs
./deploy.sh logs backend  # 只看后端日志

# 查看状态
./deploy.sh status

# 清理数据（危险！）
./deploy.sh cleanup
```

### Windows PowerShell

```powershell
# 初始化并启动
.\deploy.ps1 setup

# 启动服务
.\deploy.ps1 start

# 停止服务
.\deploy.ps1 stop

# 其他命令与 Linux 相同
```

## 🐳 Docker Compose 命令

```bash
# 启动所有服务
docker compose up -d

# 停止所有服务
docker compose down

# 停止并删除数据卷
docker compose down -v

# 查看日志
docker compose logs -f

# 查看特定服务日志
docker compose logs -f backend

# 重启服务
docker compose restart

# 重新构建并启动
docker compose up -d --build

# 使用 Nginx 反向代理启动
docker compose --profile nginx up -d
```

## 📦 服务说明

| 服务名 | 说明 | 端口 | 依赖 |
|--------|------|------|------|
| mysql | MySQL 数据库 | 3306 | - |
| redis | Redis 缓存 | 6379 | - |
| clickhouse | ClickHouse 审计数据库 | 8123, 9000 | - |
| backend | Go 后端服务 | 8080 | mysql, redis, clickhouse |
| frontend | Vue 前端服务 | 80 | backend |
| nginx | Nginx 反向代理 | 8000 | frontend, backend |

## 🔒 生产环境部署建议

### 1. 修改默认配置

```bash
# 1. 修改环境变量
vi .env

# 2. 修改后端配置
vi backend/config.yaml
```

### 2. 使用 HTTPS

配置 SSL 证书：

```yaml
# docker-compose.yml 添加证书挂载
services:
  nginx:
    volumes:
      - ./ssl/cert.pem:/etc/nginx/ssl/cert.pem
      - ./ssl/key.pem:/etc/nginx/ssl/key.pem
```

### 3. 启用 Nginx 反向代理

```bash
docker compose --profile nginx up -d
```

### 4. 配置防火墙

```bash
# 只开放必要端口
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
# 不要直接暴露 3306, 6379, 8123 到公网
```

### 5. 定期备份

```bash
# MySQL 备份
docker exec ai-gateway-mysql mysqldump -u root -p ai_gateway > backup.sql

# ClickHouse 备份
docker exec ai-gateway-clickhouse clickhouse-backup create
```

## 🐛 故障排查

### 查看服务状态

```bash
docker compose ps
docker compose logs
```

### 数据库连接失败

检查 backend/config.yaml 中的数据库配置：
- 主机名应为服务名（mysql, redis, clickhouse）
- 确保端口正确

### 端口被占用

修改 `.env` 文件中的端口映射：

```env
BACKEND_PORT=8081
FRONTEND_PORT=8082
```

### 内存不足

调整 Docker 容器资源限制：

```yaml
services:
  backend:
    deploy:
      resources:
        limits:
          memory: 1G
```

## 📞 获取帮助

- 查看日志: `docker compose logs -f`
- 提交 Issue: https://github.com/huohuo-ai/auditToken/issues
