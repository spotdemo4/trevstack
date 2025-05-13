import { sveltekit } from '@sveltejs/kit/vite';
import tailwindcss from '@tailwindcss/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [tailwindcss(), sveltekit()],
	server: {
		proxy: {
			'/grpc': {
				target: 'http://localhost:8080',
				changeOrigin: true
			},
			'/file': {
				target: 'http://localhost:8080',
				changeOrigin: true
			}
		},
		fs: {
			allow: ['package.json']
		},
		host: '0.0.0.0'
	}
});
