<script lang="ts">
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';
	import type { SortBy } from '$lib/api';
	import { format, formatDate, formatCurrency } from '$lib/formatting';
	import { SvelteURLSearchParams } from 'svelte/reactivity';

	let { data }: { data: PageData } = $props();

	let filters = $state(data.filters);

	let q = $derived(filters.q);
	let sortBy = $derived(filters.sort_by);
	let sortDir = $derived(filters.sort_dir);
	let page = $derived(filters.page);

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
			filters.sort_dir = sortDir === 'desc' ? 'asc' : 'desc';
		} else {
			filters.sort_by = column;
			filters.sort_dir = 'desc';
		}
		nav();
	}

	function sortIndicator(column: SortBy): string {
		if (sortBy !== column) return '';
		return sortDir === 'desc' ? ' ↓' : ' ↑';
	}

	function nav(overrides: Partial<typeof filters> = {}) {
		const merged = {
			...filters,
			...overrides
		};
		const params = new SvelteURLSearchParams();

		if (merged.q) params.set('q', merged.q);
		if (merged.year_from) params.set('year_from', merged.year_from.toString());
		if (merged.year_to) params.set('year_to', merged.year_to.toString());
		if (merged.category) params.set('category', merged.category);
		if (merged.source) params.set('source', merged.source);
		if (merged.sort_by) params.set('sort_by', merged.sort_by);
		if (merged.sort_dir) params.set('sort_dir', merged.sort_dir);
		if (merged.buyer_eik) params.set('buyer_eik', merged.buyer_eik);
		if (merged.supplier_eik) params.set('supplier_eik', merged.supplier_eik);
		if (merged.page > 1) params.set('page', merged.page.toString());

		goto(resolve(`/contracts?${params}`), {
			keepFocus: true,
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

<nav
	class="flex items-center gap-2 border-b border-base-content/15 py-3 font-mono text-xs text-base-content/40"
>
	<a href={resolve('/')} class="hover:text-base-content">Начало</a>
	<span>/</span>
	<a href={resolve('/contracts')} class="hover:text-base-content">Договори</a>
	<span>/</span>
</nav>

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
				bind:value={filters.q}
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
				bind:value={filters.year_from}
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
				bind:value={filters.year_to}
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
				bind:value={filters.source}
				onchange={updateFilter}
				class="select w-full rounded-sm font-mono text-xs"
			>
				<option value={undefined}>Всички</option>
				<option value="legacy">Legacy (2020-23)</option>
				<option value="ocds">OCDS (2026+)</option>
			</select>
		</div>

		<button
			onclick={() => {
				filters.q = '';
				filters.year_from = undefined;
				filters.year_to = undefined;
				filters.category = '';
				filters.source = undefined;
				filters.sort_by = 'contract_date';
				filters.sort_dir = 'desc';
				filters.buyer_eik = undefined;
				filters.supplier_eik = undefined;
				filters.page = 1;
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
				filters.page--;
				nav();
			}}
			class="btn rounded-sm font-mono text-xs btn-ghost btn-sm disabled:opacity-30"
		>
			← Предишна
		</button>

		<div class="flex items-center gap-1">
			{#each pages as page_index, idx (page_index === -1 ? `ellipsis-${idx}` : page_index)}
				{#if page_index === -1}
					<span class="px-1 font-mono text-xs text-base-content/20">…</span>
				{:else if page_index === page}
					<span class="rounded-sm bg-base-content px-2.5 py-1 font-mono text-xs text-base-100"
						>{page_index}</span
					>
				{:else}
					<button
						onclick={() => {
							filters.page = page_index;
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
				filters.page++;
				nav();
			}}
			class="btn rounded-sm font-mono text-xs btn-ghost btn-sm disabled:opacity-30"
		>
			Следваща →
		</button>
	</div>
{/if}
