export interface Annex {
	row_key: number;
	year: number;
	doc_number: string | null;
	contract_number: string | null;
	contract_date: string | null;
	published_date: string | null;
	procurement_number: string | null;
	buyer_eik: string | null;
	buyer_name: string | null;
	procurement_subject: string | null;
	procurement_category: string | null;
	eu_funded: boolean | null;
	contract_subject: string | null;
	supplier_eik: string | null;
	supplier_name: string | null;
	value_before: number | null;
	value_after: number | null;
	value_change: number | null;
	currency: string | null;
	amendment_description: string | null;
	amendment_reason: string | null;
	circumstances: string | null;
}

export interface AnnexesResponse {
	total: number;
	page: number;
	per_page: number;
	items: Annex[];
}

export interface Contract {
	data_source: string;
	row_key: string;
	ocid: string | null;
	year: number | null;
	contract_date: string | null;
	buyer_eik: string | null;
	buyer_name: string | null;
	supplier_eik: string | null;
	supplier_name: string | null;
	contract_value: number | null;
	currency: string | null;
	procurement_number: string | null;
	title: string | null;
	procurement_category: string | null;
	bid_count: number | null;
	procurement_method: string | null;
	eu_funded: boolean | null;
	legacy_type?: string | null;
}

export interface ContractsResponse {
	total: number;
	page: number;
	per_page: number;
	items: Contract[];
}

export interface Party {
	eik: string;
	legal_name: string | null;
	display_name: string | null;
	address_locality: string | null;
	address_region: string | null;
	country: string | null;
	first_seen: string | null;
	last_seen: string | null;
}

export interface PartiesResponse {
	total: number;
	page: number;
	per_page: number;
	items: Party[];
}

export interface PartyRole {
	contract_count: number;
}

export interface PartyDetail extends Party {
	as_buyer: PartyRole;
	as_supplier: PartyRole;
}

export interface TopCounterpart {
	eik: string | null;
	name: string | null;
	wins: number;
	total_value: number | null;
	pct_by_count: number;
}

export interface YearBreakdown {
	year: number;
	contract_count: number;
	total_value: number | null;
}

export interface BuyerFlags {
	supplier_dominance: boolean;
	dominant_supplier_eik?: string | null;
	dominant_supplier_pct: number;
	no_bid_count: number;
	near_threshold_count: number;
}

export interface BuyerProfile {
	eik: string;
	name: string;
	address_locality: string | null;
	address_region: string | null;
	total_contracts: number;
	total_value: number | null;
	currency: string;
	years_active: number[];
	top_suppliers: TopCounterpart[];
	year_breakdown: YearBreakdown[];
	flags: BuyerFlags;
}

export interface SupplierFlags {
	buyer_concentration: boolean;
	top_buyer_pct: number;
}

export interface SupplierProfile {
	eik: string;
	name: string;
	address_locality: string | null;
	address_region: string | null;
	total_wins: number;
	total_value: number | null;
	currency: string;
	top_buyers: TopCounterpart[];
	year_breakdown: YearBreakdown[];
	flags: SupplierFlags;
}

export interface NearThresholdAnomaly {
	contract_value: number;
	threshold: number;
	gap: number;
	gap_pct: number;
	buyer_eik: string | null;
	buyer_name: string | null;
	supplier_eik: string | null;
	supplier_name: string | null;
	contract_date: string | null;
	title: string | null;
	procurement_category: string | null;
}

export interface NoBidAnomaly {
	bid_count: number | null;
	contract_value: number | null;
	currency: string | null;
	buyer_eik: string | null;
	buyer_name: string | null;
	supplier_eik: string | null;
	supplier_name: string | null;
	contract_date: string | null;
	title: string | null;
}

export interface DominanceAnomaly {
	buyer_eik: string | null;
	buyer_name: string | null;
	supplier_eik: string | null;
	supplier_name: string | null;
	wins: number;
	total_value: number | null;
	pct_by_count: number;
	total_buyer_contracts: number;
}

export interface RepeatedAwardAnomaly {
	buyer_eik: string | null;
	buyer_name: string | null;
	supplier_eik: string | null;
	supplier_name: string | null;
	total_wins: number;
	years_active: number;
	total_value: number | null;
	years: number[];
}

export interface AnomalySection<T> {
	total: number;
	page: number;
	per_page: number;
	items: T[];
}

export interface AnomaliesResponse {
	near_threshold: AnomalySection<NearThresholdAnomaly>;
	no_bid: AnomalySection<NoBidAnomaly>;
	dominance: AnomalySection<DominanceAnomaly>;
	repeated_award: AnomalySection<RepeatedAwardAnomaly>;
}

export interface MapBuyer {
	eik: string;
	name: string;
	address_locality: string | null;
	address_region: string | null;
	total_contracts: number;
	total_value_bgn: number;
	no_bid_count: number;
	near_threshold_count: number;
	risk_score: number;
}

export interface MapBuyersResponse {
	total: number;
	items: MapBuyer[];
}

export interface NetworkNode {
	id: string;
	label: string;
	type: string;
	size: number;
}

export interface NetworkEdge {
	source: string;
	target: string;
	weight: number;
	label: string;
}

export interface NetworkResponse {
	nodes: NetworkNode[];
	edges: NetworkEdge[];
}

export interface AutocompleteParty {
	eik: string;
	name: string;
}

export interface AutocompleteContract {
	id: string;
	source: string;
	title: string;
}

export interface AutocompleteResponse {
	buyers: AutocompleteParty[];
	suppliers: AutocompleteParty[];
	contracts: AutocompleteContract[];
}

export interface MonthStat {
	year: number;
	month: number;
	contract_count: number;
	total_value_bgn: number;
	no_bid_count: number;
	competitive_count: number;
}

export interface TimeseriesResponse {
	months: MonthStat[];
}

export interface YearStat {
	year: number;
	contract_count: number;
	total_value: number | null;
}

export interface StatsResponse {
	total_contracts: number;
	total_value_bgn: number | null;
	unique_buyers: number;
	unique_suppliers: number;
	year_breakdown: YearStat[];
	data_through: string | null;
}
