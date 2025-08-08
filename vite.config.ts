import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    target: 'node18',
    outDir: 'dist',
    ssr: true,
    rollupOptions: {
      input: {
        main: 'src/main.ts',
        server: 'src/server.ts'
      },
      external: [
        '@modelcontextprotocol/sdk',
        'axios',
        'dotenv',
        'express',
        'cors',
        'child_process',
        'path',
        'url'
      ],
      output: {
        format: 'es',
        entryFileNames: '[name].js',
        banner: (chunk) => {
          if (chunk.name === 'main') {
            return '#!/usr/bin/env node';
          }
          return '';
        }
      }
    },
    minify: false
  },
  optimizeDeps: {
    exclude: [
      '@modelcontextprotocol/sdk',
      'axios',
      'dotenv',
      'express',
      'cors'
    ]
  },
  esbuild: {
    platform: 'node'
  }
});