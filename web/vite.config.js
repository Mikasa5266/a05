import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import basicSsl from '@vitejs/plugin-basic-ssl'
import viteCompression from 'vite-plugin-compression'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

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
      AutoImport({
        resolvers: [
          ElementPlusResolver({
            importStyle: 'css'
          })
        ]
      }),
      Components({
        resolvers: [
          ElementPlusResolver({
            importStyle: 'css'
          })
        ]
      }),
      tailwindcss(),
      viteCompression({
        algorithm: 'gzip',
        ext: '.gz',
        threshold: 10240,
        deleteOriginFile: false
      }),
    ],
    build: {
      sourcemap: false,
      cssCodeSplit: true,
      rollupOptions: {
        output: {
          manualChunks(id) {
            if (!id.includes('node_modules')) return

            if (
              id.includes('/node_modules/vue/') ||
              id.includes('/node_modules/vue-router/') ||
              id.includes('/node_modules/pinia/')
            ) {
              return 'vendor-vue'
            }

            if (
              id.includes('/node_modules/element-plus/') ||
              id.includes('/node_modules/@element-plus/')
            ) {
              return 'vendor-element'
            }

            if (
              id.includes('/node_modules/chart.js/') ||
              id.includes('/node_modules/vue-chartjs/')
            ) {
              return 'vendor-charts'
            }
          }
        }
      }
    },
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
        },
        '/uploads': {
          target: proxyTarget,
          changeOrigin: true,
          secure: false
        }
      }
    }
  }
})