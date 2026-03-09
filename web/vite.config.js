import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import basicSsl from '@vitejs/plugin-basic-ssl'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const proxyTarget = env.VITE_PROXY_TARGET || 'http://127.0.0.1:8080'
  const useHttps = env.VITE_DEV_HTTPS !== 'false'

  return {
    plugins: [
      basicSsl(),
      vue({
        template: {
          compilerOptions: {
            isCustomElement: (tag) => tag === 'model-viewer'
          }
        }
      }),
      tailwindcss(),
    ],
    server: {
      https: useHttps,
      host: '0.0.0.0', // 强制监听所有网络接口
      allowedHosts: true,
      cors: true, // 关闭跨域拦截
      hmr: true, // 让 Vite 自动推断 HMR 连接参数
      strictPort: false,
      open: false,
      proxy: {
        '/api/v1': {
          target: proxyTarget,
          changeOrigin: true,
          ws: true,
          secure: false
        }
      }
    }
  }
})