<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import type { PageData } from './$types';
	import { getCityCoords, normalizeCity } from '$lib/cities';

	let { data }: { data: PageData } = $props();

	// Group buyers by normalized city name
	interface CityGroup {
		city: string;
		lat: number;
		lng: number;
		buyers: any;
		totalValueBGN: number;
		totalContracts: number;
		avgRiskScore: number;
		noBidCount: number;
	}

	function buildCityGroups(buyers: any): CityGroup[] {
		const map = new Map<string, { city: string; lat: number; lng: number; buyers: any }>();
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
	const fmtEur = (n: number) => fmt(n / 1.95583) + ' €';

	let mapEl: HTMLDivElement;
	let map: any; // typed as any since L is now dynamically imported
	let selectedCity: CityGroup | null = $state(null);
	let showUnmapped = $state(false);

	onMount(async () => {
		// ✅ Dynamic import keeps Leaflet out of SSR entirely
		const L = await import('leaflet');
		await import('leaflet/dist/leaflet.css');

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
				`<strong>${g.city}</strong><br>${fmtEur(g.totalValueBGN)}<br>${g.buyers.length} купувача`,
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
							<p class="text-sm font-medium">{fmtEur(selectedCity.totalValueBGN)}</p>
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
						{#each selectedCity.buyers.slice(0, 6) as buyer (buyer)}
							<li class="border-b border-base-content/8 pb-2 text-xs last:border-0 last:pb-0">
								<a
									href={resolve(`/buyers/${buyer.eik}`)}
									class="mb-0.5 line-clamp-2 block leading-snug font-medium hover:underline"
									>{buyer.name}</a
								>
								<span class="text-base-content/50"
									>{fmtEur(buyer.total_value_bgn)} · риск {buyer.risk_score.toFixed(0)}</span
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
								{#each unmappedBuyers.slice(0, 30) as buyer (buyer)}
									<li class="truncate text-xs text-base-content/50">{buyer.name}</li>
								{/each}
							</ul>
						{/if}
					</div>
				{/if}
			{/if}
		</div>
	</div>
</div>

<!-- Anomaly criteria -->
<section class="border-t border-base-content/15 py-12 md:py-14">
	<p class="mb-2 font-mono text-[10px] tracking-[.22em] text-base-content/40 uppercase">
		Методология
	</p>
	<h2 class="mb-8 text-xl font-bold md:text-2xl" style="font-family:'Playfair Display',serif">
		Как засичаме аномалиите
	</h2>

	<div
		class="grid grid-cols-1 gap-px overflow-hidden rounded-sm border border-base-content/10 bg-base-content/10 md:grid-cols-2"
	>
		<div class="bg-base-100 p-6 lg:p-7">
			<div class="mb-4 flex items-start gap-3">
				<div class="mt-0.5 shrink-0 text-[#B85C38]">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="h-4 w-4"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M15.75 15.75 21 21m-4.5-9a6.75 6.75 0 1 1-13.5 0 6.75 6.75 0 0 1 13.5 0Z"
						/>
					</svg>
				</div>
				<div>
					<p class="mb-0.5 text-sm font-semibold">Близо до прага</p>
					<p class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">
						Near threshold
					</p>
				</div>
			</div>
			<p class="mb-3 text-sm leading-relaxed text-base-content/60">
				Договорна стойност в рамките на <strong class="text-base-content">5%</strong> под прага за открита
				процедура — признак за умишлено занижаване, за да се избегне конкурс.
			</p>
			<div class="space-y-1">
				<div class="flex items-center justify-between rounded-sm bg-base-200/60 px-3 py-1.5">
					<span class="font-mono text-[10px] text-base-content/50 uppercase">Стоки / Услуги</span>
					<span class="font-mono text-[10px] font-medium">34 000 – 35 791 €</span>
				</div>
				<div class="flex items-center justify-between rounded-sm bg-base-200/60 px-3 py-1.5">
					<span class="font-mono text-[10px] text-base-content/50 uppercase">Строителство</span>
					<span class="font-mono text-[10px] font-medium">128 000 – 135 002 €</span>
				</div>
			</div>
		</div>

		<div class="bg-base-100 p-6 lg:p-7">
			<div class="mb-4 flex items-start gap-3">
				<div class="mt-0.5 shrink-0 text-[#B85C38]">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="h-4 w-4"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z"
						/>
					</svg>
				</div>
				<div>
					<p class="mb-0.5 text-sm font-semibold">Без конкуренция</p>
					<p class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">No bid</p>
				</div>
			</div>
			<p class="mb-3 text-sm leading-relaxed text-base-content/60">
				Поръчката е спечелена с <strong class="text-base-content">≤ 1 подадена оферта</strong> — без реална
				конкуренция. Системното повторение е по-силен сигнал от единичния случай.
			</p>
			<div class="flex items-center justify-between rounded-sm bg-base-200/60 px-3 py-1.5">
				<span class="font-mono text-[10px] text-base-content/50 uppercase">Условие</span>
				<span class="font-mono text-[10px] font-medium">bid_count ≤ 1</span>
			</div>
		</div>

		<div class="bg-base-100 p-6 lg:p-7">
			<div class="mb-4 flex items-start gap-3">
				<div class="mt-0.5 shrink-0 text-[#B85C38]">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="h-4 w-4"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M7.5 21 3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5"
						/>
					</svg>
				</div>
				<div>
					<p class="mb-0.5 text-sm font-semibold">Доминиращ доставчик</p>
					<p class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">
						Dominance
					</p>
				</div>
			</div>
			<p class="mb-3 text-sm leading-relaxed text-base-content/60">
				Един доставчик печели <strong class="text-base-content">≥ 80%</strong> от всички договори на даден
				купувач при минимум 5 спечелени — признак на зависимост или скрито договаряне.
			</p>
			<div class="space-y-1">
				<div class="flex items-center justify-between rounded-sm bg-base-200/60 px-3 py-1.5">
					<span class="font-mono text-[10px] text-base-content/50 uppercase">Дял по брой</span>
					<span class="font-mono text-[10px] font-medium">≥ 80%</span>
				</div>
				<div class="flex items-center justify-between rounded-sm bg-base-200/60 px-3 py-1.5">
					<span class="font-mono text-[10px] text-base-content/50 uppercase">Мин. договори</span>
					<span class="font-mono text-[10px] font-medium">≥ 5</span>
				</div>
			</div>
		</div>

		<div class="bg-base-100 p-6 lg:p-7">
			<div class="mb-4 flex items-start gap-3">
				<div class="mt-0.5 shrink-0 text-[#B85C38]">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="h-4 w-4"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99"
						/>
					</svg>
				</div>
				<div>
					<p class="mb-0.5 text-sm font-semibold">Системно повторение</p>
					<p class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">
						Repeated award
					</p>
				</div>
			</div>
			<p class="mb-3 text-sm leading-relaxed text-base-content/60">
				Двойка купувач–доставчик с трайна шарка: многогодишна история, множество договора и <strong
					class="text-base-content">≥ 50% без конкуренция</strong
				>.
			</p>
			<div class="space-y-1">
				<div class="flex items-center justify-between rounded-sm bg-base-200/60 px-3 py-1.5">
					<span class="font-mono text-[10px] text-base-content/50 uppercase">Активни години</span>
					<span class="font-mono text-[10px] font-medium">≥ 2</span>
				</div>
				<div class="flex items-center justify-between rounded-sm bg-base-200/60 px-3 py-1.5">
					<span class="font-mono text-[10px] text-base-content/50 uppercase">Общо договори</span>
					<span class="font-mono text-[10px] font-medium">≥ 3</span>
				</div>
				<div class="flex items-center justify-between rounded-sm bg-base-200/60 px-3 py-1.5">
					<span class="font-mono text-[10px] text-base-content/50 uppercase">Без конкуренция</span>
					<span class="font-mono text-[10px] font-medium">≥ 50%</span>
				</div>
			</div>
		</div>
	</div>
</section>
