import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    target: 'node18',
    outDir: 'dist',
    ssr: true,
    lib: {
      entry: 'src/main.ts',
      formats: ['es'],
      fileName: 'main'
    },
    rollupOptions: {
      external: ['express', 'cors'],
      output: {
        banner: '#!/usr/bin/env node'
      }
    },
    minify: false
  },
  optimizeDeps: {
    exclude: ['express', 'cors']
  },
  esbuild: {
    platform: 'node'
  }
});