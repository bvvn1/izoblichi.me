import type { PageLoad } from './$types';
import { api, type PartyParams } from '$lib/api';

export const load: PageLoad = async ({ fetch, url }) => {
	const p = url.searchParams;

	const filters = {
		q: p.get('q') ?? undefined,
		page: p.get('page') ? Number(p.get('page')) : 1
	};

	const combinedFilters: PartyParams = { ...filters, per_page: 25, role: undefined };

	const response = await api(fetch).parties(combinedFilters);

	return { response, filters };
};
