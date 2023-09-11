import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/kit/vite';

const customBuild = ({ routes }) => ({
  // Customize the output of each route
  async build({ entryPoints, manifest }) {
    for (const [url, { entry }] of entryPoints) {
      if (url !== '/') {
        // Modify the entry HTML of non-root routes
        entry.html = entry.html.replace(/\.\/_app\//g, '../_app/');

      }
    }
  },
});

/** @type {import('@sveltejs/kit').Config} */
const config = {
  // Consult https://kit.svelte.dev/docs/integrations#preprocessors
  // for more information about preprocessors
  preprocess: vitePreprocess(),
  kit: {
    adapter: adapter({
      // default options are shown. On some platforms
      // these options are set automatically — see below
      paths: {
        assets: './static',
        profiles: '',
      },
      pages: 'build',
      assets: 'build',
      fallback: '404.html',
      precompress: false,
      strict: true
    }),
    alias: {
      $components: './src/components',
      $stores: './src/stores',
      $types: './src/types'
    }
  },

  vitePlugin: {
    inspector: true
  },
  target: '#svelte',
  vite: {
    plugins: [customBuild],
  }
};

export default config;
