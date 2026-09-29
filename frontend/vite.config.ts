import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
    // Proxy API calls to the Go backend during development.
    // This means the frontend can call /dr/... directly without hitting
    // CORS issues (the browser sees same-origin requests).
    proxy: {
      '/dr': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/mock': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
      '/health': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
})
