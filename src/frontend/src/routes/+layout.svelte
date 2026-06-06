<script lang="ts">
	import '../app.css';
	const tickerEntries = ['Купувачи', 'Анализ', 'Прозрачност'];
	import type { LayoutData } from './$types';
	let { children, data }: { children; data: LayoutData } = $props();
</script>

<div class="mx-auto max-w-5xl px-6">
	<nav class="flex items-center justify-between border-b-2 border-base-content py-4">
		<span class="font-mono text-xs font-medium tracking-widest uppercase">
			Обществени поръчки
		</span>
		<span
			class="rounded-sm bg-base-content px-3 py-1 font-mono text-xs tracking-widest text-base-100"
		>
			Open Data
		</span>
	</nav>

	<div class="w-full overflow-hidden border-b border-base-content/15 py-2 select-none">
		<div class="flex min-w-full">
			<div class="ticker-items flex min-w-full shrink-0 justify-around gap-8 pr-8">
				{#each tickerEntries as stat (stat)}
					{@render ticker_item(stat)}
				{/each}
			</div>
			<div
				class="ticker-items flex min-w-full shrink-0 justify-around gap-8 pr-8"
				aria-hidden="true"
			>
				{#each tickerEntries as stat (stat)}
					{@render ticker_item(stat)}
				{/each}
			</div>
		</div>
	</div>

	{@render children()}

	<footer class="flex items-center justify-between border-t border-base-content/15 py-6">
		<span class="font-mono text-xs text-base-content/40">
			Data through {data.stats.data_through}
		</span>
		<span class="font-mono text-xs text-base-content/40">Open source</span>
	</footer>
</div>

<svelte:head>
	<title>изобличи.ме</title>
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link
		href="https://fonts.googleapis.com/css2?family=Playfair+Display:ital,wght@0,700;0,900;1,400&family=Plus+Jakarta+Sans:wght@400;500;600&display=swap"
		rel="stylesheet"
	/>
</svelte:head>

{#snippet ticker_item(entry: string)}
	<span
		class="inline-flex shrink-0 items-center gap-2 font-mono text-xs tracking-widest text-base-content/50 uppercase"
	>
		<span class="h-1.5 w-1.5 shrink-0 rounded-full bg-[#B85C38]"></span>
		{entry}
	</span>
{/snippet}

<style>
	.ticker-items {
		animation: ticker-loop 25s linear infinite;
	}

	@keyframes ticker-loop {
		0% {
			transform: translate3d(0, 0, 0);
		}
		100% {
			transform: translate3d(-100%, 0, 0);
		}
	}
</style>
