import { defineConfig, loadEnv } from "vite";
import vue from "@vitejs/plugin-vue";
import tailwindcss from "@tailwindcss/vite";
import basicSsl from "@vitejs/plugin-basic-ssl";
import electron from "vite-plugin-electron/simple";
import viteCompression from "vite-plugin-compression";
import AutoImport from "unplugin-auto-import/vite";
import Components from "unplugin-vue-components/vite";
import { ElementPlusResolver } from "unplugin-vue-components/resolvers";

export default defineConfig(({ mode, command }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const buildTarget = String(
    env.VITE_BUILD_TARGET || process.env.BUILD_TARGET || mode || "web",
  ).toLowerCase();
  const isDesktopTarget =
    buildTarget === "desktop" || buildTarget === "electron";
  const proxyTarget = env.VITE_PROXY_TARGET || "http://127.0.0.1:8082";
  const useHttps = env.VITE_DEV_HTTPS !== "false";
  const isServeCommand = command === "serve";

  const plugins = [
    basicSsl(),
    vue({
      template: {
        compilerOptions: {
          isCustomElement: (tag) => tag === "model-viewer",
        },
      },
    }),
    AutoImport({
      resolvers: [
        ElementPlusResolver({
          importStyle: "css",
        }),
      ],
    }),
    Components({
      resolvers: [
        ElementPlusResolver({
          importStyle: "css",
        }),
      ],
    }),
    tailwindcss(),
    viteCompression({
      algorithm: "gzip",
      ext: ".gz",
      threshold: 10240,
      deleteOriginFile: false,
    }),
  ];

  if (isDesktopTarget) {
    plugins.push(
      electron({
        main: {
          entry: "electron/main.ts",
        },
        preload: {
          input: "electron/preload.ts",
        },
      }),
    );
  }

  return {
    plugins,
    optimizeDeps: {
      // ngrok is sensitive to many tiny module requests; aggressively prebundle heavy deps.
      force: isServeCommand,
      include: [
        "vue",
        "vue-router",
        "pinia",
        "axios",
        "dayjs",
        "element-plus",
        "@element-plus/icons-vue",
        "lucide-vue-next",
        "echarts",
        "echarts/core",
        "echarts/charts",
        "echarts/components",
        "echarts/renderers",
        "zrender",
        "chart.js",
        "vue-chartjs",
        "monaco-editor",
      ],
    },
    build: {
      sourcemap: false,
      cssCodeSplit: true,
      rollupOptions: {
        output: {
          manualChunks(id) {
            if (!id.includes("node_modules")) return undefined;

            const normalizedId = id.replace(/\\/g, "/");

            if (
              normalizedId.includes("/node_modules/vue/") ||
              normalizedId.includes("/node_modules/@vue/") ||
              normalizedId.includes("/node_modules/vue-router/") ||
              normalizedId.includes("/node_modules/pinia/")
            ) {
              return "vendor-vue-core";
            }

            if (
              normalizedId.includes("/node_modules/monaco-editor/") ||
              normalizedId.includes("/node_modules/monaco-")
            ) {
              return "vendor-monaco";
            }

            if (
              normalizedId.includes("/node_modules/echarts/") ||
              normalizedId.includes("/node_modules/zrender/")
            ) {
              return "vendor-echarts";
            }

            if (
              normalizedId.includes("/node_modules/chart.js/") ||
              normalizedId.includes("/node_modules/vue-chartjs/")
            ) {
              return "vendor-chartjs";
            }

            if (
              normalizedId.includes("/node_modules/element-plus/") ||
              normalizedId.includes("/node_modules/@element-plus/")
            ) {
              return "vendor-element";
            }

            if (
              normalizedId.includes("/node_modules/lucide-vue-next/") ||
              normalizedId.includes("/node_modules/@element-plus/icons-vue/")
            ) {
              return "vendor-icons";
            }

            return "vendor-misc";
          },
        },
      },
    },
    server: {
      https: useHttps,
      host: "0.0.0.0", // 强制监听所有网络接口
      allowedHosts: true,
      cors: true, // 关闭跨域拦截
      headers: {
        // Public tunnel debugging is sensitive to stale module cache; force fresh fetches.
        "Cache-Control": "no-store",
      },
      hmr: true, // 让 Vite 自动推断 HMR 连接参数
      strictPort: false,
      open: false,
      proxy: {
        "/api/v1": {
          target: proxyTarget,
          changeOrigin: true,
          ws: true,
          secure: false,
        },
        "/uploads": {
          target: proxyTarget,
          changeOrigin: true,
          secure: false,
        },
      },
    },
  };
});
