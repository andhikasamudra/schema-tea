import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

const apiProxyTarget = process.env.API_PROXY_TARGET || 'http://localhost:8080'

const server = {
  port: 5173,
  proxy: {
    '/api': apiProxyTarget,
  },
}

if (process.env.CHOKIDAR_USEPOLLING === 'true') {
  server.watch = { usePolling: true }
}

export default defineConfig({
  plugins: [vue()],
  server,
})
