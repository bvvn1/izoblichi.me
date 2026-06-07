<script lang="ts">
	import { resolve } from '$app/paths';
	import type { PageData } from './$types';
	import { format, formatDate } from '$lib/formatting';

	let { data }: { data: PageData } = $props();
	const party = $derived(data.party);

	const displayName = $derived(
		party.display_name || party.legal_name || party.eik
	);

	const isBuyer = $derived(party.as_buyer.contract_count > 0);
	const isSupplier = $derived(party.as_supplier.contract_count > 0);
</script>

<svelte:head>
	<title>{displayName} — изобличи.ме</title>
</svelte:head>

<!-- Breadcrumb -->
<nav
	class="flex items-center gap-2 border-b border-base-content/15 py-3 font-mono text-xs text-base-content/40"
>
	<a href={resolve('/')} class="hover:text-base-content">Начало</a>
	<span>/</span>
	<a href={resolve('/parties')} class="hover:text-base-content">Юридически лица</a>
	<span>/</span>
	<span class="text-base-content/60">{displayName}</span>
</nav>

<!-- Header -->
<section class="border-b border-base-content/15 py-10 md:py-14">
	<p class="mb-3 font-mono text-[10px] tracking-[.22em] text-base-content/40 uppercase">
		Юридическо лице · ЕИК {party.eik}
	</p>
	<h1
		class="font-display mb-2 max-w-3xl text-[clamp(26px,4vw,44px)] leading-[1.15] font-black tracking-tight"
	>
		{displayName}
	</h1>
	{#if party.address_locality || party.address_region}
		<p class="font-mono text-sm text-base-content/50">
			{[party.address_locality, party.address_region].filter(Boolean).join(', ')}
		</p>
	{/if}
</section>

<!-- Role badges -->
<div class="flex flex-wrap gap-3 border-b border-base-content/15 py-5">
	{#if isBuyer}
		<div class="flex items-center gap-2 rounded-sm border border-blue-500/30 bg-blue-500/8 px-3 py-2">
			<span class="font-mono text-[10px] font-semibold tracking-wider text-blue-500 uppercase">
				Възложител
			</span>
			<span class="font-mono text-xs text-base-content/60">
				{format(party.as_buyer.contract_count)} договора
			</span>
		</div>
	{/if}
	{#if isSupplier}
		<div class="flex items-center gap-2 rounded-sm border border-[#B85C38]/30 bg-[#B85C38]/8 px-3 py-2">
			<span class="font-mono text-[10px] font-semibold tracking-wider text-[#B85C38] uppercase">
				Изпълнител
			</span>
			<span class="font-mono text-xs text-base-content/60">
				{format(party.as_supplier.contract_count)} договора
			</span>
		</div>
	{/if}
	{#if !isBuyer && !isSupplier}
		<p class="font-mono text-xs text-base-content/40">
			Няма регистрирани договори в системата.
		</p>
	{/if}
</div>

<!-- Detail rows -->
<div class="divide-y divide-base-content/15 border-b border-base-content/15">
	{#if party.address_locality}
		<div class="flex items-baseline gap-4 px-0 py-4">
			<span class="w-32 shrink-0 font-mono text-[10px] tracking-wider text-base-content/40 uppercase">
				Населено място
			</span>
			<span class="text-sm text-base-content/80">{party.address_locality}</span>
		</div>
	{/if}
	{#if party.address_region}
		<div class="flex items-baseline gap-4 px-0 py-4">
			<span class="w-32 shrink-0 font-mono text-[10px] tracking-wider text-base-content/40 uppercase">
				Област
			</span>
			<span class="text-sm text-base-content/80">{party.address_region}</span>
		</div>
	{/if}
	{#if party.country}
		<div class="flex items-baseline gap-4 px-0 py-4">
			<span class="w-32 shrink-0 font-mono text-[10px] tracking-wider text-base-content/40 uppercase">
				Държава
			</span>
			<span class="text-sm text-base-content/80">{party.country}</span>
		</div>
	{/if}
	<div class="flex items-baseline gap-4 px-0 py-4">
		<span class="w-32 shrink-0 font-mono text-[10px] tracking-wider text-base-content/40 uppercase">
			Първа поява
		</span>
		<span class="text-sm text-base-content/80">{formatDate(party.first_seen)}</span>
	</div>
	<div class="flex items-baseline gap-4 px-0 py-4">
		<span class="w-32 shrink-0 font-mono text-[10px] tracking-wider text-base-content/40 uppercase">
			Последна поява
		</span>
		<span class="text-sm text-base-content/80">{formatDate(party.last_seen)}</span>
	</div>
</div>

<!-- Actions -->
<div class="flex flex-wrap gap-3 py-8">
	{#if isBuyer}
		<a
			href={resolve(`/buyers/${party.eik}`)}
			class="btn rounded-sm px-5 font-mono text-xs tracking-wider btn-neutral"
		>
			Профил като възложител →
		</a>
	{/if}
	{#if isSupplier}
		<a
			href={resolve(`/suppliers/${party.eik}`)}
			class="btn rounded-sm px-5 font-mono text-xs tracking-wider btn-outline"
		>
			Профил като изпълнител →
		</a>
	{/if}
	<a
		href={resolve(`/contracts?party_eik=${party.eik}`)}
		class="btn rounded-sm px-5 font-mono text-xs tracking-wider btn-ghost"
	>
		Всички договори →
	</a>
</div>

<style>
	.font-display {
		font-family: 'Playfair Display', Georgia, serif;
	}
</style>
