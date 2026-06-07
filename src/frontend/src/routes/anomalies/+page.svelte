<script lang="ts">
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';
	import type { AnomalyType } from '$lib/api';
	import { format, formatPercent, formatCurrency, formatDate } from '$lib/formatting';
	import { SvelteURLSearchParams } from 'svelte/reactivity';
	import type {
		NearThresholdAnomaly,
		NoBidAnomaly,
		DominanceAnomaly,
		RepeatedAwardAnomaly
	} from '$lib/types';

	let { data }: { data: PageData } = $props();

	let filters = $state(data.filters);

	const type = $derived<AnomalyType>(filters.type ?? 'near_threshold');
	const page = $derived(filters.page ?? 1);
	const section = $derived(data.anomalies[type]);

	const totalPages = $derived(Math.ceil(section.total / section.per_page));
	const pages = $derived.by(() => {
		if (totalPages <= 7) return Array.from({ length: totalPages }, (_, i) => i + 1);
		const p: number[] = [1];
		let start = Math.max(2, page - 2);
		let end = Math.min(totalPages - 1, page + 2);
		if (page <= 3) end = 5;
		if (page >= totalPages - 2) start = totalPages - 4;
		if (start > 2) p.push(-1);
		for (let i = start; i <= end; i++) p.push(i);
		if (end < totalPages - 1) p.push(-1);
		p.push(totalPages);
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
			desc: 'Двойки "възложител-изпълнител" с повтарящи се поръчки без конкуренция'
		}
	];

	function nav(overrides: { type?: AnomalyType; page?: number } = {}) {
		const params = new SvelteURLSearchParams();

		const activeType = overrides.type ?? type;
		const activePage = overrides.page ?? page;
		if (activeType !== 'near_threshold') params.set('type', activeType);
		if (filters.year_from) params.set('year_from', filters.year_from.toString());
		if (filters.year_to) params.set('year_to', filters.year_to.toString());
		if (activePage > 1) params.set('page', activePage.toString());
		if (filters.buyer_eik) params.set('buyer_eik', filters.buyer_eik);
		if (filters.supplier_eik) params.set('supplier_eik', filters.supplier_eik);

		goto(resolve(`/anomalies?${params}`), {
			keepFocus: true,
			invalidateAll: true
		});
	}
</script>

<svelte:head>
	<title>Аномалии — изобличи.ме</title>
</svelte:head>

{#snippet partyLink(name: string | null, eik: string | null, role: 'buyers' | 'suppliers')}
	{#if name}
		<a
			href={resolve(`/${role}/${eik}`)}
			class="line-clamp-2 block max-w-45 text-xs leading-snug transition-colors hover:text-[#B85C38]"
		>
			{name}
		</a>
	{:else}
		<span class="text-xs text-base-content/30">—</span>
	{/if}
{/snippet}

{#snippet emptyRow(colspan: number)}
	<tr>
		<td {colspan} class="py-12 text-center text-sm text-base-content/40">
			Няма намерени аномалии.
		</td>
	</tr>
{/snippet}

<nav
	class="flex items-center gap-2 border-b border-base-content/15 py-3 font-mono text-xs text-base-content/40"
>
	<a href={resolve('/')} class="hover:text-base-content">Начало</a>
	<span>/</span>
	<a href={resolve('/anomalies')} class="hover:text-base-content">Аномалии</a>
</nav>

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

<div class="border-b border-base-content/15">
	<div class="flex overflow-x-auto" role="tablist">
		{#each tabs as tab (tab.key)}
			<button
				role="tab"
				aria-selected={type === tab.key}
				onclick={() => nav({ type: tab.key, page: 1 })}
				class="border-b-2 px-4 py-3 font-mono text-xs whitespace-nowrap transition-colors
					{type === tab.key
					? 'border-[#B85C38] text-[#B85C38]'
					: 'border-transparent text-base-content/50 hover:text-base-content'}"
			>
				{tab.label}
			</button>
		{/each}
	</div>
	<p class="pt-2 pb-4 text-xs text-base-content/50">
		{tabs.find((t) => t.key === type)?.desc}
	</p>
</div>

<!-- Filters -->
<div class="border-b border-base-content/15 py-5">
	<div class="flex flex-wrap items-end gap-3">
		{#each [{ id: 'yearFrom', label: 'Година от', key: 'year_from', placeholder: '2020' }, { id: 'yearTo', label: 'Година до', key: 'year_to', placeholder: '2026' }] as filter (filter.id)}
			<div style="min-width:90px">
				<label
					for={filter.id}
					class="mb-1 block font-mono text-[10px] tracking-wider text-base-content/40 uppercase"
				>
					{filter.label}
				</label>
				<input
					id={filter.id}
					type="number"
					min="2020"
					max="2026"
					placeholder={filter.placeholder}
					bind:value={filters[filter.key]}
					onchange={() => nav({ page: 1 })}
					class="input w-full rounded-sm font-mono text-xs"
				/>
			</div>
		{/each}
		<button
			onclick={() => {
				filters.year_from = undefined;
				filters.year_to = undefined;
				nav({ page: 1 });
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
		Показани {(section.page - 1) * section.per_page + 1}–{Math.min(
			section.page * section.per_page,
			section.total
		)}
		от {format(section.total)} резултата
	</p>
</div>

{#if type === 'near_threshold'}
	<div class="overflow-x-auto rounded-sm border border-base-content/15">
		<table class="table w-full table-xs">
			<thead>
				<tr class="border-b border-base-content/15">
					{#each ['#', 'Стойност', 'Праг', 'Разлика %', 'Възложител', 'Изпълнител', 'Дата', 'Категория'] as heading, i (heading)}
						<th
							class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase {i > 0 &&
							i < 4
								? 'text-right'
								: ''} {i === 0 ? 'w-8' : ''}"
						>
							{heading}
						</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				{#if section.items.length === 0}
					{@render emptyRow(8)}
				{:else}
					{#each section.items as NearThresholdAnomaly[] as item, i (item)}
						{@const rowNum = (section.page - 1) * section.per_page + i + 1}
						<tr class="border-b border-base-content/8 transition-colors hover:bg-base-200/50">
							<td class="font-mono text-xs text-base-content/30">{rowNum}</td>
							<td class="text-right text-xs font-medium whitespace-nowrap"
								>{formatCurrency(item.contract_value, 'BGN')}</td
							>
							<td class="text-right font-mono text-xs whitespace-nowrap text-base-content/60"
								>{formatCurrency(item.threshold, 'BGN')}</td
							>
							<td class="text-right font-mono text-xs whitespace-nowrap text-[#B85C38]"
								>{formatPercent(item.gap_pct)}</td
							>
							<td>{@render partyLink(item.buyer_name, item.buyer_eik, 'buyers')}</td>
							<td>{@render partyLink(item.supplier_name, item.supplier_eik, 'suppliers')}</td>
							<td class="font-mono text-xs text-base-content/70"
								>{formatDate(item.contract_date)}</td
							>
							<td
								><span class="line-clamp-1 block max-w-30 text-xs text-base-content/60"
									>{item.procurement_category || '—'}</span
								></td
							>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
{/if}

{#if type === 'no_bid'}
	<div class="overflow-x-auto rounded-sm border border-base-content/15">
		<table class="table w-full table-xs">
			<thead>
				<tr class="border-b border-base-content/15">
					{#each ['#', 'Оферти', 'Стойност', 'Възложител', 'Изпълнител', 'Дата', 'Предмет'] as heading, i (heading)}
						<th
							class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase {i > 0 &&
							i < 3
								? 'text-right'
								: ''} {i === 0 ? 'w-8' : ''}"
						>
							{heading}
						</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				{#if section.items.length === 0}
					{@render emptyRow(7)}
				{:else}
					{#each section.items as NoBidAnomaly[] as item, i (item)}
						{@const rowNum = (section.page - 1) * section.per_page + i + 1}
						<tr class="border-b border-base-content/8 transition-colors hover:bg-base-200/50">
							<td class="font-mono text-xs text-base-content/30">{rowNum}</td>
							<td class="text-right font-mono text-xs">
								<span
									class={item.bid_count != null && item.bid_count <= 1
										? 'font-medium text-[#B85C38]'
										: ''}
								>
									{item.bid_count ?? '—'}
								</span>
							</td>
							<td class="text-right text-xs font-medium whitespace-nowrap"
								>{formatCurrency(item.contract_value, item.currency)}</td
							>
							<td>{@render partyLink(item.buyer_name, item.buyer_eik, 'buyers')}</td>
							<td>{@render partyLink(item.supplier_name, item.supplier_eik, 'suppliers')}</td>
							<td class="font-mono text-xs text-base-content/70"
								>{formatDate(item.contract_date)}</td
							>
							<td
								><span class="line-clamp-2 block max-w-50 text-xs leading-snug"
									>{item.title || '—'}</span
								></td
							>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
{/if}

{#if type === 'dominance'}
	<div class="overflow-x-auto rounded-sm border border-base-content/15">
		<table class="table w-full table-xs">
			<thead>
				<tr class="border-b border-base-content/15">
					{#each ['#', 'Възложител', 'Изпълнител', 'Спечелени', '% от поръчките', 'Обща стойност', 'Общо поръчки'] as heading, i (heading)}
						<th
							class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase {i >= 3
								? 'text-right'
								: ''} {i === 0 ? 'w-8' : ''}"
						>
							{heading}
						</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				{#if section.items.length === 0}
					{@render emptyRow(7)}
				{:else}
					{#each section.items as DominanceAnomaly[] as item, i (item)}
						{@const rowNum = (section.page - 1) * section.per_page + i + 1}
						<tr class="border-b border-base-content/8 transition-colors hover:bg-base-200/50">
							<td class="font-mono text-xs text-base-content/30">{rowNum}</td>
							<td>{@render partyLink(item.buyer_name, item.buyer_eik, 'buyers')}</td>
							<td>{@render partyLink(item.supplier_name, item.supplier_eik, 'suppliers')}</td>
							<td class="text-right font-mono text-xs font-medium">{format(item.wins)}</td>
							<td class="text-right font-mono text-xs text-[#B85C38]"
								>{formatPercent(item.pct_by_count)}</td
							>
							<td class="text-right text-xs font-medium whitespace-nowrap"
								>{formatCurrency(item.total_value, 'BGN')}</td
							>
							<td class="text-right font-mono text-xs text-base-content/60"
								>{format(item.total_buyer_contracts)}</td
							>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
{/if}

{#if type === 'repeated_award'}
	<div class="overflow-x-auto rounded-sm border border-base-content/15">
		<table class="table w-full table-xs">
			<thead>
				<tr class="border-b border-base-content/15">
					{#each ['#', 'Възложител', 'Изпълнител', 'Общо поръчки', 'Активни години', 'Обща стойност', 'Години'] as heading, i (heading)}
						<th
							class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase {i >= 3
								? 'text-right'
								: ''} {i === 0 ? 'w-8' : ''}"
						>
							{heading}
						</th>
					{/each}
				</tr>
			</thead>
			<tbody>
				{#if section.items.length === 0}
					{@render emptyRow(7)}
				{:else}
					{#each section.items as RepeatedAwardAnomaly[] as item, i (item)}
						{@const rowNum = (section.page - 1) * section.per_page + i + 1}
						<tr class="border-b border-base-content/8 transition-colors hover:bg-base-200/50">
							<td class="font-mono text-xs text-base-content/30">{rowNum}</td>
							<td>{@render partyLink(item.buyer_name, item.buyer_eik, 'buyers')}</td>
							<td>{@render partyLink(item.supplier_name, item.supplier_eik, 'suppliers')}</td>
							<td class="text-right font-mono text-xs font-medium">{format(item.total_wins)}</td>
							<td class="text-right font-mono text-xs">{item.years_active}</td>
							<td class="text-right text-xs font-medium whitespace-nowrap"
								>{formatCurrency(item.total_value, 'BGN')}</td
							>
							<td class="text-right font-mono text-[10px] whitespace-nowrap text-base-content/50"
								>{item.years?.join(', ') || '—'}</td
							>
						</tr>
					{/each}
				{/if}
			</tbody>
		</table>
	</div>
{/if}

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
