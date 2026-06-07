<script lang="ts">
	import { resolve } from '$app/paths';
	import type { PageData } from './$types';
	import { format, formatCurrency, formatPercent } from '$lib/formatting';

	let { data }: { data: PageData } = $props();
	const profile = $derived(data.profile);

	const maxYearCount = $derived(
		Math.max(...profile.year_breakdown.map((y) => y.contract_count), 1)
	);
</script>

<svelte:head>
	<title>{profile.name} — изобличи.ме</title>
</svelte:head>

<!-- Breadcrumb -->
<nav
	class="flex items-center gap-2 border-b border-base-content/15 py-3 font-mono text-xs text-base-content/40"
>
	<a href={resolve('/')} class="hover:text-base-content">Начало</a>
	<span>/</span>
	<a href={resolve('/parties')} class="hover:text-base-content">Участници</a>
	<span>/</span>
	<span class="text-base-content/60">Купувач</span>
</nav>

<!-- Header -->
<section class="border-b border-base-content/15 py-10 md:py-14">
	<p class="mb-3 font-mono text-[10px] tracking-[.22em] text-base-content/40 uppercase">
		Профил на купувач · ЕИК {profile.eik}
	</p>
	<h1
		class="font-display mb-2 max-w-3xl text-[clamp(26px,4vw,44px)] leading-[1.15] font-black tracking-tight"
	>
		{profile.name}
	</h1>
	{#if profile.address_locality || profile.address_region}
		<p class="font-mono text-sm text-base-content/50">
			{[profile.address_locality, profile.address_region].filter(Boolean).join(', ')}
		</p>
	{/if}
</section>

<!-- Flags -->
{#if profile.flags.supplier_dominance || profile.flags.no_bid_count > 0 || profile.flags.near_threshold_count > 0}
	<div class="flex flex-wrap gap-3 border-b border-base-content/15 py-5">
		{#if profile.flags.supplier_dominance}
			<div
				class="flex items-center gap-2 rounded-sm border border-warning/40 bg-warning/10 px-3 py-2"
			>
				<span class="text-base">⚠️</span>
				<div>
					<p class="font-mono text-[10px] font-semibold tracking-wider text-warning uppercase">
						Доминиращ доставчик
					</p>
					<p class="text-xs text-base-content/60">
						Един доставчик печели {formatPercent(profile.flags.dominant_supplier_pct)} от поръчките
					</p>
				</div>
			</div>
		{/if}
		{#if profile.flags.no_bid_count > 0}
			<div class="flex items-center gap-2 rounded-sm border border-error/30 bg-error/8 px-3 py-2">
				<span class="text-base">🔴</span>
				<div>
					<p class="font-mono text-[10px] font-semibold tracking-wider text-error uppercase">
						Поръчки без конкуренция
					</p>
					<p class="text-xs text-base-content/60">
						{profile.flags.no_bid_count} договора с един кандидат
					</p>
				</div>
			</div>
		{/if}
		{#if profile.flags.near_threshold_count > 0}
			<div class="flex items-center gap-2 rounded-sm border border-info/30 bg-info/8 px-3 py-2">
				<span class="text-base">🔵</span>
				<div>
					<p class="font-mono text-[10px] font-semibold tracking-wider text-info uppercase">
						Близо до прага
					</p>
					<p class="text-xs text-base-content/60">
						{profile.flags.near_threshold_count} договора под прага за открита процедура
					</p>
				</div>
			</div>
		{/if}
	</div>
{/if}

<!-- Stats -->
<div
	class="grid grid-cols-2 divide-x divide-y divide-base-content/15 border-b border-base-content/15 md:grid-cols-4 md:divide-y-0"
>
	<div class="p-5 md:p-6">
		<p class="font-display mb-1 text-2xl font-black text-[#B85C38] sm:text-3xl">
			{format(profile.total_contracts)}
		</p>
		{#if profile.total_contracts == 1}
			<p class="font-mono text-[10px] tracking-widest text-base-content/50 uppercase">Договор</p>
		{:else}
			<p class="font-mono text-[10px] tracking-widest text-base-content/50 uppercase">Договора</p>
		{/if}
	</div>
	<div class="p-5 md:p-6">
		<p class="font-display mb-1 text-2xl font-black text-[#B85C38] sm:text-3xl">
			{formatCurrency(profile.total_value, profile.currency)}
		</p>
		<p class="font-mono text-[10px] tracking-widest text-base-content/50 uppercase">
			Обща стойност
		</p>
	</div>
	<div class="p-5 md:p-6">
		<p class="font-display mb-1 text-2xl font-black text-[#B85C38] sm:text-3xl">
			{profile.top_suppliers.length}
		</p>
		<p class="font-mono text-[10px] tracking-widest text-base-content/50 uppercase">
			Топ доставчици
		</p>
	</div>
	<div class="p-5 md:p-6">
		<p class="font-display mb-1 text-2xl font-black text-[#B85C38] sm:text-3xl">
			{profile.years_active.length > 0
				? profile.years_active[0] + '–' + profile.years_active[profile.years_active.length - 1]
				: '—'}
		</p>
		<p class="font-mono text-[10px] tracking-widest text-base-content/50 uppercase">Период</p>
	</div>
</div>

<!-- Year breakdown -->
{#if profile.year_breakdown.length > 0}
	<section class="border-b border-base-content/15 py-8 md:py-10">
		<p class="mb-6 font-mono text-[10px] tracking-[.22em] text-base-content/40 uppercase">
			Активност по години
		</p>
		<div class="space-y-2">
			{#each profile.year_breakdown as row (row.year)}
				<div class="flex items-center gap-4">
					<span class="w-10 shrink-0 font-mono text-xs text-base-content/50">{row.year}</span>
					<div class="flex flex-1 items-center gap-2">
						<div
							class="h-5 min-w-0.5 rounded-sm bg-[#B85C38]/80 transition-all"
							style="width: {(row.contract_count / maxYearCount) * 100}%"
						></div>
						<span class="shrink-0 font-mono text-xs text-base-content/60">{row.contract_count}</span
						>
					</div>
					<span class="w-28 shrink-0 text-right font-mono text-xs text-base-content/40"
						>{formatCurrency(row.total_value, 'BGN')}</span
					>
				</div>
			{/each}
		</div>
	</section>
{/if}

<!-- Top suppliers -->
{#if profile.top_suppliers.length > 0}
	<section class="border-b border-base-content/15 py-8 md:py-10">
		<p class="mb-6 font-mono text-[10px] tracking-[.22em] text-base-content/40 uppercase">
			Топ доставчици
		</p>
		<div class="overflow-x-auto">
			<table class="w-full text-sm">
				<thead>
					<tr class="border-b border-base-content/15">
						<th
							class="pb-2 text-left font-mono text-[10px] tracking-wider text-base-content/40 uppercase"
							>#</th
						>
						<th
							class="pb-2 text-left font-mono text-[10px] tracking-wider text-base-content/40 uppercase"
							>Доставчик</th
						>
						<th
							class="pb-2 text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase"
							>Договори</th
						>
						<th
							class="pb-2 text-right font-mono text-[10px] tracking-wider text-base-content/40 uppercase"
							>Стойност</th
						>
						<th
							class="pb-2 pl-4 font-mono text-[10px] tracking-wider text-base-content/40 uppercase"
							>Дял</th
						>
					</tr>
				</thead>
				<tbody class="divide-y divide-base-content/10">
					{#each profile.top_suppliers as s, i (i)}
						<tr class="group">
							<td class="py-3 pr-3 font-mono text-xs text-base-content/30">{i + 1}</td>
							<td class="py-3 pr-4">
								{#if s.eik}
									<a
										href={resolve(`/suppliers/${s.eik}`)}
										class="font-medium text-base-content transition-colors group-hover:text-[#B85C38] hover:underline"
									>
										{s.name ?? s.eik}
									</a>
								{:else}
									<span class="text-base-content/60">{s.name ?? '—'}</span>
								{/if}
							</td>
							<td class="py-3 pr-4 text-right font-mono text-xs text-base-content/70">{s.wins}</td>
							<td class="py-3 pr-4 text-right font-mono text-xs text-base-content/70"
								>{formatCurrency(s.total_value, 'BGN')}</td
							>
							<td class="py-3 pl-4">
								<div class="flex items-center gap-2">
									<div class="h-2 w-20 overflow-hidden rounded-sm bg-base-200">
										<div
											class="h-full rounded-sm bg-[#B85C38]/70"
											style="width: {s.pct_by_count}%"
										></div>
									</div>
									<span class="font-mono text-[10px] text-base-content/50"
										>{formatPercent(s.pct_by_count)}</span
									>
								</div>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</section>
{/if}

<!-- Actions -->
<div class="flex flex-wrap gap-3 py-8">
	<a
		href={resolve(`/contracts?buyer_eik=${profile.eik}`)}
		class="btn rounded-sm px-5 font-mono text-xs tracking-wider btn-neutral"
	>
		Всички договори →
	</a>
	<a
		href={resolve(`/anomalies?buyer_eik=${profile.eik}`)}
		class="btn rounded-sm px-5 font-mono text-xs tracking-wider btn-outline"
	>
		Аномалии за този купувач
	</a>
</div>

<style>
	.font-display {
		font-family: 'Playfair Display', Georgia, serif;
	}
</style>
