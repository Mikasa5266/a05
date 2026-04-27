<div align="center">
  <img src="./docs/assets/logo-placeholder.png" width="120" alt="AI Interview Pro Logo" />

  # AI Interview Pro

  ### 🤖 面向真实求职场景的 AI 面试与能力提升平台

  <p>
    <a href="./README.zh-CN.md">旧版中文说明</a> ·
    <a href="./README-zh-CN.md">历史文档</a>
  </p>
</div>

<div align="center">

![Vue](https://img.shields.io/badge/Vue-3-42b883?logo=vuedotjs&logoColor=white)
![Electron](https://img.shields.io/badge/Electron-Desktop-47848f?logo=electron&logoColor=white)
![Go](https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)
![Vite](https://img.shields.io/badge/Vite-7.x-646CFF?logo=vite&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-yellow.svg)

</div>

## ✨ 一句话简介

**AI Interview Pro** 是一个支持 **Web + Electron 桌面端** 的全栈智能面试系统，集成 AI 追问、语音转写、面试报告、成长分析与多角色协作（学生 / 企业 / 高校），并可通过 Docker 在服务器快速部署。

## 🚀 核心特性

- 🧩 **多端一致体验**：同一套业务能力同时支持 Web 与桌面端。
- 🎙️ **真实对话式面试**：支持语音作答、实时追问、评分与改进建议。
- 🧠 **多模型能力编排**：LLM/ASR/TTS 可配置，便于扩展与替换。
- 🧑‍💼 **多角色系统**：学生训练、企业招聘、高校就业支持一体化。
- 📊 **闭环成长反馈**：报告、历史记录、能力趋势持续追踪。
- 🛠️ **工程化落地**：支持 Docker Compose 一键部署与 Electron 安装包分发。

## 🔐 安全与合规能力（已落地）

- **实名注册核验（手机号 + 身份证）**：学生注册接口要求提交 `real_name`、`phone`、`id_card_no`，后端执行手机号与身份证校验位校验，仅保存身份证哈希与脱敏值。
- **访问日志 / 操作日志留存**：后端中间件自动记录 `access_log` 与 `operation_log` 到安全审计表，默认保留不少于 180 天（可配置，最小值受控为 180）。
- **防 SQL 注入 / 防 XSS**：启用请求安全拦截（SQL 注入/XSS 特征检测）、参数化查询（GORM 占位符）与用户内容 HTML 转义存储。
- **违法信息过滤 + 巡查 + 处置**：发布内容命中违规词将被拦截；用户可提交举报；管理员可审核处置；系统定时巡查社区内容并自动生成待处理举报单。
- **安全责任人 / 应急联系人**：通过公开接口 `GET /api/v1/security/contact` 提供责任人与应急联系人信息。

推荐配置项（`server/config.yaml`）：

- `SECURITY_LOG_RETENTION_DAYS`：日志留存天数（最小 180）
- `SECURITY_PATROL_INTERVAL_MINUTES`：系统巡查周期
- `SECURITY_PATROL_SCAN_LIMIT`：单次巡查扫描上限
- `SECURITY_RESPONSIBLE_NAME / PHONE / EMAIL`：安全责任人
- `SECURITY_EMERGENCY_NAME / PHONE / EMAIL`：应急联系人

## 🖼️ 演示截图
<img width="958" height="598" alt="Snipaste_2026-04-07_13-40-03" src="https://github.com/user-attachments/assets/534cfae6-e3b1-49a0-a7cb-e0f925a6cff4" />
<img width="958" height="598" alt="Snipaste_2026-04-07_13-40-23" src="https://github.com/user-attachments/assets/112a737c-fc5d-459f-a006-e1fc9da1b8ac" />
<img width="1280" height="764" alt="Snipaste_2026-04-07_13-40-57" src="https://github.com/user-attachments/assets/dc50be45-7b29-48c7-a4f0-e20789c66675" />




## 🏗️ 技术栈与架构

### Tech Stack

| 层级 | 方案 |
| --- | --- |
| 前端 Web | Vue 3 + Vite + Pinia + Vue Router + Element Plus |
| 桌面端 | Electron + vite-plugin-electron + electron-builder |
| 后端 | Go + Gin + GORM + JWT |
| 数据层 | MySQL 8 |
| AI 能力 | LLM 抽象层 + Whisper 兼容 ASR + TTS + OCR |
| 部署 | Docker Compose + Nginx + 可选 Coturn |

### 架构图

```mermaid
flowchart LR
  U[用户] --> W[Web 前端
Vue3 + Vite]
  U --> D[Desktop 客户端
Electron]

  W --> N[Nginx / HTTPS]
  D --> N

  N --> B[Go API
Gin + JWT]
  B --> M[(MySQL)]
  B --> A[LLM / ASR / TTS]
  B --> K[Knowledge Base]
```

## ⚡ 快速开始（傻瓜式）

### 0) 环境准备

- Node.js 18+
- Go 1.25+
- Docker + Docker Compose

### 1) 重置演示数据环境（推荐先执行）

```bash
cd server
go run ./tools/create_accounts
```

说明：该工具会清理并重建演示数据（学生/企业/高校账号、岗位、题库、社区样例内容）。

### 2) 运行 Web 端（开发）

```bash
cd web
npm install
npm run dev:web
```

默认会启动 Vite 开发服务，打开终端提示地址即可访问。

### 3) 运行 Electron 客户端（开发）

```bash
cd web
npm install
npm run dev:desktop
```

### 4) 打包 Electron 客户端（Windows .exe）

```bash
cd web
npm run dist:win
```

产物目录：`web/release/`

### 5) Docker Compose 一键启动服务端（推荐部署方式）

```bash
cp .env.example .env
docker compose up -d --build
```

Windows PowerShell：

```powershell
Copy-Item .env.example .env
docker compose up -d --build
```

说明：`docker-compose.yml` 默认开启 `INIT_DEMO_ACCOUNTS=true`，后端启动时会幂等补齐演示登录账号（不会重复创建同邮箱账号）。

常用排查命令：

```bash
docker compose ps
docker compose logs -f backend
docker compose logs -f frontend
```

## 👤 演示账号

统一密码：`123456`

### 学生端

| 邮箱 | 显示名称 |
| --- | --- |
| `student01@test.com` | 林书远 |
| `student02@test.com` | 顾清禾 |
| `student03@test.com` | 唐知行 |
| `student04@test.com` | 魏景初 |
| `student05@test.com` | 许明川 |

### 企业端

| 邮箱 | 显示名称 |
| --- | --- |
| `enterprise01@test.com` | 星澜智联科技有限公司 |
| `enterprise02@test.com` | 远川数据智能有限公司 |
| `enterprise03@test.com` | 海岳工业软件有限公司 |

### 高校端

| 邮箱 | 显示名称 |
| --- | --- |
| `university01@test.com` | 江南数字工程大学就业中心 |
| `university02@test.com` | 华南应用科技大学就业中心 |

## 📦 客户端下载

请前往本仓库 **GitHub Releases** 下载最新桌面安装包（`.exe`）：

👉 `https://github.com/<YOUR_ORG>/<YOUR_REPO>/releases`

> 建议发布时同时上传：安装包、`latest.yml`、`.blockmap`，以支持自动更新。

## 📁 项目结构（精简）

```text
.
├─ server/                # Go 后端服务
├─ web/                   # Vue + Electron 前端
├─ knowledge_base/        # 知识库与提示词
├─ docker-compose.yml     # 一键部署编排
└─ README.md
```

## 🤝 团队与致谢

- 团队：`我们叫什么队`
- 感谢：所有贡献者、测试同学与开源社区。

---

如果你觉得这个项目对你有帮助，欢迎点个 ⭐ Star 支持一下！
