import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { Agent } from 'node:http'

export default defineConfig({
  plugins: [vue()],
  server: {
    host: '0.0.0.0',
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        // Reuse API connections instead of forwarding Connection: close to browsers.
        agent: new Agent({ keepAlive: true }),
        changeOrigin: true,
        xfwd: true,
        ws: true,
      },
    },
  },
})
