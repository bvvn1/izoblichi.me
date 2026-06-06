import type {
	StatsResponse,
	TimeseriesResponse,
	ContractsResponse,
	Contract,
	AnnexesResponse,
	PartiesResponse,
	PartyDetail,
	BuyerProfile,
	SupplierProfile,
	AnomaliesResponse,
	NetworkResponse,
	MapBuyersResponse,
	AutocompleteResponse
} from '$lib/types';

type Fetch = typeof globalThis.fetch;

const BASE_URL = import.meta.env.VITE_API_BASE ?? 'http://localhost:3000';

type Params = Record<string, string | number | boolean | undefined>;

export type SortBy = 'contract_date' | 'contract_value' | 'year' | 'buyer_name' | 'supplier_name';
export type SortDir = 'asc' | 'desc';
export type DataSource = 'legacy' | 'ocds';
export type AnomalyType = 'near_threshold' | 'no_bid' | 'dominance' | 'repeated_award';

export interface PaginationParams {
	page?: number;
	per_page?: number;
}

export interface ContractParams extends PaginationParams {
	q?: string;
	buyer_eik?: string;
	supplier_eik?: string;
	year_from?: number;
	year_to?: number;
	min_value?: number;
	max_value?: number;
	category?: string;
	source?: DataSource;
	sort_by?: SortBy;
	sort_dir?: SortDir;
}

export interface AnnexParams extends PaginationParams {
	q?: string;
	buyer_eik?: string;
	supplier_eik?: string;
	procurement_number?: string;
	year_from?: number;
	year_to?: number;
}

export interface AnomalyParams extends PaginationParams {
	type?: AnomalyType;
	buyer_eik?: string;
	supplier_eik?: string;
	year_from?: number;
	year_to?: number;
}

export interface PartyParams extends PaginationParams {
	q?: string;
	role?: 'buyer' | 'supplier';
}

function client(fetch: Fetch) {
	async function get<T>(path: string, params?: Params): Promise<T> {
		const url = new URL(BASE_URL + path);
		if (params) {
			for (const [key, value] of Object.entries(params)) {
				if (value !== undefined && value !== '') url.searchParams.set(key, String(value));
			}
		}
		const response = await fetch(url);
		if (!response.ok) throw new Error(`${response.status} ${url.pathname}`);
		return response.json();
	}

	// CSV endpoints return a Blob — trigger download with URL.createObjectURL
	async function download(path: string, params?: Params): Promise<Blob> {
		const url = new URL(BASE_URL + path);
		if (params) {
			for (const [key, value] of Object.entries(params)) {
				if (value !== undefined && value !== '') url.searchParams.set(key, String(value));
			}
		}
		const response = await fetch(url, { headers: { Accept: 'text/csv' } });
		if (!response.ok) throw new Error(`${response.status} ${url.pathname}`);
		return response.blob();
	}

	return {
		stats: () => get<StatsResponse>('/stats'),
		timeseries: () => get<TimeseriesResponse>('/stats/timeseries'),

		contracts: (params?: ContractParams) => get<ContractsResponse>('/contracts', params as Params),
		contract: (source: DataSource, rowKey: string) =>
			get<Contract>(`/contracts/${source}/${rowKey}`),
		exportContracts: (params?: ContractParams) => download('/export/contracts', params as Params),

		annexes: (params?: AnnexParams) => get<AnnexesResponse>('/annexes', params as Params),

		parties: (params?: PartyParams) => get<PartiesResponse>('/parties', params as Params),
		party: (eik: string) => get<PartyDetail>(`/parties/${eik}`),

		buyer: (eik: string) => get<BuyerProfile>(`/buyers/${eik}`),
		supplier: (eik: string) => get<SupplierProfile>(`/suppliers/${eik}`),

		anomalies: (params?: AnomalyParams) => get<AnomaliesResponse>('/anomalies', params as Params),
		exportAnomalies: (
			type: AnomalyType,
			params?: Omit<AnomalyParams, 'type' | 'page' | 'per_page'>
		) => download(`/export/anomalies/${type}`, params as Params),

		buyerNetwork: (eik: string) => get<NetworkResponse>(`/network/buyer/${eik}`),
		supplierNetwork: (eik: string) => get<NetworkResponse>(`/network/supplier/${eik}`),

		mapBuyers: () => get<MapBuyersResponse>('/map/buyers'),

		autocomplete: (q: string) => get<AutocompleteResponse>('/search/autocomplete', { q })
	};
}

export { client as api };
