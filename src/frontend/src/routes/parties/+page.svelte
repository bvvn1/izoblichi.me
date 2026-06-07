<script lang="ts">
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';
	import { format, formatDate } from '$lib/formatting';
	import { SvelteURLSearchParams } from 'svelte/reactivity';

	let { data }: { data: PageData } = $props();

	let filters = $state(data.filters);

	let q = $derived(filters.q);
	let page = $derived(filters.page);

	let searchTimeout: ReturnType<typeof setTimeout>;

	function updateSearch() {
		clearTimeout(searchTimeout);
		searchTimeout = setTimeout(() => {
			filters.page = 1;
			nav();
		}, 400);
	}

	function nav() {
		const params = new SvelteURLSearchParams();
		if (q) params.set('q', q);
		if (page > 1) params.set('page', page.toString());

		goto(resolve(`/parties?${params}`), {
			keepFocus: true,
			invalidateAll: true
		});
	}

	const parties = $derived(data.response);
	const totalPages = $derived(Math.ceil(parties.total / parties.per_page));
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
	<title>Юридически лица — изобличи.ме</title>
</svelte:head>

<nav
	class="flex items-center gap-2 border-b border-base-content/15 py-3 font-mono text-xs text-base-content/40"
>
	<a href={resolve('/')} class="hover:text-base-content">Начало</a>
	<span>/</span>
	<a href={resolve('/parties')} class="hover:text-base-content">Юридически лица</a>
</nav>

<section class="border-b border-base-content/15 py-10">
	<p class="mb-3 font-mono text-xs tracking-[.14em] text-base-content/50 uppercase">Данни</p>
	<h1
		class="font-display mb-2 text-3xl leading-tight font-bold"
		style="font-family:'Playfair Display',serif"
	>
		Юридически лица
	</h1>
	<p class="max-w-lg text-sm text-base-content/60">
		{format(parties.total)} организации — всички възложители и изпълнители в системата.
	</p>
</section>

<!-- Search -->
<div class="border-b border-base-content/15 py-5">
	<div class="flex flex-wrap items-end gap-3">
		<div class="min-w-0 flex-1" style="flex-basis: 320px">
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
				placeholder="Търси по име или ЕИК…"
				class="input w-full rounded-sm font-mono text-xs"
			/>
		</div>

		<button
			onclick={() => {
				filters.q = undefined;
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
		Показани {(parties.page - 1) * parties.per_page + 1}–{Math.min(
			parties.page * parties.per_page,
			parties.total
		)} от {format(parties.total)} резултата
	</p>
</div>

<!-- Table -->
<div class="overflow-x-auto rounded-sm border border-base-content/15">
	<table class="table w-full table-xs">
		<thead>
			<tr class="border-b border-base-content/15">
				<th class="w-8 font-mono text-[10px] tracking-wider text-base-content/40 uppercase">#</th>
				<th class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Име</th>
				<th class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">ЕИК</th>
				<th class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase"
					>Населено място</th
				>
				<th class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">Област</th>
				<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase"
					>Първа поява</th
				>
				<th class="text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase"
					>Последна поява</th
				>
			</tr>
		</thead>
		<tbody>
			{#if parties.items.length === 0}
				<tr>
					<td colspan="7" class="py-12 text-center text-sm text-base-content/40">
						Няма намерени юридически лица.
					</td>
				</tr>
			{:else}
				{#each parties.items as party, i (i)}
					{@const rowNum = (parties.page - 1) * parties.per_page + i + 1}
					<tr class="border-b border-base-content/8 transition-colors hover:bg-base-200/50">
						<td class="font-mono text-xs text-base-content/30">{rowNum}</td>
						<td>
							<a
								href={resolve(`/parties/${party.eik}`)}
								class="line-clamp-2 block max-w-60 text-xs leading-snug font-medium transition-colors hover:text-[#B85C38]"
							>
								{party.display_name || party.legal_name || party.eik}
							</a>
						</td>
						<td>
							<span class="font-mono text-xs text-base-content/60">{party.eik}</span>
						</td>
						<td>
							<span class="line-clamp-1 block max-w-60 text-xs">
								{party.address_locality || '—'}
							</span>
						</td>
						<td>
							<span class="line-clamp-1 block max-w-60 text-xs">
								{party.address_region || '—'}
							</span>
						</td>
						<td class="text-right font-mono text-xs whitespace-nowrap text-base-content/70">
							{formatDate(party.first_seen)}
						</td>
						<td class="text-right font-mono text-xs whitespace-nowrap text-base-content/70">
							{formatDate(party.last_seen)}
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
