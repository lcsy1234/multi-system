import { defineConfig } from 'vite';

// A classic script works from file:// when recipients double-click index.html.
export default defineConfig({
  define: { 'process.env.NODE_ENV': JSON.stringify('production') },
  build: {
    outDir: '../exports/config-json-offline',
    emptyOutDir: false,
    lib: { entry: 'src/offline.tsx', name: 'OfflineConverterApp', formats: ['iife'], fileName: () => 'app.js' },
  },
});
