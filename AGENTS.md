# AI Gateway - Agent开发指南

## 项目概述

企业级AI网关管理系统，提供MaaS（Model as a Service）能力，支持多模型管理、用户配额控制、审计日志和风险行为检测。

## 技术栈

### 后端
- Go 1.21+
- Gin (Web框架)
- GORM (ORM)
- JWT (认证)
- Redis (缓存/限流)
- ClickHouse (审计日志)
- MySQL/PostgreSQL (主数据库)

### 前端
- Vue 3
- Element Plus (UI)
- ECharts (图表)
- Vuex (状态管理)

## 项目结构

```
ai-gateway/
├── backend/
│   ├── cmd/server/          # 入口
│   ├── internal/
│   │   ├── config/          # 配置
│   │   ├── handler/         # HTTP处理器
│   │   ├── middleware/      # 中间件
│   │   ├── model/           # 数据模型
│   │   ├── repository/      # 数据访问
│   │   ├── service/         # 业务逻辑
│   │   └── audit/           # 审计/风险检测
│   ├── pkg/llm/             # LLM客户端
│   └── deployments/         # Docker配置
└── frontend/                # Vue前端
```

## 核心功能模块

### 1. 认证与授权
- 文件: `internal/middleware/auth.go`
- JWT Token认证
- API Key认证
- 角色权限控制

### 2. 配额管理
- 文件: `internal/middleware/quota.go`
- 按天/周/月限制Token使用
- Redis缓存配额状态

### 3. 审计日志
- 文件: `internal/middleware/audit.go`
- ClickHouse存储所有请求响应
- 异步写入减少延迟

### 4. 风险检测
- 文件: `internal/audit/detector.go`
- Token滥用检测
- 非工作时间访问
- 敏感信息获取
- 异常请求频率
- IP异常
- 提示词注入检测

### 5. LLM代理
- 文件: `pkg/llm/client.go`
- 支持多提供商
- 流式响应支持

## 开发注意事项

### 添加新的风险检测规则
1. 在 `internal/audit/detector.go` 中添加检测方法
2. 在 `DetectRisk` 中调用新方法
3. 如有必要，更新 `model.RiskType` 枚举

### 添加新的LLM提供商
1. 在 `internal/model/aimodel.go` 添加提供商常量
2. 在 `pkg/llm/client.go` 实现请求构建逻辑

### 数据库变更
1. 修改 `internal/model/` 中的模型
2. 在 `internal/repository/database.go` 的 `AutoMigrate` 中添加新表

## 配置项

重要配置在 `backend/config.example.yaml`：

- `server`: 服务端口和模式
- `database`: 主数据库配置
- `redis`: 缓存配置
- `clickhouse`: 审计日志数据库
- `jwt`: 认证密钥
- `audit`: 风险检测阈值

## 默认账号

- 用户名: admin
- 密码: admin123

## API调用示例

```bash
# 对话
POST /v1/chat/completions
Header: Authorization: Bearer {API_KEY}
Body: {
  "model": "gpt-3.5-turbo",
  "messages": [{"role": "user", "content": "Hello"}]
}

# 获取模型列表
GET /v1/models
Header: Authorization: Bearer {API_KEY}
```

## 部署

使用Docker Compose:
```bash
cd backend/deployments
docker-compose up -d
```

依赖服务：
- MySQL (端口3306)
- Redis (端口6379)
- ClickHouse (端口8123)

## 测试

后端启动后访问:
- API: http://localhost:8080
- 前端: http://localhost:3000

## 扩展建议

1. 添加更多LLM提供商支持（Claude、Gemini等）
2. 实现告警通知（钉钉/企业微信/邮件）
3. 添加模型性能监控和自动故障转移
4. 实现请求/响应内容过滤
5. 添加多租户支持
