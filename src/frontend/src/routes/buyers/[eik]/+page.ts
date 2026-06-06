import type { PageLoad } from './$types';
import { api } from '$lib/api';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ fetch, params }) => {
	try {
		const profile = await api(fetch).buyer(params.eik);
		return { profile };
	} catch {
		throw error(404, 'Купувачът не е намерен');
	}
};
