import type { PageLoad } from './$types';
import { api } from '$lib/api';

export const load: PageLoad = async ({ fetch }) => {
	const mapData = await api(fetch).mapBuyers();
	return { mapData };
};
