import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
  ],
  server: {
    host: '0.0.0.0', // 强制监听所有网络接口
    allowedHosts: [
      // 1. 具体的 ngrok 域名（精准匹配）
      'unpaid-lincoln-drenchingly.ngrok-free.dev',
      // 2. 通配符匹配所有 ngrok-free.dev 域名（避免下次换域名再改）
      '.ngrok-free.dev',
      // 3. 兜底允许所有本地/局域网地址
      'localhost',
      '127.0.0.1'
    ],
    cors: true, // 关闭跨域拦截
    hmr: true, // 让 Vite 自动推断 HMR 连接参数
    strictPort: false,
    open: false
  }
})