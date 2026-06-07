import type { PageLoad } from './$types';
import { api } from '$lib/api';
import type { AnomalyType } from '$lib/api';
import { oneOf } from '$lib/field_validation';

const VALID_ANOMALY_TYPE: AnomalyType[] = [
	'near_threshold',
	'no_bid',
	'dominance',
	'repeated_award'
];

export const load: PageLoad = async ({ fetch, url }) => {
	const p = url.searchParams;

	const filters = {
		type: oneOf(p.get('type'), VALID_ANOMALY_TYPE, 'near_threshold' satisfies AnomalyType),
		year_from: p.get('year_from') ? Number(p.get('year_from')) : undefined,
		year_to: p.get('year_to') ? Number(p.get('year_to')) : undefined,
		page: p.get('page') ? Number(p.get('page')) : 1,
		buyer_eik: p.get('buyer_eik') ?? undefined,
		supplier_eik: p.get('supplier_eik') ?? undefined
	};

	const combinedFilters = { ...filters, per_page: 25 };

	const anomalies = await api(fetch).anomalies(combinedFilters);

	return {
		filters,
		anomalies
	};
};
