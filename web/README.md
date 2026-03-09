# Vue 3 + Vite

This template should help get you started developing with Vue 3 in Vite. The template uses Vue 3 `<script setup>` SFCs, check out the [script setup docs](https://v3.vuejs.org/api/sfc-script-setup.html#sfc-script-setup) to learn more.

Learn more about IDE Support for Vue in the [Vue Docs Scaling up Guide](https://vuejs.org/guide/scaling-up/tooling.html#ide-support).

## Cross-Network Access With ngrok

Use this when interviewers are not in the same LAN.

### 1) Start backend server

Run backend on local machine first (default `http://127.0.0.1:8080`).

### 2) Configure frontend env

`web/.env` already supports reverse proxy mode:

```env
VITE_API_URL=/api/v1
VITE_PROXY_TARGET=http://127.0.0.1:8080
VITE_DEV_HTTPS=true
```

### 3) Configure ngrok

Install ngrok (Windows):

```powershell
winget install --id Ngrok.Ngrok --exact --accept-source-agreements --accept-package-agreements
```

1. Copy `web/ngrok.yml.example` to `web/ngrok.yml`.
2. Replace `authtoken` with your real token from ngrok dashboard.

Or run helper script:

```powershell
./scripts/setup-ngrok.ps1 -AuthToken "YOUR_NGROK_AUTHTOKEN"
```

### 4) Start services

In `web/` directory:

```powershell
npm run dev:public
```

In another terminal:

```powershell
ngrok start --config ./ngrok.yml interview-web
```

You can also use:

```powershell
./scripts/start-public.ps1
```

### 5) Share URL

Share ngrok HTTPS URL (for example `https://xxxx.ngrok-free.app`) to interviewer.

Keep notes:

- Always use the `https://` ngrok URL, not local IP.
- Browser camera/mic permission must be granted on both sides.
- If backend is not local, set `VITE_PROXY_TARGET` to the actual backend origin.
