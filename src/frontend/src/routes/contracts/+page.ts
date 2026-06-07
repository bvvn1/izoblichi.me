import type { PageLoad } from './$types';
import { api } from '$lib/api';
import type { SortBy, SortDir, DataSource } from '$lib/api';
import { oneOf } from '$lib/field_validation';

const VALID_SORT_BY: SortBy[] = [
	'contract_date',
	'contract_value',
	'year',
	'buyer_name',
	'supplier_name'
];
const VALID_SORT_DIR: SortDir[] = ['asc', 'desc'];
const VALID_SOURCE: DataSource[] = ['legacy', 'ocds'];

export const load: PageLoad = async ({ fetch, url }) => {
	const p = url.searchParams;

	const filters = {
		q: p.get('q') ?? undefined,
		year_from: p.get('year_from') ? Number(p.get('year_from')) : undefined,
		year_to: p.get('year_to') ? Number(p.get('year_to')) : undefined,
		category: p.get('category') ?? '',
		source: oneOf(p.get('source'), VALID_SOURCE, undefined),
		sort_by: oneOf(p.get('sort_by'), VALID_SORT_BY, 'contract_date' satisfies SortBy),
		sort_dir: oneOf(p.get('sort_dir'), VALID_SORT_DIR, 'desc' satisfies SortDir),
		buyer_eik: p.get('buyer_eik') ?? undefined,
		supplier_eik: p.get('supplier_eik') ?? undefined,
		page: p.get('page') ? Number(p.get('page')) : 1
	};

	const contracts = await api(fetch).contracts(filters);

	return {
		filters,
		contracts
	};
};
