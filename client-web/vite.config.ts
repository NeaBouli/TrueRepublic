import { defineConfig } from 'vitest/config'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 3001,
  },
  build: {
    manifest: true,
    // Raw protobuf code compresses by almost 8:1. The stricter project gate
    // checks raw and gzip entry, route, chunk, and total sizes after each build.
    chunkSizeWarningLimit: 1100,
  },
  test: {
    globals: true,
    environment: 'happy-dom',
    setupFiles: './src/test/setup.ts',
  },
})
