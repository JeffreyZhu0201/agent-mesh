import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

const root = path.resolve(__dirname, '../..');

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@agentmesh/ui': path.resolve(root, 'packages/ui/src'),
      '@agentmesh/hooks': path.resolve(root, 'packages/hooks/src'),
      '@agentmesh/stores': path.resolve(root, 'packages/stores/src'),
      '@agentmesh/api': path.resolve(root, 'packages/api/src'),
      '@agentmesh/plugin-sdk': path.resolve(root, 'packages/plugin-sdk/src'),
      '@agentmesh/types': path.resolve(root, 'packages/types/src'),
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
