import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';
import adapter from '@sveltejs/adapter-static';
//import { vitePreprocess } from '@sveltejs/kit/vite';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	// Consult https://kit.svelte.dev/docs/integrations#preprocessors
	// for more information about preprocessors
	preprocess: [vitePreprocess({})],
	kit: {
		prerender: {
			handleHttpError: 'warn',
			// TODO: whitelist entries array with svelte's own routes (e.g. not swagger or anything provided by the server)
			crawl: false
		},
		adapter: adapter({
			// default options are shown. On some platforms
			// these options are set automatically — see below
			paths: {
				assets: './static',
				profiles: ''
			},
			pages: 'build',
			assets: 'build',
			fallback: '404.html',
			precompress: false,
			strict: false
		}),
		alias: {
			$components: './src/components',
			$stores: './src/stores',
			$lib: './src/lib'
		}
	},

	vitePlugin: {
		inspector: true
	},
	target: '#svelte'
};

export default config;
