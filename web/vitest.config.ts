import path from 'path';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  resolve: {
    alias: {
      '@agentmesh/api': path.resolve(__dirname, 'packages/api/src'),
      '@agentmesh/stores': path.resolve(__dirname, 'packages/stores/src'),
      '@agentmesh/plugin-sdk': path.resolve(__dirname, 'packages/plugin-sdk/src'),
      '@agentmesh/ui': path.resolve(__dirname, 'packages/ui/src'),
      react: path.resolve(__dirname, 'node_modules/react'),
      'react-dom': path.resolve(__dirname, 'node_modules/react-dom'),
      'react-router-dom': path.resolve(__dirname, 'node_modules/react-router-dom'),
    },
  },
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./vitest.setup.ts'],
    include: ['packages/**/*.test.ts', 'apps/**/src/**/*.test.{ts,tsx}'],
    coverage: {
      provider: 'v8',
      include: [
        'packages/stores/src/**/*.ts',
        'packages/plugin-sdk/src/**/*.ts',
        'packages/api/src/**/*.ts',
      ],
      exclude: ['packages/**/src/index.ts', 'packages/**/src/types.ts'],
      thresholds: {
        lines: 50,
        functions: 50,
        statements: 50,
      },
    },
  },
});
