import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./vitest.setup.ts'],
    include: ['packages/**/*.test.ts'],
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
