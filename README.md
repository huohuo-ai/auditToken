# AI Gateway - 企业AI网关管理系统

一个功能完善的企业级AI网关系统，提供MaaS（Model as a Service）能力，支持多模型管理、用户配额控制、审计日志和风险行为检测。

## 功能特性

### 1. MaaS AI网关能力
- 支持多种大模型提供商（OpenAI、Azure、Anthropic等）
- 统一的API接口，兼容OpenAI格式
- 模型配置管理，支持动态添加/修改/删除模型
- 员工API Key管理，一员工一Key
- 流式响应支持

### 2. Token使用限制
- 灵活的配额管理：按天、按周、按月设置限额
- 实时配额检查，防止超额使用
- 配额使用统计和可视化
- 管理员可针对每个用户设置不同的配额

### 3. 审计功能
- 所有请求和响应数据存入ClickHouse
- 支持TB级日志存储和快速查询
- 完整的请求追踪（Request ID）
- 多维度日志检索（时间、用户、模型、IP等）

### 4. 高危行为检测
智能风险检测系统，自动识别以下风险行为：
- **Token滥用**：单次或累计请求token数异常
- **非工作时间访问**：配置工作时间外的访问检测
- **敏感信息获取**：检测尝试获取密码、密钥等敏感信息的请求
- **异常请求频率**：短时间内大量请求
- **IP异常**：可疑IP地址或地理位置异常
- **提示词注入**：检测DAN模式、开发者模式等注入攻击
- **异常请求模式**：检测其他异常行为模式

### 5. 其他功能
- **用户管理**：支持多用户，区分管理员和普通用户
- **速率限制**：基于Redis的滑动窗口限流
- **缓存机制**：用户、模型信息多级缓存
- **监控告警**：风险事件实时告警（可对接钉钉/企业微信）
- **数据可视化**：仪表盘展示使用趋势和统计

## 技术架构

### 后端技术栈
- **Go 1.21+**：高性能后端服务
- **Gin**：Web框架
- **GORM**：ORM框架，支持MySQL/PostgreSQL
- **JWT**：用户认证
- **Redis**：缓存和速率限制
- **ClickHouse**：审计日志存储和分析

### 前端技术栈
- **Vue 3**：前端框架
- **Element Plus**：UI组件库
- **ECharts**：数据可视化
- **Axios**：HTTP客户端

## 快速开始

### 环境要求
- Go 1.21+
- Node.js 16+
- MySQL 8.0+ 或 PostgreSQL 13+
- Redis 6+
- ClickHouse

### 后端部署

1. 复制配置文件
```bash
cd backend
cp config.example.yaml config.yaml
# 编辑 config.yaml 配置数据库等信息
```

2. 安装依赖
```bash
go mod download
```

3. 运行服务
```bash
go run cmd/server/main.go
```

### 前端部署

1. 安装依赖
```bash
cd frontend
npm install
```

2. 开发模式运行
```bash
npm run serve
```

3. 生产构建
```bash
npm run build
```

### Docker部署

```bash
cd backend/deployments
docker-compose up -d
```

## 默认账号

- 用户名：`admin`
- 密码：`admin123`

登录后请立即修改默认密码。

## API使用示例

### 获取模型列表
```bash
curl http://localhost:8080/v1/models \
  -H "Authorization: Bearer YOUR_API_KEY"
```

### 对话请求
```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'
```

### 流式对话
```bash
curl http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [{"role": "user", "content": "Hello!"}],
    "stream": true
  }'
```

## 项目结构

```
ai-gateway/
├── backend/
│   ├── cmd/server/          # 入口文件
│   ├── internal/
│   │   ├── config/          # 配置管理
│   │   ├── handler/         # HTTP处理器
│   │   ├── middleware/      # 中间件
│   │   ├── model/           # 数据模型
│   │   ├── repository/      # 数据访问层
│   │   ├── service/         # 业务逻辑层
│   │   └── audit/           # 审计和风险检测
│   ├── pkg/llm/             # LLM客户端
│   ├── deployments/         # Docker部署配置
│   └── config.example.yaml  # 配置示例
├── frontend/
│   ├── src/
│   │   ├── api/             # API接口
│   │   ├── views/           # 页面视图
│   │   ├── router/          # 路由配置
│   │   ├── store/           # Vuex状态管理
│   │   └── utils/           # 工具函数
│   └── package.json
└── README.md
```

## 配置说明

### 后端配置 (config.yaml)

```yaml
server:
  port: "8080"
  mode: "release"  # debug/release

database:
  driver: "mysql"
  host: "localhost"
  port: 3306
  username: "root"
  password: "password"
  database: "ai_gateway"

redis:
  host: "localhost"
  port: 6379

clickhouse:
  host: "localhost"
  port: 8123
  database: "ai_gateway"

jwt:
  secret: "your-secret-key"
  expires_in: 86400

audit:
  off_hours_start: 22      # 非工作时间开始
  off_hours_end: 6         # 非工作时间结束
  token_threshold_hourly: 100000
  suspicious_ip_list:      # 可疑IP列表
    - "192.168.1.100"
```

## 风险检测配置

风险检测规则可通过管理界面或数据库配置：

### 敏感信息检测
内置检测：密码、密钥、身份证号、银行卡号、薪资信息等

### 自定义检测模式
可在数据库`prompt_patterns`表中添加自定义正则规则：

```sql
INSERT INTO prompt_patterns (pattern, pattern_type, risk_level, description) 
VALUES ('keyword', 'sensitive_info', 'high', '检测特定关键词');
```

## 性能优化

- **连接池**：数据库和Redis连接池配置
- **缓存策略**：多级缓存减少数据库压力
- **异步处理**：审计日志异步写入
- **ClickHouse**：列式存储优化日志查询性能

## 安全建议

1. **修改默认密码**：首次登录后修改admin密码
2. **使用HTTPS**：生产环境配置SSL证书
3. **定期轮换API Key**：建议定期重新生成API Key
4. **IP白名单**：可配置允许的IP范围
5. **敏感信息加密**：数据库中的API Key应加密存储

## 常见问题

### 1. ClickHouse连接失败
- 检查ClickHouse服务是否启动
- 确认端口8123是否开放
- 检查配置中的用户名密码是否正确

### 2. API Key认证失败
- 确认API Key是否正确
- 检查用户状态是否为active
- 查看请求头格式是否为 `Authorization: Bearer YOUR_API_KEY`

### 3. 配额超限
- 联系管理员调整配额限制
- 或等待配额重置（日/周/月）

## 许可证

MIT License

## 贡献

欢迎提交Issue和Pull Request。

## 联系方式

如有问题或建议，请通过以下方式联系：
- 邮箱：your-email@company.com
- 项目地址：https://github.com/yourcompany/ai-gateway
