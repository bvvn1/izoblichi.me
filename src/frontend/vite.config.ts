import tailwindcss from '@tailwindcss/vite';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

const viteConfig = defineConfig({ plugins: [tailwindcss(), sveltekit()] });

export default viteConfig;
