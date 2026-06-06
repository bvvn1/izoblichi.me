<script lang="ts">
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';
	import type { AnomalyType } from '$lib/api';

	let { data }: { data: PageData } = $props();

	const formatter = new Intl.NumberFormat('bg-BG', {
		notation: 'compact',
		maximumFractionDigits: 1
	});
	const bgnFormatter = new Intl.NumberFormat('bg-BG', {
		style: 'currency',
		currency: 'BGN',
		minimumFractionDigits: 0,
		maximumFractionDigits: 0
	});
	const pctFormatter = new Intl.NumberFormat('bg-BG', {
		style: 'percent',
		minimumFractionDigits: 1,
		maximumFractionDigits: 1
	});

	const fmt = (n: number | null) => (n == null ? '—' : formatter.format(n));
	const fmtBgn = (n: number | null) => (n == null ? '—' : bgnFormatter.format(n));
	const fmtPct = (n: number | null) => {
		if (n == null) return '—';
		return pctFormatter.format(n / 100);
	};
	const fmtDate = (d: string | null) => {
		if (!d) return '—';
		const date = new Date(d);
		return date.toLocaleDateString('bg-BG', { year: 'numeric', month: 'short', day: 'numeric' });
	};

	const filters = $derived(data.filters);
	const type = $derived<AnomalyType>(filters.type || 'near_threshold');
	let yearFrom = $state(filters.year_from?.toString() || '');
	let yearTo = $state(filters.year_to?.toString() || '');
	let page = $state(filters.page || 1);

	function sectionFor(t: AnomalyType) {
		switch (t) {
			case 'near_threshold': return data.result.near_threshold;
			case 'no_bid': return data.result.no_bid;
			case 'dominance': return data.result.dominance;
			case 'repeated_award': return data.result.repeated_award;
		}
	}

	const section = $derived(sectionFor(type));
	const totalPages = $derived(Math.ceil(section.total / section.per_page));
	const pages = $derived.by(() => {
		const p: number[] = [];
		const maxVisible = 7;
		if (totalPages <= maxVisible) {
			for (let i = 1; i <= totalPages; i++) p.push(i);
		} else {
			p.push(1);
			let start = Math.max(2, page - 2);
			let end = Math.min(totalPages - 1, page + 2);
			if (page <= 3) end = 5;
			if (page >= totalPages - 2) start = totalPages - 4;
			if (start > 2) p.push(-1);
			for (let i = start; i <= end; i++) p.push(i);
			if (end < totalPages - 1) p.push(-1);
			p.push(totalPages);
		}
		return p;
	});

	const tabs: { key: AnomalyType; label: string; desc: string }[] = [
		{
			key: 'near_threshold',
			label: 'Близо до праг',
			desc: 'Поръчки оценени точно под законовия праг за открита процедура'
		},
		{
			key: 'no_bid',
			label: 'Без конкуренция',
			desc: 'Поръчки с 0 или 1 оферта — без реална конкуренция'
		},
		{
			key: 'dominance',
			label: 'Доминация',
			desc: 'Фирми които печелят над 80% от поръчките на даден възложител'
		},
		{
			key: 'repeated_award',
			label: 'Повтарящи се',
			desc: 'Двойки възложител-изпълнител с повтарящи се поръчки без конкуренция'
		}
	];

	function switchTab(t: AnomalyType) {
		page = 1;
		nav(t);
	}

	function updateFilter() {
		page = 1;
		nav(type);
	}

	function nav(t?: AnomalyType) {
		const params = new URLSearchParams();
		const activeType = t || type;
		if (activeType !== 'near_threshold') params.set('type', activeType);
		if (yearFrom) params.set('year_from', yearFrom);
		if (yearTo) params.set('year_to', yearTo);
		if (page > 1) params.set('page', String(page));

		const qs = params.toString();
		goto(resolve('/anomalies') + (qs ? '?' + qs : ''), { replaceState: true, invalidateAll: true });
	}
</script>

<svelte:head>
	<title>Аномалии — изобличи.ме</title>
</svelte:head>

<section class="border-b border-base-content/15 py-10">
	<p class="mb-3 font-mono text-xs tracking-[.14em] text-base-content/50 uppercase">Анализ</p>
	<h1
		class="font-display mb-2 text-3xl leading-tight font-bold"
		style="font-family:'Playfair Display',serif"
	>
		Аномалии
	</h1>
	<p class="max-w-lg text-sm text-base-content/60">
		Автоматично открити модели които заслужават внимание — оценка близо до праг, липса на
		конкуренция, доминация и повтарящи се възлагания.
	</p>
</section>

<!-- Tabs -->
<div class="border-b border-base-content/15">
	<div class="flex overflow-x-auto" role="tablist">
		{#each tabs as tab}
			<button
				role="tab"
				aria-selected={type === tab.key}
				onclick={() => switchTab(tab.key)}
				class="whitespace-nowrap border-b-2 px-4 py-3 font-mono text-xs transition-colors {type === tab.key
					? 'border-[#B85C38] text-[#B85C38]'
					: 'border-transparent text-base-content/50 hover:text-base-content'}"
			>
				{tab.label}
			</button>
		{/each}
	</div>
	<p class="pb-4 pt-2 text-xs text-base-content/50">
		{tabs.find((t) => t.key === type)?.desc}
	</p>
</div>

<!-- Filters -->
<div class="border-b border-base-content/15 py-5">
	<div class="flex flex-wrap items-end gap-3">
		<div style="min-width: 90px">
			<label for="yearFrom" class="mb-1 block font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Година от</label>
			<input
				id="yearFrom"
				type="number"
				min="2020"
				max="2026"
				bind:value={yearFrom}
				onchange={updateFilter}
				placeholder="2020"
				class="input w-full rounded-sm font-mono text-xs"
			/>
		</div>

		<div style="min-width: 90px">
			<label for="yearTo" class="mb-1 block font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Година до</label>
			<input
				id="yearTo"
				type="number"
				min="2020"
				max="2026"
				bind:value={yearTo}
				onchange={updateFilter}
				placeholder="2026"
				class="input w-full rounded-sm font-mono text-xs"
			/>
		</div>

		<button
			onclick={() => { yearFrom = ''; yearTo = ''; page = 1; nav(type); }}
			class="btn btn-outline btn-sm rounded-sm font-mono text-xs"
		>
			Изчисти
		</button>
	</div>
</div>

<!-- Results summary -->
<div class="flex items-center justify-between py-3">
	<p class="font-mono text-xs text-base-content/40">
		Показани {((section.page - 1) * section.per_page) + 1}–{Math.min(section.page * section.per_page, section.total)} от {fmt(section.total)} резултата
	</p>
</div>

<!-- Near Threshold Table -->
{#if type === 'near_threshold'}
	<div class="overflow-x-auto rounded-sm border border-base-content/15">
		<table class="table table-xs w-full">
			<thead>
				<tr class="border-b border-base-content/15">
					<th class="w-8 font-mono text-[10px] tracking-wider text-base-content/40 uppercase">#</th>
					<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Стойност</th>
					<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Праг</th>
					<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Разлика %</th>
					<th class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Възложител</th>
					<th class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Изпълнител</th>
					<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Дата</th>
					<th class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Категория</th>
				</tr>
			</thead>
			<tbody>
				{#if section.items.length === 0}
					<tr>
						<td colspan="8" class="py-12 text-center text-sm text-base-content/40">
							Няма намерени аномалии.
						</td>
					</tr>
				{:else}
					{#each section.items as item, i}
						{@const rowNum = (section.page - 1) * section.per_page + i + 1}
						<tr class="border-b border-base-content/8 transition-colors hover:bg-base-200/50">
							<td class="font-mono text-xs text-base-content/30">{rowNum}</td>
							<td class="text-right whitespace-nowrap text-xs font-medium">
								{fmtBgn(item.contract_value)}
							</td>
							<td class="text-right whitespace-nowrap font-mono text-xs text-base-content/60">
								{fmtBgn(item.threshold)}
							</td>
							<td class="text-right whitespace-nowrap font-mono text-xs text-[#B85C38]">
								{fmtPct(item.gap_pct)}
							</td>
							<td>
								{#if item.buyer_name}
									<a
										href={resolve(`/buyers/${item.buyer_eik}`)}
										class="line-clamp-2 block max-w-[180px] text-xs leading-snug transition-colors hover:text-[#B85C38]"
									>
										{item.buyer_name}
									</a>
								{:else}
									<span class="text-xs text-base-content/30">—</span>
								{/if}
							</td>
							<td>
								{#if item.supplier_name}
									<a
										href={resolve(`/suppliers/${item.supplier_eik}`)}
										class="line-clamp-2 block max-w-[180px] text-xs leading-snug transition-colors hover:text-[#B85C38]"
									>
										{item.supplier_name}
									</a>
								{:else}
									<span class="text-xs text-base-content/30">—</span>
								{/if}
							</td>
							<td class="text-right whitespace-nowrap font-mono text-xs text-base-content/70">
								{fmtDate(item.contract_date)}
							</td>
							<td>
								<span class="line-clamp-1 block max-w-[120px] text-xs text-base-content/60">
									{item.procurement_category || '—'}
								</span>
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
{/if}

<!-- No Bid Table -->
{#if type === 'no_bid'}
	<div class="overflow-x-auto rounded-sm border border-base-content/15">
		<table class="table table-xs w-full">
			<thead>
				<tr class="border-b border-base-content/15">
					<th class="w-8 font-mono text-[10px] tracking-wider text-base-content/40 uppercase">#</th>
					<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Оферти</th>
					<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Стойност</th>
					<th class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Възложител</th>
					<th class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Изпълнител</th>
					<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Дата</th>
					<th class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Предмет</th>
				</tr>
			</thead>
			<tbody>
				{#if section.items.length === 0}
					<tr>
						<td colspan="7" class="py-12 text-center text-sm text-base-content/40">
							Няма намерени аномалии.
						</td>
					</tr>
				{:else}
					{#each section.items as item, i}
						{@const rowNum = (section.page - 1) * section.per_page + i + 1}
						<tr class="border-b border-base-content/8 transition-colors hover:bg-base-200/50">
							<td class="font-mono text-xs text-base-content/30">{rowNum}</td>
							<td class="text-right font-mono text-xs">
								<span class={item.bid_count != null && item.bid_count <= 1 ? 'text-[#B85C38] font-medium' : ''}>
									{item.bid_count != null ? item.bid_count : '—'}
								</span>
							</td>
							<td class="text-right whitespace-nowrap text-xs font-medium">
								{fmtBgn(item.contract_value)}
							</td>
							<td>
								{#if item.buyer_name}
									<a
										href={resolve(`/buyers/${item.buyer_eik}`)}
										class="line-clamp-2 block max-w-[180px] text-xs leading-snug transition-colors hover:text-[#B85C38]"
									>
										{item.buyer_name}
									</a>
								{:else}
									<span class="text-xs text-base-content/30">—</span>
								{/if}
							</td>
							<td>
								{#if item.supplier_name}
									<a
										href={resolve(`/suppliers/${item.supplier_eik}`)}
										class="line-clamp-2 block max-w-[180px] text-xs leading-snug transition-colors hover:text-[#B85C38]"
									>
										{item.supplier_name}
									</a>
								{:else}
									<span class="text-xs text-base-content/30">—</span>
								{/if}
							</td>
							<td class="text-right whitespace-nowrap font-mono text-xs text-base-content/70">
								{fmtDate(item.contract_date)}
							</td>
							<td>
								<span class="line-clamp-2 block max-w-[200px] text-xs leading-snug">
									{item.title || '—'}
								</span>
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
{/if}

<!-- Dominance Table -->
{#if type === 'dominance'}
	<div class="overflow-x-auto rounded-sm border border-base-content/15">
		<table class="table table-xs w-full">
			<thead>
				<tr class="border-b border-base-content/15">
					<th class="w-8 font-mono text-[10px] tracking-wider text-base-content/40 uppercase">#</th>
					<th class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Възложител</th>
					<th class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Изпълнител</th>
					<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Спечелени</th>
					<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase">% от поръчките</th>
					<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Обща стойност</th>
					<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Общо поръчки</th>
				</tr>
			</thead>
			<tbody>
				{#if section.items.length === 0}
					<tr>
						<td colspan="7" class="py-12 text-center text-sm text-base-content/40">
							Няма намерени аномалии.
						</td>
					</tr>
				{:else}
					{#each section.items as item, i}
						{@const rowNum = (section.page - 1) * section.per_page + i + 1}
						<tr class="border-b border-base-content/8 transition-colors hover:bg-base-200/50">
							<td class="font-mono text-xs text-base-content/30">{rowNum}</td>
							<td>
								{#if item.buyer_name}
									<a
										href={resolve(`/buyers/${item.buyer_eik}`)}
										class="line-clamp-2 block max-w-[200px] text-xs leading-snug transition-colors hover:text-[#B85C38]"
									>
										{item.buyer_name}
									</a>
								{:else}
									<span class="text-xs text-base-content/30">—</span>
								{/if}
							</td>
							<td>
								{#if item.supplier_name}
									<a
										href={resolve(`/suppliers/${item.supplier_eik}`)}
										class="line-clamp-2 block max-w-[200px] text-xs leading-snug transition-colors hover:text-[#B85C38]"
									>
										{item.supplier_name}
									</a>
								{:else}
									<span class="text-xs text-base-content/30">—</span>
								{/if}
							</td>
							<td class="text-right font-mono text-xs font-medium">{formatter.format(item.wins)}</td>
							<td class="text-right font-mono text-xs text-[#B85C38]">
								{fmtPct(item.pct_by_count)}
							</td>
							<td class="text-right whitespace-nowrap text-xs font-medium">
								{fmtBgn(item.total_value)}
							</td>
							<td class="text-right font-mono text-xs text-base-content/60">
								{formatter.format(item.total_buyer_contracts)}
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
{/if}

<!-- Repeated Award Table -->
{#if type === 'repeated_award'}
	<div class="overflow-x-auto rounded-sm border border-base-content/15">
		<table class="table table-xs w-full">
			<thead>
				<tr class="border-b border-base-content/15">
					<th class="w-8 font-mono text-[10px] tracking-wider text-base-content/40 uppercase">#</th>
					<th class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Възложител</th>
					<th class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Изпълнител</th>
					<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Общо поръчки</th>
					<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Активни години</th>
					<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Обща стойност</th>
					<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Години</th>
				</tr>
			</thead>
			<tbody>
				{#if section.items.length === 0}
					<tr>
						<td colspan="7" class="py-12 text-center text-sm text-base-content/40">
							Няма намерени аномалии.
						</td>
					</tr>
				{:else}
					{#each section.items as item, i}
						{@const rowNum = (section.page - 1) * section.per_page + i + 1}
						<tr class="border-b border-base-content/8 transition-colors hover:bg-base-200/50">
							<td class="font-mono text-xs text-base-content/30">{rowNum}</td>
							<td>
								{#if item.buyer_name}
									<a
										href={resolve(`/buyers/${item.buyer_eik}`)}
										class="line-clamp-2 block max-w-[200px] text-xs leading-snug transition-colors hover:text-[#B85C38]"
									>
										{item.buyer_name}
									</a>
								{:else}
									<span class="text-xs text-base-content/30">—</span>
								{/if}
							</td>
							<td>
								{#if item.supplier_name}
									<a
										href={resolve(`/suppliers/${item.supplier_eik}`)}
										class="line-clamp-2 block max-w-[200px] text-xs leading-snug transition-colors hover:text-[#B85C38]"
									>
										{item.supplier_name}
									</a>
								{:else}
									<span class="text-xs text-base-content/30">—</span>
								{/if}
							</td>
							<td class="text-right font-mono text-xs font-medium">{formatter.format(item.total_wins)}</td>
							<td class="text-right font-mono text-xs">{item.years_active}</td>
							<td class="text-right whitespace-nowrap text-xs font-medium">
								{fmtBgn(item.total_value)}
							</td>
							<td class="text-right whitespace-nowrap font-mono text-[10px] text-base-content/50">
								{item.years?.join(', ') || '—'}
							</td>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
{/if}

<!-- Pagination -->
{#if totalPages > 1}
	<div class="flex items-center justify-between py-6">
		<button
			disabled={page <= 1}
			onclick={() => { page--; nav(); }}
			class="btn btn-ghost btn-sm rounded-sm font-mono text-xs disabled:opacity-30"
		>
			← Предишна
		</button>

		<div class="flex items-center gap-1">
			{#each pages as p}
				{#if p === -1}
					<span class="px-1 font-mono text-xs text-base-content/20">…</span>
				{:else if p === page}
					<span class="rounded-sm bg-base-content px-2.5 py-1 font-mono text-xs text-base-100">{p}</span>
				{:else}
					<button
						onclick={() => { page = p; nav(); }}
						class="rounded-sm px-2.5 py-1 font-mono text-xs transition-colors hover:bg-base-200"
					>
						{p}
					</button>
				{/if}
			{/each}
		</div>

		<button
			disabled={page >= totalPages}
			onclick={() => { page++; nav(); }}
			class="btn btn-ghost btn-sm rounded-sm font-mono text-xs disabled:opacity-30"
		>
			Следваща →
		</button>
	</div>
{/if}
