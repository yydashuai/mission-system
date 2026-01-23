import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig({
  plugins: [vue()],
  server: {
    host: '0.0.0.0',
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://192.168.40.11:8001',
        changeOrigin: true,
      },
      '/apis': {
        target: 'http://192.168.40.11:8001',
        changeOrigin: true,
      },
      '/healthz': {
        target: 'http://192.168.40.11:8001',
        changeOrigin: true,
      },
    },
  },
})
