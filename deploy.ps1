# ============================================
# AI Gateway Docker Compose 部署脚本 (Windows PowerShell)
# ============================================

param(
    [Parameter(Position=0)]
    [string]$Command = "setup",
    
    [switch]$Pull
)

# 颜色函数
function Write-Info($message) {
    Write-Host "[INFO] $message" -ForegroundColor Cyan
}

function Write-Success($message) {
    Write-Host "[SUCCESS] $message" -ForegroundColor Green
}

function Write-Warn($message) {
    Write-Host "[WARN] $message" -ForegroundColor Yellow
}

function Write-Error($message) {
    Write-Host "[ERROR] $message" -ForegroundColor Red
}

# 检查 Docker
function Check-Docker {
    Write-Info "检查 Docker 环境..."
    
    try {
        $dockerVersion = docker --version
        Write-Success "Docker 已安装: $dockerVersion"
    } catch {
        Write-Error "Docker 未安装，请先安装 Docker"
        exit 1
    }
    
    try {
        $composeVersion = docker compose version
        Write-Success "Docker Compose 已安装: $composeVersion"
    } catch {
        Write-Error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    }
}

# 检查环境变量
function Check-Env {
    Write-Info "检查环境变量配置..."
    
    if (-not (Test-Path .env)) {
        Write-Warn ".env 文件不存在，将使用 .env.example 创建"
        Copy-Item .env.example .env
        Write-Warn "请编辑 .env 文件修改默认配置，然后重新运行此脚本"
        exit 1
    }
    
    Write-Success "环境变量配置检查通过"
}

# 检查配置文件
function Check-Config {
    Write-Info "检查配置文件..."
    
    if (-not (Test-Path backend/config.yaml)) {
        Write-Warn "backend/config.yaml 不存在，将使用 Docker 配置模板"
        Copy-Item backend/config.docker.yaml backend/config.yaml
        Write-Warn "请编辑 backend/config.yaml 修改配置（特别是 JWT Secret）"
    }
    
    Write-Success "配置文件检查通过"
}

# 创建目录
function Create-Directories {
    Write-Info "创建必要的目录..."
    
    New-Item -ItemType Directory -Force -Path data\mysql | Out-Null
    New-Item -ItemType Directory -Force -Path data\redis | Out-Null
    New-Item -ItemType Directory -Force -Path data\clickhouse | Out-Null
    New-Item -ItemType Directory -Force -Path logs\nginx | Out-Null
    
    Write-Success "目录创建完成"
}

# 拉取代码
function Pull-Code {
    if ($Pull) {
        Write-Info "拉取最新代码..."
        git pull origin main
        Write-Success "代码更新完成"
    }
}

# 构建镜像
function Build-Images {
    Write-Info "构建 Docker 镜像..."
    docker compose build --no-cache
    Write-Success "镜像构建完成"
}

# 启动服务
function Start-Services {
    Write-Info "启动服务..."
    docker compose up -d
    Write-Success "服务启动完成"
}

# 等待服务就绪
function Wait-ForServices {
    Write-Info "等待服务就绪..."
    
    Write-Host "等待 MySQL 就绪..."
    Start-Sleep -Seconds 5
    
    Write-Host "等待 Redis 就绪..."
    Start-Sleep -Seconds 2
    
    Write-Host "等待 ClickHouse 就绪..."
    Start-Sleep -Seconds 5
    
    Write-Host "等待后端服务就绪..."
    Start-Sleep -Seconds 5
    
    Write-Success "所有服务已就绪"
}

# 显示状态
function Show-Status {
    Write-Info "服务状态："
    docker compose ps
    
    Write-Host ""
    Write-Info "访问地址："
    Write-Host "  - 前端界面: http://localhost"
    Write-Host "  - 后端 API: http://localhost:8080"
    Write-Host "  - 健康检查: http://localhost:8080/health"
    Write-Host ""
    Write-Info "默认账号："
    Write-Host "  - 用户名: admin"
    Write-Host "  - 密码: admin123"
    Write-Host ""
    Write-Warn "生产环境请立即修改默认密码！"
}

# 停止服务
function Stop-Services {
    Write-Info "停止服务..."
    docker compose down
    Write-Success "服务已停止"
}

# 重启服务
function Restart-Services {
    Write-Info "重启服务..."
    docker compose restart
    Write-Success "服务已重启"
}

# 查看日志
function Show-Logs($service) {
    if ($service) {
        docker compose logs -f $service
    } else {
        docker compose logs -f
    }
}

# 更新服务
function Update-Services {
    Write-Info "更新服务..."
    docker compose pull
    docker compose up -d --build
    Write-Success "服务更新完成"
}

# 清理数据
function Cleanup {
    Write-Warn "此操作将删除所有数据，包括数据库！"
    $confirm = Read-Host "是否确认？[y/N]"
    
    if ($confirm -eq "y" -or $confirm -eq "Y") {
        docker compose down -v
        Remove-Item -Recurse -Force data/
        Write-Success "数据已清理"
    } else {
        Write-Info "操作已取消"
    }
}

# 显示帮助
function Show-Help {
    Write-Host "AI Gateway Docker Compose 部署脚本 (PowerShell)"
    Write-Host ""
    Write-Host "用法: .\deploy.ps1 [命令] [-Pull]"
    Write-Host ""
    Write-Host "命令:"
    Write-Host "  setup       初始化并启动所有服务"
    Write-Host "  start       启动服务"
    Write-Host "  stop        停止服务"
    Write-Host "  restart     重启服务"
    Write-Host "  build       重新构建镜像"
    Write-Host "  update      更新服务"
    Write-Host "  logs        查看所有服务日志"
    Write-Host "  status      查看服务状态"
    Write-Host "  cleanup     清理所有数据（危险！）"
    Write-Host "  help        显示帮助"
    Write-Host ""
    Write-Host "选项:"
    Write-Host "  -Pull       启动前拉取最新代码"
}

# 主逻辑
switch ($Command) {
    "setup" {
        Check-Docker
        Check-Env
        Check-Config
        Create-Directories
        Pull-Code
        Build-Images
        Start-Services
        Wait-ForServices
        Show-Status
    }
    "start" {
        Check-Docker
        Start-Services
        Show-Status
    }
    "stop" {
        Stop-Services
    }
    "restart" {
        Restart-Services
        Show-Status
    }
    "build" {
        Build-Images
    }
    "update" {
        Update-Services
    }
    "logs" {
        Show-Logs $args[0]
    }
    "status" {
        docker compose ps
    }
    "cleanup" {
        Cleanup
    }
    "help" {
        Show-Help
    }
    default {
        Write-Error "未知命令: $Command"
        Show-Help
        exit 1
    }
}
