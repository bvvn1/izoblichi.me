import type { PageLoad } from './$types';
import { api } from '$lib/api';

export const load: PageLoad = async ({ fetch, url }) => {
	const q = url.searchParams.get('q') || undefined;
	const page = Number(url.searchParams.get('page')) || 1;

	const result = await api(fetch).parties({
		q,
		role: 'buyer',
		page,
		per_page: 25
	});

	return { result, filters: { q, page } };
};
