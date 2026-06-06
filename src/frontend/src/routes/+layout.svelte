<script lang="ts">
	import '../app.css';
    const tickerEntries = ["Купувачи", "Анализ", "Прозрачност"];
	import type { LayoutData } from './$types'
  	let { children, data }: { children, data: LayoutData } = $props();
</script>

<div class="max-w-5xl mx-auto px-6">
<nav class="flex items-center justify-between py-4 border-b-2 border-base-content">
  <span class="font-mono text-xs font-medium tracking-widest uppercase">
    Обществени поръчки
  </span>
  <span class="bg-base-content text-base-100 font-mono text-xs px-3 py-1 rounded-sm tracking-widest">
    Open Data
  </span>
</nav>

<div class="w-full overflow-hidden border-b border-base-content/15 py-2 select-none">
  <div class="ticker-track flex w-max">
    
    <div class="ticker-items flex gap-8 pr-8">
      {#each tickerEntries as stat (stat)}
        {@render ticker_item(stat)}
      {/each}
    </div>

    <div class="ticker-items flex gap-8 pr-8" aria-hidden="true">
      {#each tickerEntries as stat (stat)}
        {@render ticker_item(stat)}
      {/each}
    </div>

  </div>
</div>

{@render children()}

<footer class="py-6 flex items-center justify-between border-t border-base-content/15">
  <span class="font-mono text-xs text-base-content/40">
  Data through {data.stats.data_through}
</span>
  <span class="font-mono text-xs text-base-content/40">Open source</span>
</footer>
</div>

<svelte:head>
    <title>изобличи.ме</title>
    <link rel="preconnect" href="https://fonts.googleapis.com" />
    <link href="https://fonts.googleapis.com/css2?family=Playfair+Display:wght@700;900&family=IBM+Plex+Mono:wght@400;500&family=IBM+Plex+Sans:wght@400;500&display=swap" rel="stylesheet" />
</svelte:head>

{#snippet ticker_item(entry: string)}
  <span class="font-mono text-xs tracking-widest uppercase text-base-content/50 inline-flex items-center gap-2 shrink-0">
    <span class="w-1.5 h-1.5 rounded-full bg-[#B85C38] shrink-0"></span>
    {entry} · {entry} · {entry}
  </span>
{/snippet}

<style>
  .ticker-track {
    display: flex;
    width: max-content;
  }

  .ticker-items {
    display: flex;
    white-space: nowrap;
    /* Adjust time (20s) depending on how fast you want it to scroll */
    animation: ticker-loop 20s linear infinite; 
  }

  @keyframes ticker-loop {
    0% {
      transform: translateX(0);
    }
    100% {
      transform: translateX(-100%); 
    }
  }
</style>
