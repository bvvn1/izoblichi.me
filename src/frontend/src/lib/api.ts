type Fetch = typeof fetch

export interface MapBuyer {
  eik: string
  name: string
  address_locality: string | null
  address_region: string | null
  total_contracts: number
  total_value_bgn: number
  no_bid_count: number
  near_threshold_count: number
  risk_score: number
}

export interface MapBuyersResponse {
  total: number
  items: MapBuyer[]
}

const BASE_URL = import.meta.env.VITE_API_BASE ?? 'http://localhost:3000'

function client(fetch: Fetch) {
  async function get<T>(path: string, params?: Record<string, string | number | undefined>): Promise<T> {
    const url = new URL(BASE_URL + path)
    if (params) {
      for (const [key, value] of Object.entries(params)) {
        if (value !== undefined && value !== '') url.searchParams.set(key, String(value))
      }
    }
    const response = await fetch(url)
    if (!response.ok) throw new Error(`${response.status} ${url.pathname}`)
    return response.json()
  }

  return {
    stats: ()                      => get<Stats>('/stats'),
    contracts: (p?: ContractParams) => get<ContractList>('/contracts', p as Record<string, string | number | undefined>),
    buyer: (eik: string)           => get<BuyerProfile>(`/buyers/${eik}`),
    supplier: (eik: string)        => get<SupplierProfile>(`/suppliers/${eik}`),
    parties: (p?: PartyParams)     => get<PartyList>('/parties', p as Record<string, string | number | undefined>),
    party: (eik: string)           => get<Party>(`/parties/${eik}`),
    anomalies: ()                  => get<Anomalies>('/anomalies'),
    mapBuyers: ()                  => get<MapBuyersResponse>('/map/buyers'),
  }
}

export { client as api }