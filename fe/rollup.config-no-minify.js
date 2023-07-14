import svelte from 'rollup-plugin-svelte';
import commonjs from '@rollup/plugin-commonjs';
import resolve from '@rollup/plugin-node-resolve';
import css from 'rollup-plugin-css-only';
import sveltePreprocess from 'svelte-preprocess';
import typescript from '@rollup/plugin-typescript';

const production = !process.env.ROLLUP_WATCH;

export default {
  input: 'src/main.ts',
  output: {
    sourcemap: true,
    format: 'iife',
    name: 'app',
    file: 'public/build/bundle.js'
  },
  plugins: [
    svelte({
      preprocess: sveltePreprocess({ sourceMap: !production }),
      compilerOptions: {
        // enable run-time checks when not in production
        dev: !production
      }
    }),
    css({ output: 'bundle.css' }),
    resolve({
      browser: true,
      dedupe: ['svelte'],
      exportConditions: ['svelte']
    }),
    commonjs(),
    typescript({
      sourceMap: !production,
      inlineSources: !production
    })
  ],
  watch: {
    clearScreen: false
  }
};
