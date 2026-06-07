import type { PageLoad } from './$types';
import { api } from '$lib/api';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ fetch, params }) => {
	try {
		const party = await api(fetch).party(params.eik);
		return { party };
	} catch {
		throw error(404, 'Юридическото лице не е намерено');
	}
};
