<script lang="ts">
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';
	import type { SortBy } from '$lib/api';
	import { format, formatDate, formatCurrency } from '$lib/formatting';
	import { SvelteURLSearchParams } from 'svelte/reactivity';

	let { data }: { data: PageData } = $props();

	let q = $derived(data.filters.q);
	let yearFrom = $derived(data.filters.year_from?.toString() ?? '');
	let yearTo = $derived(data.filters.year_to?.toString() ?? '');
	let category = $derived(data.filters.category);
	let source = $derived(data.filters.source);
	let sortBy = $derived(data.filters.sort_by);
	let sortDir = $derived(data.filters.sort_dir);
	let page = $derived(data.filters.page);

	let searchTimeout: ReturnType<typeof setTimeout>;

	function updateSearch() {
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => nav({ q, page: 1 }), 650);
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

	function nav(overrides: Partial<typeof data.filters> = {}) {
		const merged = {
			...data.filters,
			q,
			yearFrom,
			yearTo,
			category,
			source,
			sortBy,
			sortDir,
			page,
			...overrides
		};
		const params = new SvelteURLSearchParams();

		if (merged.q) params.set('q', merged.q);
		if (merged.year_from) params.set('year_from', String(merged.year_from));
		if (merged.year_to) params.set('year_to', String(merged.year_to));
		if (merged.category) params.set('category', merged.category);
		if (merged.source) params.set('source', merged.source);
		if (merged.sort_by) params.set('sort_by', merged.sort_by);
		if (merged.sort_dir) params.set('sort_dir', merged.sort_dir);
		if (merged.page > 1) params.set('page', String(merged.page));

		goto(resolve(`/contracts?${params}`), {
			keepFocus: true,
			replaceState: true,
			invalidateAll: true
		});
	}

	const contracts = $derived(data.contracts);
	const totalPages = $derived(Math.ceil(contracts.total / contracts.per_page));
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
		{format(contracts.total)} договора от 2020–2026 г. Търсене, филтриране и сортиране по всички полета.
	</p>
</section>

<!-- Filters -->
<div class="border-b border-base-content/15 py-5">
	<div class="flex flex-wrap items-end gap-3">
		<div class="min-w-0 flex-1" style="flex-basis: 240px">
			<label
				for="search"
				class="mb-1 block font-mono text-[10px] tracking-wider text-base-content/40 uppercase"
				>Търсене</label
			>
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
			<label
				for="yearFrom"
				class="mb-1 block font-mono text-[10px] tracking-wider text-base-content/40 uppercase"
				>Година от</label
			>
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
			<label
				for="yearTo"
				class="mb-1 block font-mono text-[10px] tracking-wider text-base-content/40 uppercase"
				>Година до</label
			>
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
			<label
				for="source"
				class="mb-1 block font-mono text-[10px] tracking-wider text-base-content/40 uppercase"
				>Източник</label
			>
			<select
				id="source"
				bind:value={source}
				onchange={updateFilter}
				class="select w-full rounded-sm font-mono text-xs"
			>
				<option value="">Всички</option>
				<option value="legacy">Legacy (2020-23)</option>
				<option value="ocds">OCDS (2026+)</option>
			</select>
		</div>

		<button
			onclick={() => {
				q = '';
				yearFrom = '';
				yearTo = '';
				category = '';
				source = undefined;
				sortBy = 'contract_date';
				sortDir = 'desc';
				page = 1;
				nav();
			}}
			class="btn rounded-sm font-mono text-xs btn-outline btn-sm"
		>
			Изчисти
		</button>
	</div>
</div>

<!-- Results summary -->
<div class="flex items-center justify-between py-3">
	<p class="font-mono text-xs text-base-content/40">
		Показани {(contracts.page - 1) * contracts.per_page + 1}–{Math.min(
			contracts.page * contracts.per_page,
			contracts.total
		)} от {format(contracts.total)} резултата
	</p>
</div>

<!-- Table -->
<div class="overflow-x-auto rounded-sm border border-base-content/15">
	<table class="table w-full table-xs">
		<thead>
			<tr class="border-b border-base-content/15">
				<th class="w-8 font-mono text-[10px] tracking-wider text-base-content/40 uppercase">#</th>
				<th class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase"
					>Заглавие / Предмет</th
				>
				<th
					class="cursor-pointer font-mono text-[10px] tracking-wider text-base-content/40 uppercase select-none"
					onclick={() => toggleSort('buyer_name')}
				>
					Възложител{sortIndicator('buyer_name')}
				</th>
				<th
					class="cursor-pointer font-mono text-[10px] tracking-wider text-base-content/40 uppercase select-none"
					onclick={() => toggleSort('supplier_name')}
				>
					Изпълнител{sortIndicator('supplier_name')}
				</th>
				<th
					class="cursor-pointer text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase select-none"
					onclick={() => toggleSort('contract_value')}
				>
					Стойност{sortIndicator('contract_value')}
				</th>
				<th
					class="cursor-pointer text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase select-none"
					onclick={() => toggleSort('contract_date')}
				>
					Дата{sortIndicator('contract_date')}
				</th>
				<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase"
					>Оферти</th
				>
				<th class="text-center font-mono text-[10px] tracking-wider text-base-content/40 uppercase"
					>Източник</th
				>
			</tr>
		</thead>
		<tbody>
			{#if contracts.items.length === 0}
				<tr>
					<td colspan="8" class="py-12 text-center text-sm text-base-content/40">
						Няма намерени договори.
					</td>
				</tr>
			{:else}
				{#each contracts.items as contract, i (contract.row_key)}
					{@const rowNum = (contracts.page - 1) * contracts.per_page + i + 1}
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
								<span
									class="mt-0.5 inline-block rounded-sm bg-base-200 px-1.5 font-mono text-[10px] text-base-content/40"
								>
									{contract.procurement_category}
								</span>
							{/if}
						</td>
						<td>
							{#if contract.buyer_name}
								<a
									href={resolve(`/buyers/${contract.buyer_eik}`)}
									class="line-clamp-2 block max-w-45 text-xs leading-snug transition-colors hover:text-[#B85C38]"
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
									class="line-clamp-2 block max-w-45 text-xs leading-snug transition-colors hover:text-[#B85C38]"
								>
									{contract.supplier_name}
								</a>
							{:else}
								<span class="text-xs text-base-content/30">—</span>
							{/if}
						</td>
						<td class="text-right">
							<span class="text-xs font-medium whitespace-nowrap">
								{formatCurrency(contract.contract_value, contract.currency)}
							</span>
							{#if contract.eu_funded}
								<span
									class="ml-1 rounded-sm bg-[#003399]/10 px-1 font-mono text-[9px] text-[#003399]"
									title="ЕС финансиране">ЕС</span
								>
							{/if}
						</td>
						<td class="text-right font-mono text-xs whitespace-nowrap text-base-content/70">
							{formatDate(contract.contract_date)}
						</td>
						<td class="text-right font-mono text-xs">
							{contract.bid_count != null ? contract.bid_count : '—'}
						</td>
						<td class="text-center">
							<span
								class="inline-block rounded-sm px-1.5 font-mono text-[9px] font-medium uppercase {contract.data_source ===
								'ocds'
									? 'bg-[#4A7C59]/10 text-[#4A7C59]'
									: 'bg-base-200 text-base-content/50'}"
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
			onclick={() => {
				page--;
				nav();
			}}
			class="btn rounded-sm font-mono text-xs btn-ghost btn-sm disabled:opacity-30"
		>
			← Предишна
		</button>

		<div class="flex items-center gap-1">
			{#each pages as page_index (page_index)}
				{#if page_index === -1}
					<span class="px-1 font-mono text-xs text-base-content/20">…</span>
				{:else if page_index === page}
					<span class="rounded-sm bg-base-content px-2.5 py-1 font-mono text-xs text-base-100"
						>{page_index}</span
					>
				{:else}
					<button
						onclick={() => {
							page = page_index;
							nav();
						}}
						class="rounded-sm px-2.5 py-1 font-mono text-xs transition-colors hover:bg-base-200"
					>
						{page_index}
					</button>
				{/if}
			{/each}
		</div>

		<button
			disabled={page >= totalPages}
			onclick={() => {
				page++;
				nav();
			}}
			class="btn rounded-sm font-mono text-xs btn-ghost btn-sm disabled:opacity-30"
		>
			Следваща →
		</button>
	</div>
{/if}
