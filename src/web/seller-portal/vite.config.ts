import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

import { visualizer } from 'rollup-plugin-visualizer'
export default defineConfig({
  plugins: [visualizer({ open: false, filename: 'dist/bundle-stats.html' }),vue()],
  server: {
    port: 3002,
    proxy: {
      '/api': { target: process.env.VITE_API_URL ?? 'http://partner-bff', changeOrigin: true },
    },
  },
})
