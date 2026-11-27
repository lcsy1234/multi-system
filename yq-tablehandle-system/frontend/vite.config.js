import { defineConfig } from 'vite';

export default defineConfig({
  worker: { format: 'iife', rollupOptions: { output: { inlineDynamicImports: true } } },
  server: {
    host: '127.0.0.1', port: 5175, strictPort: true,
    proxy: { '/api': 'http://127.0.0.1:8191' },
  },
});
