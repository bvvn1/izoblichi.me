<script lang="ts">
	import { resolve } from '$app/paths';

  import type { PageData } from './$types'

  let { data }: { data: PageData } = $props()

  const format = (n: number) =>
    Intl.NumberFormat('bg-BG', { notation: 'compact', maximumFractionDigits: 1 }).format(n)

  const formatBgn = (n: number) =>
    Intl.NumberFormat('bg-BG', { notation: 'compact', maximumFractionDigits: 1 }).format(n) + ' лв.'
</script>

<section class="py-16 md:py-24 border-b border-base-content/15">
  <p class="font-mono text-l font-semibold tracking-[.2em] uppercase text-base-content/40 mb-6">
    изобличи.ме
  </p>
  
  <h1 class="font-display text-[clamp(38px,7vw,68px)] font-black leading-[1.1] tracking-tight mb-6 max-w-2xl">
    Проследи 
    <span class="italic underline font-serif font-normal text-base-content/40">обществените</span> 
    пари.
  </h1>
  
  <p class="text-base md:text-lg leading-relaxed text-base-content/60 max-w-xl mb-8 balancing">
    Търсете, филтрирайте и анализирайте договорите, сключени от държавните институции. 
    Откривайте аномалии, разплитайте връзките между купувачи и доставчици и изисквайте прозрачност.
  </p>
  
  <div class="flex gap-3 flex-wrap">
    <a href={resolve('/contracts')} class="btn btn-neutral rounded-sm font-mono text-xs tracking-wider px-5 hover:gap-3 transition-all">
      Разгледай договорите →
    </a>
    <a href={resolve("/anomalies")} class="btn btn-outline rounded-sm font-mono text-xs tracking-wider px-5">
      Виж аномалиите
    </a>
  </div>
</section>

<div class="grid grid-cols-3 border-b border-base-content/15">
  <div class="py-7 border-r border-base-content/15">
  <p class="font-display text-4xl font-bold leading-none mb-1">
    <span class="text-[#B85C38]">{format(data.stats.total_contracts)}</span>
  </p>
  <p class="font-mono text-xs tracking-widest uppercase text-base-content/50">Contracts indexed</p>
</div>

<div class="py-7 pl-6 border-r border-base-content/15">
  <p class="font-display text-4xl font-bold leading-none mb-1">
    <span class="text-[#B85C38]">{formatBgn(data.stats.total_value_bgn)}</span>
  </p>
  <p class="font-mono text-xs tracking-widest uppercase text-base-content/50">Total value</p>
</div>

<div class="py-7 pl-6">
  <p class="font-display text-4xl font-bold leading-none mb-1">
    <span class="text-[#B85C38]">{format(data.stats.unique_buyers + data.stats.unique_suppliers)}</span>
  </p>
  <p class="font-mono text-xs tracking-widest uppercase text-base-content/50">Unique parties</p>
</div>
</div>

<section class="py-10 border-b border-base-content/15">
  <p class="font-mono text-xs tracking-[.14em] uppercase text-base-content/50 mb-6">What you can do</p>
  <div class="grid grid-cols-2">
    {#each [
      { icon: '🔍', title: 'Search contracts', desc: 'Full-text search across all contract titles with year and party filters.', href: '/contracts' },
      { icon: '🏛', title: 'Buyer profiles', desc: 'Every contracting authority with their full procurement history.', href: '/buyers' },
      { icon: '💼', title: 'Supplier profiles', desc: 'Trace who wins contracts — across ministries, years, and sectors.', href: '/suppliers' },
      { icon: '⚠️', title: 'Anomaly flags', desc: 'Automated detection of threshold proximity, single-bidder awards, and concentration patterns.', href: '/anomalies' },
    ] as const as feature, index (feature)}
      <a href={resolve(feature.href)}
        class="block p-6 border-base-content/15 hover:bg-base-200 transition-colors
          {index % 2 === 0 ? 'pr-6 border-r' : 'pl-6'}
          {index < 2 ? 'border-b' : ''}">
        <p class="text-xl mb-3">{feature.icon}</p>
        <p class="font-medium text-sm mb-1">{feature.title}</p>
        <p class="text-sm text-base-content/55 leading-relaxed">{feature.desc}</p>
      </a>
    {/each}
  </div>
</section>

<!-- Anomaly CTA -->
<div class="bg-base-content text-base-100 p-8 my-10 rounded-sm">
  <p class="font-mono text-xs tracking-[.14em] uppercase opacity-50 mb-3">Latest intelligence</p>
  <h2 class="font-display text-3xl font-bold leading-tight mb-3">
    Anomalies detected<br>in recent awards.
  </h2>
  <p class="text-sm opacity-70 leading-relaxed max-w-md mb-5">
    Our engine flags high supplier concentration, contracts suspiciously close to
    direct-award thresholds, repeated single-bidder patterns, and more.
  </p>
  <a href={resolve("/anomalies")} class="font-mono text-xs tracking-widest uppercase border-b border-base-100/30 pb-0.5 hover:border-base-100 transition-colors">
    Explore anomalies →
  </a>
</div>

<style>
  .font-display { font-family: 'Playfair Display', serif; }
</style>