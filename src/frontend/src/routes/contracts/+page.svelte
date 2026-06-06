<script lang="ts">
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';
	import type { SortBy, SortDir } from '$lib/api';

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

	const fmt = (n: number | null) => (n == null ? '—' : formatter.format(n));
	const fmtValue = (n: number | null, currency: string | null) => {
		if (n == null) return '—';
		if (currency === 'BGN') return bgnFormatter.format(n);
		return formatter.format(n) + ' ' + (currency || '');
	};
	const fmtDate = (d: string | null) => {
		if (!d) return '—';
		const date = new Date(d);
		return date.toLocaleDateString('bg-BG', { year: 'numeric', month: 'short', day: 'numeric' });
	};

	const filters = $derived(data.filters);
	let q = $state(data.filters.q || '');
	let yearFrom = $state(filters.year_from?.toString() || '');
	let yearTo = $state(filters.year_to?.toString() || '');
	let category = $state(filters.category || '');
	let source = $state(filters.source || '');
	let sortBy = $state<SortBy>(filters.sort_by || 'contract_date');
	let sortDir = $state<SortDir>(filters.sort_dir || 'desc');
	let page = $state(filters.page || 1);

	let searchTimeout: ReturnType<typeof setTimeout>;

	function updateSearch() {
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => {
			page = 1;
			nav();
		}, 400);
	}

	function updateFilter() {
		page = 1;
		nav();
	}

	function toggleSort(column: SortBy) {
		if (sortBy === column) {
			sortDir = sortDir === 'desc' ? 'asc' : 'desc';
		} else {
			sortBy = column;
			sortDir = 'desc';
		}
		nav();
	}

	function sortIndicator(column: SortBy): string {
		if (sortBy !== column) return '';
		return sortDir === 'desc' ? ' ↓' : ' ↑';
	}

	function nav() {
		const params = new URLSearchParams();
		if (q) params.set('q', q);
		if (yearFrom) params.set('year_from', yearFrom);
		if (yearTo) params.set('year_to', yearTo);
		if (category) params.set('category', category);
		if (source) params.set('source', source);
		if (sortBy !== 'contract_date') params.set('sort_by', sortBy);
		if (sortDir !== 'desc') params.set('sort_dir', sortDir);
		if (page > 1) params.set('page', String(page));

		const qs = params.toString();
		goto(resolve('/contracts') + (qs ? '?' + qs : ''), { replaceState: true, invalidateAll: true });
	}

	const result = $derived(data.result);
	const totalPages = $derived(Math.ceil(result.total / result.per_page));
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
</script>

<svelte:head>
	<title>Договори — изобличи.ме</title>
</svelte:head>

<section class="border-b border-base-content/15 py-10">
	<p class="mb-3 font-mono text-xs tracking-[.14em] text-base-content/50 uppercase">Данни</p>
	<h1
		class="font-display mb-2 text-3xl leading-tight font-bold"
		style="font-family:'Playfair Display',serif"
	>
		Договори
	</h1>
	<p class="max-w-lg text-sm text-base-content/60">
		{fmt(result.total)} договора от 2020–2026 г. Търсене, филтриране и сортиране по всички
		полета.
	</p>
</section>

<!-- Filters -->
<div class="border-b border-base-content/15 py-5">
	<div class="flex flex-wrap items-end gap-3">
		<div class="min-w-0 flex-1" style="flex-basis: 240px">
			<label for="search" class="mb-1 block font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Търсене</label>
			<input
				id="search"
				type="text"
				bind:value={q}
				oninput={updateSearch}
				placeholder="Търси по заглавие, купувач или доставчик…"
				class="input w-full rounded-sm font-mono text-xs"
			/>
		</div>

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

		<div style="min-width: 120px">
			<label for="source" class="mb-1 block font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Източник</label>
			<select id="source" bind:value={source} onchange={updateFilter} class="select w-full rounded-sm font-mono text-xs">
				<option value="">Всички</option>
				<option value="legacy">Legacy (2020-23)</option>
				<option value="ocds">OCDS (2026+)</option>
			</select>
		</div>

		<button
			onclick={() => { q = ''; yearFrom = ''; yearTo = ''; category = ''; source = ''; sortBy = 'contract_date'; sortDir = 'desc'; page = 1; nav(); }}
			class="btn btn-outline btn-sm rounded-sm font-mono text-xs"
		>
			Изчисти
		</button>
	</div>
</div>

<!-- Results summary -->
<div class="flex items-center justify-between py-3">
	<p class="font-mono text-xs text-base-content/40">
		Показани {((result.page - 1) * result.per_page) + 1}–{Math.min(result.page * result.per_page, result.total)} от {fmt(result.total)} резултата
	</p>
</div>

<!-- Table -->
<div class="overflow-x-auto rounded-sm border border-base-content/15">
	<table class="table table-xs w-full">
		<thead>
			<tr class="border-b border-base-content/15">
				<th class="w-8 font-mono text-[10px] tracking-wider text-base-content/40 uppercase">#</th>
				<th class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Заглавие / Предмет</th>
				<th class="cursor-pointer font-mono text-[10px] tracking-wider text-base-content/40 uppercase select-none" onclick={() => toggleSort('buyer_name')}>
					Възложител{sortIndicator('buyer_name')}
				</th>
				<th class="cursor-pointer font-mono text-[10px] tracking-wider text-base-content/40 uppercase select-none" onclick={() => toggleSort('supplier_name')}>
					Изпълнител{sortIndicator('supplier_name')}
				</th>
				<th class="cursor-pointer text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase select-none" onclick={() => toggleSort('contract_value')}>
					Стойност{sortIndicator('contract_value')}
				</th>
				<th class="cursor-pointer text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase select-none" onclick={() => toggleSort('contract_date')}>
					Дата{sortIndicator('contract_date')}
				</th>
				<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Оферти</th>
				<th class="text-center font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Източник</th>
			</tr>
		</thead>
		<tbody>
			{#if result.items.length === 0}
				<tr>
					<td colspan="8" class="py-12 text-center text-sm text-base-content/40">
						Няма намерени договори.
					</td>
				</tr>
			{:else}
				{#each result.items as contract, i}
					{@const rowNum = (result.page - 1) * result.per_page + i + 1}
					<tr class="border-b border-base-content/8 transition-colors hover:bg-base-200/50">
						<td class="font-mono text-xs text-base-content/30">{rowNum}</td>
						<td>
							<a
								href={resolve(`/contracts/${contract.data_source}/${contract.row_key}`)}
								class="line-clamp-2 block max-w-xs text-xs leading-snug font-medium transition-colors hover:text-[#B85C38]"
							>
								{contract.title || contract.procurement_number || '—'}
							</a>
							{#if contract.procurement_category}
								<span class="mt-0.5 inline-block rounded-sm bg-base-200 px-1.5 font-mono text-[10px] text-base-content/40">
									{contract.procurement_category}
								</span>
							{/if}
						</td>
						<td>
							{#if contract.buyer_name}
								<a
									href={resolve(`/buyers/${contract.buyer_eik}`)}
									class="line-clamp-2 block max-w-[180px] text-xs leading-snug transition-colors hover:text-[#B85C38]"
								>
									{contract.buyer_name}
								</a>
							{:else}
								<span class="text-xs text-base-content/30">—</span>
							{/if}
						</td>
						<td>
							{#if contract.supplier_name}
								<a
									href={resolve(`/suppliers/${contract.supplier_eik}`)}
									class="line-clamp-2 block max-w-[180px] text-xs leading-snug transition-colors hover:text-[#B85C38]"
								>
									{contract.supplier_name}
								</a>
							{:else}
								<span class="text-xs text-base-content/30">—</span>
							{/if}
						</td>
						<td class="text-right">
							<span class="whitespace-nowrap text-xs font-medium">
								{fmtValue(contract.contract_value, contract.currency)}
							</span>
							{#if contract.eu_funded}
								<span class="ml-1 rounded-sm bg-[#003399]/10 px-1 font-mono text-[9px] text-[#003399]" title="ЕС финансиране">ЕС</span>
							{/if}
						</td>
						<td class="text-right whitespace-nowrap font-mono text-xs text-base-content/70">
							{fmtDate(contract.contract_date)}
						</td>
						<td class="text-right font-mono text-xs">
							{contract.bid_count != null ? contract.bid_count : '—'}
						</td>
						<td class="text-center">
							<span
								class="inline-block rounded-sm px-1.5 font-mono text-[9px] font-medium uppercase {contract.data_source === 'ocds' ? 'bg-[#4A7C59]/10 text-[#4A7C59]' : 'bg-base-200 text-base-content/50'}"
							>
								{contract.data_source}
							</span>
						</td>
					</tr>
				{/each}
			{/if}
		</tbody>
	</table>
</div>

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
