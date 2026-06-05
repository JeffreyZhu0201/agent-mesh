import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@agentmesh/ui': path.resolve(__dirname, 'packages/ui/src'),
      '@agentmesh/hooks': path.resolve(__dirname, 'packages/hooks/src'),
      '@agentmesh/stores': path.resolve(__dirname, 'packages/stores/src'),
      '@agentmesh/api': path.resolve(__dirname, 'packages/api/src'),
      '@agentmesh/plugin-sdk': path.resolve(__dirname, 'packages/plugin-sdk/src'),
      '@agentmesh/types': path.resolve(__dirname, 'packages/types/src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
});
