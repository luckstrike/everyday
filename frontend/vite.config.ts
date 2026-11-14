import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';

// The API checks that a request's Origin matches its Host before it lets anything
// be changed. Vite's string shorthand for a proxy rewrites Host to the target, so
// the proxy is written out here to keep the page's own Host.
const api = { '/api': { target: 'http://127.0.0.1:8080', changeOrigin: false } };

export default defineConfig({
  plugins: [svelte()],
  server: { port: 5173, strictPort: true, proxy: api },
  preview: { port: 4173, proxy: api },
});
