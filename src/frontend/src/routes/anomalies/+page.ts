import type { PageLoad } from './$types';
import { api } from '$lib/api';
import type { AnomalyType } from '$lib/api';

export const load: PageLoad = async ({ fetch, url }) => {
	const type = (url.searchParams.get('type') as AnomalyType) || 'near_threshold';
	const year_from = url.searchParams.get('year_from') || undefined;
	const year_to = url.searchParams.get('year_to') || undefined;
	const page = Number(url.searchParams.get('page')) || 1;

	const result = await api(fetch).anomalies({
		type,
		year_from: year_from ? Number(year_from) : undefined,
		year_to: year_to ? Number(year_to) : undefined,
		page,
		per_page: 25
	});

	return {
		result,
		filters: {
			type,
			year_from: year_from ? Number(year_from) : undefined,
			year_to: year_to ? Number(year_to) : undefined,
			page
		}
	};
};
