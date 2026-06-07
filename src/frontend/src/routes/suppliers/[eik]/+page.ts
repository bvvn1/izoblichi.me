import type { PageLoad } from './$types';
import { api } from '$lib/api';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ fetch, params }) => {
	try {
		const [profile, network] = await Promise.all([
			api(fetch).supplier(params.eik),
			api(fetch).supplierNetwork(params.eik)
		]);
		return { profile, network };
	} catch {
		throw error(404, 'Доставчикът не е намерен');
	}
};
