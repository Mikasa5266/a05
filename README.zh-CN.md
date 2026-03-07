# AI Interview Pro (a05)

语言切换: [English](./README.md) | **中文**

AI Interview Pro 是一个全栈的智能面试训练与人才协同平台。
项目覆盖学生求职训练、企业招聘管理、高校就业支持，以及 AI 驱动的知识与社区能力。

English version: [README.md](./README.md)

## 核心亮点

- 三端门户统一架构：`student`、`enterprise`、`university`
- 简历解析与岗位匹配（结构化结果）
- 模拟面试主流程 + AI 追问策略 + 风格模式
- 语音答题链路（录音 + Whisper 兼容 ASR 转写）
- 面试报告、成长分析、历史记录
- 企业模块：人才池、岗位管理、能力标准、招聘分析
- 高校模块：学生跟踪、帮扶动作、人才推送
- 社区模块：发帖、评论、点赞、导师预约

## 技术栈

- 后端：Go、Gin、GORM、MySQL、JWT
- 前端：Vue 3、Vite、Pinia、Vue Router、Element Plus
- AI/基础设施：LLM Provider 抽象、Whisper 兼容 ASR、OCR（Tesseract + pdftoppm）

## 仓库结构

```text
.
|- server/                 # Go API 服务
|  |- config.yaml          # 运行配置（DB/JWT/LLM/ASR/OCR）
|  |- main.go              # 服务入口
|  |- router/              # 路由注册
|  |- handler/             # HTTP 处理器
|  |- service/             # 业务逻辑与 AI 编排
|  |- repository/          # 数据访问层
|  |- model/               # 数据模型
|  |- pkg/                 # 三方集成（ASR/LLM/websocket）
|  |- utils/               # 文件解析与 OCR 工具
|  |- tools/               # 运维/清理脚本
|
|- web/                    # Vue 前端
|  |- src/
|  |  |- views/            # 三端页面
|  |  |- components/       # 公共组件
|  |  |- stores/           # Pinia 状态管理
|  |  |- api/              # API 封装
|  |  |- utils/request.js  # Axios 基础客户端
|
|- knowledge_base/         # 知识库与提示词内容
|- scoring_rubrics/        # 评分规则配置
```

## 环境要求

- Go `1.25+`（与 `server/go.mod` 对齐）
- Node.js `18+` 与 npm
- MySQL `8+`
- OCR 可选依赖（处理扫描版 PDF 时建议安装）：
  - `tesseract`
  - `pdftoppm`（Poppler）

## 快速开始

### 1) 启动后端

```bash
cd server
go mod tidy
```

按本机环境修改 `server/config.yaml`。

注意：
- 发布前请替换数据库/JWT/LLM/ASR 凭据。
- `main.go` 通过相对路径读取 `config.yaml`，请在 `server/` 目录执行启动命令。

```bash
cd server
go run main.go
```

默认 API 前缀：`http://localhost:8080/api/v1`

### 2) 启动前端

```bash
cd web
npm install
```

可选：创建 `web/.env.development`

```env
VITE_API_URL=http://localhost:8080/api/v1
```

```bash
cd web
npm run dev
```

### 3) 访问系统

- 打开终端中 Vite 输出地址（通常 `http://localhost:5173`）
- 进入门户选择页，按角色进入对应端

## 演示账号初始化

可一键创建企业端和高校端演示账号：

```bash
cd server
go run ./tools/create_accounts.go
```

默认创建：
- `enterprise@test.com` / `123456`
- `university@test.com` / `123456`

学生账号可在 UI 或注册接口创建。

## 常用命令

后端：

```bash
cd server
go test ./...
go run main.go
```

前端：

```bash
cd web
npm run dev
npm run build
npm run preview
```

清理低质量追问题（默认 dry-run）：

```bash
cd server
go run ./tools/cleanup_garbage_questions
# 执行删除请追加 -apply
```

## API 领域总览

所有接口都在 `/api/v1` 下。

- 认证与用户：`/register`、`/login`、`/user/*`
- 面试流程：`/interview/*`、盲盒场景、风格揭示、语音分析
- 简历：`/resume/parse`、`/resume/generate-questions`
- 报告与成长：`/reports/*`、`/growth/stats`
- AI 对话：`/ai/chat`、`/interview/:id/ai-chat`
- 企业端：`/enterprise/*`
- 高校端：`/university/*`
- 社区：`/community/*`

说明：除注册/登录外，大多数接口需要 JWT。

## API 调用示例

基础地址：

```text
http://localhost:8080/api/v1
```

### 1) 注册（默认学生）

```bash
curl -X POST "http://localhost:8080/api/v1/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "demo_student",
    "email": "demo_student@test.com",
    "password": "123456",
    "role": "student"
  }'
```

示例响应：

```json
{
  "message": "User registered successfully",
  "user": {
    "id": 1,
    "username": "demo_student",
    "email": "demo_student@test.com",
    "role": "student"
  }
}
```

### 2) 登录

```bash
curl -X POST "http://localhost:8080/api/v1/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "demo_student@test.com",
    "password": "123456",
    "role": "student"
  }'
```

示例响应：

```json
{
  "message": "Login successful",
  "token": "<jwt_token>",
  "user": {
    "id": 1,
    "username": "demo_student",
    "email": "demo_student@test.com",
    "role": "student"
  }
}
```

### 3) 开始面试（需要 JWT）

```bash
curl -X POST "http://localhost:8080/api/v1/interview/start" \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "technical",
    "mode": "ai",
    "difficulty": "medium",
    "position": "backend"
  }'
```

### 4) 解析简历（multipart，需要 JWT）

```bash
curl -X POST "http://localhost:8080/api/v1/resume/parse" \
  -H "Authorization: Bearer <jwt_token>" \
  -F "file=@./sample_resume.pdf"
```

示例响应：

```json
{
  "resume": {
    "name": "Candidate Name",
    "skills": ["Go", "MySQL", "Vue"]
  },
  "matches": [
    {
      "jobTitle": "Backend Engineer",
      "score": 0.89,
      "reason": "Strong backend stack match"
    }
  ]
}
```

### 5) 检查 OCR 状态（需要 JWT）

```bash
curl -X GET "http://localhost:8080/api/v1/system/ocr/status" \
  -H "Authorization: Bearer <jwt_token>"
```

## 部署指南

### A. 本机部署（单机）

1. 安装 MySQL、Go、Node.js。
2. 配置 `server/config.yaml`。
3. 在 `server/` 启动后端：`go run main.go`。
4. 构建前端：

```bash
cd web
npm install
npm run build
```

5. 使用 Nginx/Caddy 或其他静态服务托管 `web/dist`。
6. 将前端 API 地址指向后端公网地址（`VITE_API_URL`）。

### B. Docker 部署（推荐一致性）

后端镜像示例：

```dockerfile
# server/Dockerfile
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/. .
RUN go build -o app main.go

FROM alpine:3.21
WORKDIR /app
COPY --from=builder /app/app ./app
COPY server/config.yaml ./config.yaml
EXPOSE 8080
CMD ["./app"]
```

前端镜像示例：

```dockerfile
# web/Dockerfile
FROM node:22-alpine AS builder
WORKDIR /app
COPY web/package*.json ./
RUN npm install
COPY web/. .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

最小 `docker-compose.yml` 示例：

```yaml
services:
  mysql:
    image: mysql:8.4
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: interview_ai
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

  backend:
    build:
      context: .
      dockerfile: server/Dockerfile
    ports:
      - "8080:8080"
    depends_on:
      - mysql

  frontend:
    build:
      context: .
      dockerfile: web/Dockerfile
    ports:
      - "80:80"
    depends_on:
      - backend

volumes:
  mysql_data:
```

### C. 云部署（VM 或容器平台）

1. 先准备托管 MySQL，并使用独立账号权限。
2. 部署后端服务：
   - `config.yaml` 指向云数据库地址。
   - `server.host` 保持 `0.0.0.0`。
   - 配置反向代理与 HTTPS。
3. 部署前端静态资源（`web/dist`）到 CDN/对象存储。
4. 配置域名、跨域和回源策略。
5. 增加可观测性：
   - 访问日志
   - 错误告警
   - 健康检查（含 OCR 与登录链路）
6. 落实密钥轮换、最小权限、数据库备份与恢复演练。

生产检查清单：

- 替换 `server/config.yaml` 中所有本地/测试密钥。
- 按环境分离配置（dev/staging/prod）。
- 全链路启用 TLS。
- 限制数据库对公网暴露。
- 定期验证备份可恢复性。

## 常见问题排查

- 在仓库根目录启动后端失败：
  - 原因：`main.go` 按相对路径读取 `config.yaml`。
  - 解决：进入 `server/` 后执行 `go run main.go`。

- `8080` 端口被占用：
  - 修改 `server/config.yaml` 中 `server.port`，或结束占用进程。

- 扫描版 PDF 简历识别失败：
  - 检查 `tesseract_path`、`pdftoppm_path`、`tessdata_path`。
  - 确认语言包存在（如 `chi_sim`、`eng`）。

- 前端无法请求后端：
  - 检查 `VITE_API_URL` 与后端 CORS/网络配置。

## 安全说明

- 不要将真实密钥和密码提交到版本库。
- `server/config.yaml` 建议按环境拆分管理。
- 生产前务必轮换 JWT Secret 与全部第三方凭据。

## License

当前仓库未提供 LICENSE 文件。