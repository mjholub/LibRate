import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';
import { resolve } from 'path';

export default defineConfig({
  plugins: [sveltekit()],
  base: './',
  test: {
    include: ['src/**/*.{test,spec}.{js,ts}']
  },
  build: {
    sourcemap: true,
  },
  resolve: {
    alias: {
      $components: resolve(__dirname, 'src/components'),
      $stores: resolve(__dirname, 'src/stores'),
      $types: resolve(__dirname, 'src/types'),
    },
  },
});
