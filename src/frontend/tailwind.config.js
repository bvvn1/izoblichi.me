import daisyui from 'daisyui';

const twConfig = {
	content: ['./src/**/*.{html,js,svelte,ts}'],
	plugins: [daisyui],
	daisyui: {
		themes: ['light', 'dark']
	},
	theme: {
		extend: {
			fontFamily: {
				display: ['"Playfair Display"', '"Plus Jakarta Sans', 'serif'],
				mono: ['"IBM Plex Mono"', 'monospace']
			}
		}
	}
};

export default twConfig;
