import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

import { visualizer } from 'rollup-plugin-visualizer'
export default defineConfig({
  plugins: [visualizer({ open: false, filename: 'dist/bundle-stats.html' }),react()],
  server: {
    port: 3004,
    proxy: {
      '/api': { target: process.env.VITE_API_URL ?? 'http://developer-portal-backend', changeOrigin: true },
    },
  },
})
