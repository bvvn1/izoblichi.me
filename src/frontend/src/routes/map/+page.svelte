<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { PageData } from './$types';
	import type { MapBuyer } from '$lib/api';
	import { getCityCoords, normalizeCity } from '$lib/cities';

	let { data }: { data: PageData } = $props();

	// Group buyers by normalized city name
	interface CityGroup {
		city: string;
		lat: number;
		lng: number;
		buyers: MapBuyer[];
		totalValueBGN: number;
		totalContracts: number;
		avgRiskScore: number;
		noBidCount: number;
	}

	function buildCityGroups(buyers: MapBuyer[]): CityGroup[] {
		const map = new Map<string, { city: string; lat: number; lng: number; buyers: MapBuyer[] }>();
		for (const b of buyers) {
			const coords = getCityCoords(b.address_locality);
			if (!coords) continue;
			const city = normalizeCity(b.address_locality ?? '');
			if (!map.has(city)) map.set(city, { city, lat: coords[0], lng: coords[1], buyers: [] });
			map.get(city)!.buyers.push(b);
		}
		return Array.from(map.values()).map((g) => {
			const totalValueBGN = g.buyers.reduce((s, b) => s + b.total_value_bgn, 0);
			const totalContracts = g.buyers.reduce((s, b) => s + b.total_contracts, 0);
			const noBidCount = g.buyers.reduce((s, b) => s + b.no_bid_count, 0);
			const avgRiskScore =
				g.buyers.reduce((s, b) => s + b.risk_score * b.total_contracts, 0) / (totalContracts || 1);
			return { ...g, totalValueBGN, totalContracts, avgRiskScore, noBidCount };
		});
	}

	const cityGroups = buildCityGroups(data.mapData.items);
	const unmappedBuyers = data.mapData.items.filter((b) => !getCityCoords(b.address_locality));

	const maxValue = Math.max(...cityGroups.map((g) => g.totalValueBGN));

	function riskColor(score: number): string {
		if (score >= 40) return '#B85C38';
		if (score >= 20) return '#D4913A';
		return '#4A7C59';
	}

	function bubbleRadius(value: number): number {
		const MIN = 8,
			MAX = 52;
		return MIN + (MAX - MIN) * Math.sqrt(value / maxValue);
	}

	const fmt = (n: number) =>
		Intl.NumberFormat('bg-BG', { notation: 'compact', maximumFractionDigits: 1 }).format(n);
	const fmtBgn = (n: number) => fmt(n) + ' лв.';

	let mapEl: HTMLDivElement;
	let map: import('leaflet').Map;
	let selectedCity: CityGroup | null = $state(null);
	let showUnmapped = $state(false);

	onMount(async () => {
		const L = await import('leaflet');

		// Leaflet CSS
		const link = document.createElement('link');
		link.rel = 'stylesheet';
		link.href = 'https://unpkg.com/leaflet@1.9.4/dist/leaflet.css';
		document.head.appendChild(link);

		map = L.map(mapEl, {
			center: [42.73, 25.48],
			zoom: 7,
			zoomControl: true
		});

		L.tileLayer('https://{s}.basemaps.cartocdn.com/light_all/{z}/{x}/{y}{r}.png', {
			attribution: '© OpenStreetMap © CARTO',
			maxZoom: 18
		}).addTo(map);

		for (const g of cityGroups) {
			const radius = bubbleRadius(g.totalValueBGN);
			const color = riskColor(g.avgRiskScore);

			const circle = L.circleMarker([g.lat, g.lng], {
				radius,
				fillColor: color,
				fillOpacity: 0.72,
				color: '#fff',
				weight: 1.5
			}).addTo(map);

			circle.bindTooltip(
				`<strong>${g.city}</strong><br>${fmtBgn(g.totalValueBGN)}<br>${g.buyers.length} купувача`,
				{ sticky: true, className: 'leaflet-tooltip-dark' }
			);

			circle.on('click', () => {
				selectedCity = g;
			});
		}
	});

	onDestroy(() => {
		map?.remove();
	});
</script>

<svelte:head>
	<title>Карта на купувачите — изобличи.ме</title>
</svelte:head>

<section class="border-b border-base-content/15 py-10">
	<p class="mb-3 font-mono text-xs tracking-[.14em] text-base-content/50 uppercase">Карта</p>
	<h1
		class="font-display mb-2 text-3xl leading-tight font-bold"
		style="font-family:'Playfair Display',serif"
	>
		Купувачи по местоположение
	</h1>
	<p class="max-w-lg text-sm text-base-content/60">
		Всяка точка е град. Размерът показва общата стойност на поръчките. Цветът показва риска: <span
			style="color:#B85C38"
			class="font-medium">червено = висок</span
		>,
		<span style="color:#D4913A" class="font-medium">кехлибар = среден</span>,
		<span style="color:#4A7C59" class="font-medium">зелено = нисък</span>.
	</p>
</section>

<div class="py-6">
	<div class="flex flex-col gap-6 lg:flex-row">
		<!-- Map -->
		<div class="min-w-0 flex-1">
			<div
				bind:this={mapEl}
				class="w-full rounded-sm border border-base-content/15"
				style="height: min(540px, 60vw); min-height: 300px"
			></div>
			<p class="mt-2 font-mono text-xs text-base-content/30">
				{cityGroups.length} градове · {data.mapData.items.length} купувача
			</p>
		</div>

		<!-- Side panel -->
		<div class="w-full shrink-0 lg:w-72">
			{#if selectedCity}
				<div class="rounded-sm border border-base-content/15 p-4">
					<div class="mb-3 flex items-start justify-between">
						<h2 class="text-sm font-medium">{selectedCity.city}</h2>
						<button
							onclick={() => (selectedCity = null)}
							class="font-mono text-xs text-base-content/40 transition-colors hover:text-base-content"
							>✕</button
						>
					</div>

					<div class="mb-4 grid grid-cols-2 gap-3">
						<div>
							<p class="mb-0.5 font-mono text-xs text-base-content/40">Стойност</p>
							<p class="text-sm font-medium">{fmtBgn(selectedCity.totalValueBGN)}</p>
						</div>
						<div>
							<p class="mb-0.5 font-mono text-xs text-base-content/40">Договори</p>
							<p class="text-sm font-medium">{fmt(selectedCity.totalContracts)}</p>
						</div>
						<div>
							<p class="mb-0.5 font-mono text-xs text-base-content/40">Без конкуренция</p>
							<p class="text-sm font-medium">{fmt(selectedCity.noBidCount)}</p>
						</div>
						<div>
							<p class="mb-0.5 font-mono text-xs text-base-content/40">Рисков индекс</p>
							<p class="text-sm font-medium" style="color:{riskColor(selectedCity.avgRiskScore)}">
								{selectedCity.avgRiskScore.toFixed(1)}
							</p>
						</div>
					</div>

					<p class="mb-2 font-mono text-xs tracking-wider text-base-content/40 uppercase">
						Топ купувачи
					</p>
					<ul class="space-y-2">
						{#each selectedCity.buyers.slice(0, 6) as buyer}
							<li class="border-b border-base-content/8 pb-2 text-xs last:border-0 last:pb-0">
								<a
									href="/buyers/{buyer.eik}"
									class="mb-0.5 line-clamp-2 block leading-snug font-medium hover:underline"
									>{buyer.name}</a
								>
								<span class="text-base-content/50"
									>{fmtBgn(buyer.total_value_bgn)} · риск {buyer.risk_score.toFixed(0)}</span
								>
							</li>
						{/each}
						{#if selectedCity.buyers.length > 6}
							<li class="font-mono text-xs text-base-content/40">
								+{selectedCity.buyers.length - 6} повече
							</li>
						{/if}
					</ul>
				</div>
			{:else}
				<div class="rounded-sm border border-base-content/10 bg-base-200/40 p-4">
					<p class="mb-3 font-mono text-xs tracking-wider text-base-content/40 uppercase">
						Легенда
					</p>
					<ul class="space-y-2 text-xs">
						<li class="flex items-center gap-2">
							<span class="h-3 w-3 shrink-0 rounded-full" style="background:#B85C38"></span>
							Висок риск (≥40)
						</li>
						<li class="flex items-center gap-2">
							<span class="h-3 w-3 shrink-0 rounded-full" style="background:#D4913A"></span>
							Среден риск (20-40)
						</li>
						<li class="flex items-center gap-2">
							<span class="h-3 w-3 shrink-0 rounded-full" style="background:#4A7C59"></span>
							Нисък риск (&lt;20)
						</li>
					</ul>
					<p class="mt-4 font-mono text-xs leading-relaxed text-base-content/30">
						Кликнете върху точка за детайли за купувачите в съответния град.
					</p>
				</div>

				{#if unmappedBuyers.length > 0}
					<div class="mt-3">
						<button
							onclick={() => (showUnmapped = !showUnmapped)}
							class="font-mono text-xs text-base-content/40 transition-colors hover:text-base-content"
						>
							{showUnmapped ? '▾' : '▸'}
							{unmappedBuyers.length} без координати
						</button>
						{#if showUnmapped}
							<ul class="mt-2 max-h-48 space-y-1 overflow-y-auto">
								{#each unmappedBuyers.slice(0, 30) as b}
									<li class="truncate text-xs text-base-content/50">{b.name}</li>
								{/each}
							</ul>
						{/if}
					</div>
				{/if}
			{/if}
		</div>
	</div>
</div>
