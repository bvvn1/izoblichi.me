import type { LayoutLoad } from './$types';
import { api } from '$lib/api';

export const load: LayoutLoad = async ({ fetch }) => {
	const stats = await api(fetch).stats();
	return { stats };
};
