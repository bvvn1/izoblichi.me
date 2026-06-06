import type { PageLoad } from './$types';
import { api } from '$lib/api';

export const load: PageLoad = async ({ fetch, url }) => {
	const q = url.searchParams.get('q') || undefined;
	const year_from = url.searchParams.get('year_from') || undefined;
	const year_to = url.searchParams.get('year_to') || undefined;
	const category = url.searchParams.get('category') || undefined;
	const source = url.searchParams.get('source') || undefined;
	const sort_by = url.searchParams.get('sort_by') || undefined;
	const sort_dir = url.searchParams.get('sort_dir') || undefined;
	const page = Number(url.searchParams.get('page')) || 1;

	const result = await api(fetch).contracts({
		q: q,
		year_from: year_from ? Number(year_from) : undefined,
		year_to: year_to ? Number(year_to) : undefined,
		category: category,
		source: source as 'legacy' | 'ocds' | undefined,
		sort_by: (sort_by as 'contract_date' | 'contract_value' | 'year' | 'buyer_name' | 'supplier_name') ?? 'contract_date',
		sort_dir: (sort_dir as 'asc' | 'desc') ?? 'desc',
		page,
		per_page: 25
	});

	return {
		result,
		filters: {
			q: q,
			year_from: year_from ? Number(year_from) : undefined,
			year_to: year_to ? Number(year_to) : undefined,
			category: category,
			source: source,
			sort_by: sort_by,
			sort_dir: sort_dir,
			page
		}
	};
};
