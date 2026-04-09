#!/bin/bash

# ============================================
# AI Gateway Docker Compose 部署脚本
# ============================================

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 打印带颜色的消息
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查 Docker 和 Docker Compose
check_docker() {
    log_info "检查 Docker 环境..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装 Docker"
        exit 1
    fi
    
    if ! command -v docker compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装 Docker Compose"
        exit 1
    fi
    
    log_success "Docker 环境检查通过"
}

# 检查环境变量文件
check_env() {
    log_info "检查环境变量配置..."
    
    if [ ! -f .env ]; then
        log_warn ".env 文件不存在，将使用 .env.example 创建"
        cp .env.example .env
        log_warn "请编辑 .env 文件修改默认配置，然后重新运行此脚本"
        exit 1
    fi
    
    log_success "环境变量配置检查通过"
}

# 检查配置文件
check_config() {
    log_info "检查配置文件..."
    
    if [ ! -f backend/config.yaml ]; then
        log_warn "backend/config.yaml 不存在，将使用 Docker 配置模板"
        cp backend/config.docker.yaml backend/config.yaml
        log_warn "请编辑 backend/config.yaml 修改配置（特别是 JWT Secret）"
    fi
    
    log_success "配置文件检查通过"
}

# 创建必要的目录
create_dirs() {
    log_info "创建必要的目录..."
    
    mkdir -p data/mysql data/redis data/clickhouse
    mkdir -p logs/nginx
    
    log_success "目录创建完成"
}

# 拉取最新代码（可选）
pull_code() {
    if [ "$1" = "--pull" ]; then
        log_info "拉取最新代码..."
        git pull origin main
        log_success "代码更新完成"
    fi
}

# 构建镜像
build_images() {
    log_info "构建 Docker 镜像..."
    
    docker compose build --no-cache
    
    log_success "镜像构建完成"
}

# 启动服务
start_services() {
    log_info "启动服务..."
    
    docker compose up -d
    
    log_success "服务启动完成"
}

# 等待服务就绪
wait_for_services() {
    log_info "等待服务就绪..."
    
    echo "等待 MySQL 就绪..."
    sleep 5
    
    echo "等待 Redis 就绪..."
    sleep 2
    
    echo "等待 ClickHouse 就绪..."
    sleep 5
    
    echo "等待后端服务就绪..."
    sleep 5
    
    log_success "所有服务已就绪"
}

# 显示服务状态
show_status() {
    log_info "服务状态："
    docker compose ps
    
    echo ""
    log_info "访问地址："
    echo "  - 前端界面: http://localhost"
    echo "  - 后端 API: http://localhost:8080"
    echo "  - 健康检查: http://localhost:8080/health"
    echo ""
    log_info "默认账号："
    echo "  - 用户名: admin"
    echo "  - 密码: admin123"
    echo ""
    log_warn "生产环境请立即修改默认密码！"
}

# 停止服务
stop_services() {
    log_info "停止服务..."
    docker compose down
    log_success "服务已停止"
}

# 重启服务
restart_services() {
    log_info "重启服务..."
    docker compose restart
    log_success "服务已重启"
}

# 查看日志
show_logs() {
    if [ -z "$1" ]; then
        docker compose logs -f
    else
        docker compose logs -f "$1"
    fi
}

# 更新服务
update_services() {
    log_info "更新服务..."
    
    docker compose pull
    docker compose up -d --build
    
    log_success "服务更新完成"
}

# 清理数据
cleanup() {
    log_warn "此操作将删除所有数据，包括数据库！"
    read -p "是否确认？[y/N] " confirm
    
    if [ "$confirm" = "y" ] || [ "$confirm" = "Y" ]; then
        docker compose down -v
        rm -rf data/
        log_success "数据已清理"
    else
        log_info "操作已取消"
    fi
}

# 显示帮助
show_help() {
    echo "AI Gateway Docker Compose 部署脚本"
    echo ""
    echo "用法: ./deploy.sh [命令]"
    echo ""
    echo "命令:"
    echo "  setup       初始化并启动所有服务"
    echo "  start       启动服务"
    echo "  stop        停止服务"
    echo "  restart     重启服务"
    echo "  build       重新构建镜像"
    echo "  update      更新服务"
    echo "  logs        查看所有服务日志"
    echo "  logs <svc>  查看指定服务日志"
    echo "  status      查看服务状态"
    echo "  cleanup     清理所有数据（危险！）"
    echo "  help        显示帮助"
    echo ""
    echo "选项:"
    echo "  --pull      启动前拉取最新代码"
}

# 主函数
main() {
    case "${1:-setup}" in
        setup)
            check_docker
            check_env
            check_config
            create_dirs
            pull_code "$2"
            build_images
            start_services
            wait_for_services
            show_status
            ;;
        start)
            check_docker
            start_services
            show_status
            ;;
        stop)
            stop_services
            ;;
        restart)
            restart_services
            show_status
            ;;
        build)
            build_images
            ;;
        update)
            update_services
            ;;
        logs)
            show_logs "$2"
            ;;
        status)
            docker compose ps
            ;;
        cleanup)
            cleanup
            ;;
        help)
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
