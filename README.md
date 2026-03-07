# AI Interview Pro (a05)

Language: **English** | [中文](./README.zh-CN.md)

AI Interview Pro is a full-stack interview training and talent collaboration platform.
It includes student interview practice, enterprise recruiting workflows, university employment support, and an AI-powered knowledge/community layer.

Chinese version: [README.zh-CN.md](./README.zh-CN.md)

## Highlights

- Multi-portal system: `student`, `enterprise`, `university`
- Resume parsing with structured extraction and job matching
- Mock interview flow with AI follow-up strategy and style controls
- Voice answer pipeline (recording + ASR transcription via Whisper-compatible API)
- Interview reports, growth analytics, and historical records
- Enterprise modules: talent pool, jobs, standards, analytics
- University modules: student tracking, support actions, talent push
- Community modules with posts, comments, likes, mentoring bookings

## Tech Stack

- Backend: Go, Gin, GORM, MySQL, JWT
- Frontend: Vue 3, Vite, Pinia, Vue Router, Element Plus
- AI/Infra: LLM provider abstraction, Whisper-compatible ASR, OCR (Tesseract + pdftoppm)

## Repository Layout

```text
.
|- server/                 # Go API service
|  |- config.yaml          # Runtime config (DB/JWT/LLM/ASR/OCR)
|  |- main.go              # Service entry
|  |- router/              # Route registration
|  |- handler/             # HTTP handlers
|  |- service/             # Business logic and AI orchestration
|  |- repository/          # Data access layer
|  |- model/               # DB models and domain structs
|  |- pkg/                 # Integrations (ASR, LLM, websocket)
|  |- utils/               # File parsing and OCR helpers
|  |- tools/               # Utility scripts
|
|- web/                    # Vue frontend
|  |- src/
|  |  |- views/            # Pages for all portals
|  |  |- components/       # Shared UI components
|  |  |- stores/           # Pinia stores
|  |  |- api/              # API wrappers
|  |  |- utils/request.js  # Axios base client
|
|- knowledge_base/         # Domain prompts/content used by the platform
|- scoring_rubrics/        # Scoring configs
```

## Prerequisites

- Go `1.25+` (matches `server/go.mod`)
- Node.js `18+` and npm
- MySQL `8+`
- Optional for OCR fallback:
	- `tesseract` executable
	- `pdftoppm` (Poppler)

## Quick Start

### 1) Backend setup

```bash
cd server
go mod tidy
```

Edit `server/config.yaml` for your local environment.

Important:
- Replace DB/JWT/LLM/ASR credentials before sharing or deploying.
- `main.go` loads `config.yaml` from current working directory, so run backend commands inside `server/`.

Start backend:

```bash
cd server
go run main.go
```

Default API base: `http://localhost:8080/api/v1`

### 2) Frontend setup

```bash
cd web
npm install
```

Optional env file (`web/.env.development`):

```env
VITE_API_URL=http://localhost:8080/api/v1
```

Start frontend:

```bash
cd web
npm run dev
```

### 3) Open the app

- Visit the Vite URL shown in terminal (usually `http://localhost:5173`)
- Use portal selection page to enter:
	- Student portal
	- Enterprise portal
	- University portal

## Account Bootstrapping

You can create demo enterprise/university accounts with:

```bash
cd server
go run ./tools/create_accounts.go
```

The script creates:
- `enterprise@test.com` / `123456`
- `university@test.com` / `123456`

Student accounts can be created via register API/UI.

## Common Commands

Backend:

```bash
cd server
go test ./...
go run main.go
```

Frontend:

```bash
cd web
npm run dev
npm run build
npm run preview
```

Cleanup low-quality follow-up questions:

```bash
cd server
go run ./tools/cleanup_garbage_questions
# dry-run by default; add -apply to execute deletion
```

## Core API Areas

All APIs are under ` /api/v1 `.

- Auth/User: `/register`, `/login`, `/user/*`
- Interview: `/interview/*`, blindbox, style reveal, speech analyze
- Resume: `/resume/parse`, `/resume/generate-questions`
- Reports/Growth: `/reports/*`, `/growth/stats`
- AI Chat: `/ai/chat`, `/interview/:id/ai-chat`
- Enterprise: `/enterprise/*`
- University: `/university/*`
- Community: `/community/*`

Note: most routes require JWT auth.

## API Examples

Base URL:

```text
http://localhost:8080/api/v1
```

### 1) Register (student by default)

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

Example response:

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

### 2) Login

```bash
curl -X POST "http://localhost:8080/api/v1/login" \
	-H "Content-Type: application/json" \
	-d '{
		"email": "demo_student@test.com",
		"password": "123456",
		"role": "student"
	}'
```

Example response:

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

### 3) Start interview (JWT required)

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

### 4) Parse resume (multipart upload, JWT required)

```bash
curl -X POST "http://localhost:8080/api/v1/resume/parse" \
	-H "Authorization: Bearer <jwt_token>" \
	-F "file=@./sample_resume.pdf"
```

Example response:

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

### 5) OCR status check (JWT required)

```bash
curl -X GET "http://localhost:8080/api/v1/system/ocr/status" \
	-H "Authorization: Bearer <jwt_token>"
```

## Deployment

### Option A: Local deployment (single machine)

1. Install MySQL, Go, Node.js.
2. Configure `server/config.yaml`.
3. Run backend from `server/`: `go run main.go`.
4. Build frontend:

```bash
cd web
npm install
npm run build
```

5. Serve `web/dist` using Nginx/Caddy or any static server.
6. Set frontend API URL to backend public address via `VITE_API_URL`.

### Option B: Docker deployment (recommended for consistency)

Create backend Docker image:

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

Create frontend Docker image:

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

Minimal compose example:

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

### Option C: Cloud deployment (VM or container platform)

1. Provision managed MySQL and set secure credentials.
2. Deploy backend service:
	 - Ensure `config.yaml` uses cloud DB host/port.
	 - Set `server.host` to `0.0.0.0`.
	 - Configure reverse proxy and HTTPS.
3. Deploy frontend static assets (`web/dist`) to CDN/object storage.
4. Configure domain and CORS policy.
5. Set up observability:
	 - access logs
	 - error alerts
	 - health checks (`/api/v1/system/ocr/status` and auth smoke tests)
6. Rotate secrets and enforce least-privilege DB users.

Production checklist:

- Replace all local keys in `server/config.yaml`.
- Use environment-specific config management.
- Enable TLS for public traffic.
- Restrict database network exposure.
- Back up database and verify restore flow.

## Troubleshooting

- Backend fails when run from repository root:
	- Cause: `config.yaml` not found by relative path.
	- Fix: run `go run main.go` inside `server/`.

- Port `8080` already in use:
	- Change `server.port` in `server/config.yaml`, or stop occupying process.

- Resume OCR does not work for scanned PDFs:
	- Verify `tesseract_path`, `pdftoppm_path`, and `tessdata_path` in `server/config.yaml`.
	- Ensure language packs exist (for example `chi_sim`, `eng`).

- Frontend cannot reach backend:
	- Verify `VITE_API_URL` and backend CORS/network settings.

## Security Notes

- Do not commit real API keys/passwords into version control.
- Keep `server/config.yaml` environment-specific (local/staging/prod).
- Rotate JWT secret and all provider credentials before production.

## License

No license file is currently defined in this repository.