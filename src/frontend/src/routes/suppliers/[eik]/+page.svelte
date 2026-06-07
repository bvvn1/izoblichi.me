<script lang="ts">
	import { resolve } from '$app/paths';
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';
	import { format, formatCurrency, formatPercent } from '$lib/formatting';

	let { data }: { data: PageData } = $props();
	const profile = $derived(data.profile);
	const network = $derived(data.network);

	const maxYearCount = $derived(
		Math.max(...profile.year_breakdown.map((y) => y.contract_count), 1)
	);

	const yearsActive = $derived(
		profile.year_breakdown.map((y) => y.year)
	);

	// Graph state
	let hoveredNodeId = $state<string | null>(null);
	let tooltipX = $state(0);
	let tooltipY = $state(0);
	let svgEl = $state<SVGSVGElement | null>(null);
	let containerEl = $state<HTMLDivElement | null>(null);
	let containerWidth = $state(400);

	// Reactive SVG size based on container
	$effect(() => {
		const el = containerEl;
		if (!el) return;
		const ro = new ResizeObserver(([entry]) => {
			containerWidth = Math.min(entry.contentRect.width, 640);
		});
		ro.observe(el);
		return () => ro.disconnect();
	});

	const SVG_SIZE = $derived(containerWidth);
	const CX = $derived(SVG_SIZE / 2);
	const CY = $derived(SVG_SIZE / 2);

	// The first node is the supplier itself, the rest are buyers
	const supplierNode = $derived(network.nodes[0]);
	const buyerNodes = $derived(network.nodes.slice(1));

	// Map edge data by buyer EIK (edge source) for quick lookup
	const edgeMap = $derived(
		new Map(network.edges.map((e) => [e.source, e]))
	);

	// Per-buyer edge contract counts for radial positioning
	const buyerEdgeCounts = $derived(
		buyerNodes.map((n) => {
			const edge = edgeMap.get(n.id);
			return parseInt(edge?.label ?? '1') || 1;
		})
	);

	const minEdgeCount = $derived(Math.min(...buyerEdgeCounts, 1));
	const maxEdgeCount_ = $derived(Math.max(...buyerEdgeCounts, 1));

	const maxEdgeContracts = $derived(maxEdgeCount_);
	const maxBuyerSize = $derived(
		Math.max(...buyerNodes.map((n) => n.size), 1)
	);

	// Force-simulated positions (animated from home positions)
	let nodePositions = $state<{ x: number; y: number }[]>([]);

	// SSR-safe display positions: use animated positions when available,
	// fall back to computed home positions for server-side rendering
	const displayPositions = $derived(
		nodePositions.length > 0
			? nodePositions
			: buyerNodes.map((_n, i) => {
					const angle = (2 * Math.PI * i) / buyerNodes.length - Math.PI / 2;
					const count = buyerEdgeCounts[i];
					const span = maxEdgeCount_ - minEdgeCount || 1;
					const t = (count - minEdgeCount) / span;
					const innerR = SVG_SIZE * 0.16;
					const outerR = SVG_SIZE * 0.40;
					const r = outerR - (outerR - innerR) * t;
					return { x: CX + r * Math.cos(angle), y: CY + r * Math.sin(angle) };
				})
	);

	$effect(() => {
		const N = buyerNodes.length;
		if (N === 0) return;

		// Compute ideal radial home positions
		const home: { x: number; y: number }[] = [];
		for (let i = 0; i < N; i++) {
			const angle = (2 * Math.PI * i) / N - Math.PI / 2;
			const count = buyerEdgeCounts[i];
			const span = maxEdgeCount_ - minEdgeCount || 1;
			const t = (count - minEdgeCount) / span;
			const innerR = SVG_SIZE * 0.16;
			const outerR = SVG_SIZE * 0.40;
			const r = outerR - (outerR - innerR) * t;
			home.push({ x: CX + r * Math.cos(angle), y: CY + r * Math.sin(angle) });
		}

		// Start from home positions
		let pos = home.map((p) => ({ ...p }));
		nodePositions = home.map((p) => ({ ...p }));

		let iter = 0;
		const MAX_ITER = 70;
		let cancelled = false;

		function step() {
			if (cancelled || iter >= MAX_ITER) return;

			// Collision resolution: push apart overlapping nodes
			for (let i = 0; i < N; i++) {
				for (let j = i + 1; j < N; j++) {
					const dx = pos[j].x - pos[i].x;
					const dy = pos[j].y - pos[i].y;
					const dist = Math.sqrt(dx * dx + dy * dy) || 1;
					const ri = 8 + (buyerNodes[i].size / maxBuyerSize) * 24;
					const rj = 8 + (buyerNodes[j].size / maxBuyerSize) * 24;
					const minDist = (ri + rj) * 1.5 + 28;
					if (dist < minDist) {
						const push = (minDist - dist) / 2;
						pos[i].x -= (push * dx) / dist;
						pos[i].y -= (push * dy) / dist;
						pos[j].x += (push * dx) / dist;
						pos[j].y += (push * dy) / dist;
					}
				}
			}

			// Home attraction: pull back toward ideal position
			for (let i = 0; i < N; i++) {
				pos[i].x += (home[i].x - pos[i].x) * 0.08;
				pos[i].y += (home[i].y - pos[i].y) * 0.08;
			}

			// Clamp to SVG bounds
			const margin = 20;
			for (let i = 0; i < N; i++) {
				pos[i].x = Math.max(margin, Math.min(SVG_SIZE - margin, pos[i].x));
				pos[i].y = Math.max(margin, Math.min(SVG_SIZE - margin, pos[i].y));
			}

			nodePositions = pos.map((p) => ({ ...p }));
			iter++;
			requestAnimationFrame(step);
		}

		requestAnimationFrame(step);
		return () => {
			cancelled = true;
		};
	});

	function svgCoords(e: MouseEvent): { x: number; y: number } {
		const svg = svgEl;
		if (!svg) return { x: e.offsetX, y: e.offsetY };
		const rect = svg.getBoundingClientRect();
		const scaleX = SVG_SIZE / rect.width;
		const scaleY = SVG_SIZE / rect.height;
		return {
			x: (e.clientX - rect.left) * scaleX,
			y: (e.clientY - rect.top) * scaleY
		};
	}

	function nodeRadius(size: number): number {
		return 8 + (size / maxBuyerSize) * 24;
	}

	function edgeThickness(label: string): number {
		const count = parseInt(label) || 1;
		return 1 + (count / maxEdgeContracts) * 5;
	}

	function truncatedLabel(name: string): string {
		return name.length > 22 ? name.slice(0, 20) + '…' : name;
	}

	function handleNodeHover(e: MouseEvent, nodeId: string) {
		hoveredNodeId = nodeId;
		const coords = svgCoords(e);
		tooltipX = coords.x;
		tooltipY = coords.y;
	}

	function handleNodeLeave() {
		hoveredNodeId = null;
	}

	function handleBuyerClick(eik: string) {
		goto(resolve(`/buyers/${eik}`));
	}
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
	<a href={resolve('/suppliers')} class="hover:text-base-content">Доставчици</a>
	<span>/</span>
	<span class="text-base-content/60">Доставчик</span>
</nav>

<!-- Header -->
<section class="border-b border-base-content/15 py-10 md:py-14">
	<p class="mb-3 font-mono text-[10px] tracking-[.22em] text-base-content/40 uppercase">
		Профил на доставчик · ЕИК {profile.eik}
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
{#if profile.flags.buyer_concentration}
	<div class="flex flex-wrap gap-3 border-b border-base-content/15 py-5">
		<div
			class="flex items-center gap-2 rounded-sm border border-warning/40 bg-warning/10 px-3 py-2"
		>
			<span class="text-base">⚠️</span>
			<div>
				<p class="font-mono text-[10px] font-semibold tracking-wider text-warning uppercase">
					Концентрация при един купувач
				</p>
				<p class="text-xs text-base-content/60">
					Един купувач представлява {formatPercent(profile.flags.top_buyer_pct)} от договорите
				</p>
			</div>
		</div>
	</div>
{/if}

<!-- Stats -->
<div
	class="grid grid-cols-2 divide-x divide-y divide-base-content/15 border-b border-base-content/15 md:grid-cols-4 md:divide-y-0"
>
	<div class="p-5 md:p-6">
		<p class="font-display mb-1 text-2xl font-black text-[#B85C38] sm:text-3xl">
			{format(profile.total_wins)}
		</p>
		{#if profile.total_wins === 1}
			<p class="font-mono text-[10px] tracking-widest text-base-content/50 uppercase">Договор</p>
		{:else}
			<p class="font-mono text-[10px] tracking-widest text-base-content/50 uppercase">Договора</p>
		{/if}
	</div>
	<div class="p-5 md:p-6">
		<p class="font-display mb-1 text-2xl font-black text-[#B85C38] sm:text-3xl">
			{formatCurrency(profile.total_value, 'BGN')}
		</p>
		<p class="font-mono text-[10px] tracking-widest text-base-content/50 uppercase">
			Обща стойност
		</p>
	</div>
	<div class="p-5 md:p-6">
		<p class="font-display mb-1 text-2xl font-black text-[#B85C38] sm:text-3xl">
			{profile.top_buyers.length}
		</p>
		<p class="font-mono text-[10px] tracking-widest text-base-content/50 uppercase">
			Топ купувачи
		</p>
	</div>
	<div class="p-5 md:p-6">
		<p class="font-display mb-1 text-2xl font-black text-[#B85C38] sm:text-3xl">
			{yearsActive.length > 0
				? yearsActive[0] + '–' + yearsActive[yearsActive.length - 1]
				: '—'}
		</p>
		<p class="font-mono text-[10px] tracking-widest text-base-content/50 uppercase">Период</p>
	</div>
</div>

<!-- Graph Visualization -->
{#if buyerNodes.length > 0}
	<section class="border-b border-base-content/15 py-8 md:py-10">
		<p class="mb-6 font-mono text-[10px] tracking-[.22em] text-base-content/40 uppercase">
			Връзки с купувачи
		</p>
		<div class="relative mx-auto max-w-[640px]" bind:this={containerEl}>
			<svg
				viewBox="0 0 {SVG_SIZE} {SVG_SIZE}"
				class="h-auto w-full"
				role="img"
				aria-label="Граф на връзките между доставчик и купувачи"
				bind:this={svgEl}
			>
				<!-- Edges: lines from each buyer to the center supplier -->
				{#each buyerNodes as buyer, i (buyer.id)}
					{@const pos = displayPositions[i]}
					{@const edge = edgeMap.get(buyer.id)}
					{@const isHovered = hoveredNodeId === buyer.id}
					<line
						x1={pos.x}
						y1={pos.y}
						x2={CX}
						y2={CY}
						stroke={isHovered ? '#B85C38' : '#d4d4d4'}
						stroke-width={isHovered ? edgeThickness(edge?.label ?? '1') + 1 : edgeThickness(edge?.label ?? '1')}
						stroke-linecap="round"
						opacity={hoveredNodeId ? (isHovered ? 1 : 0.15) : 0.7}
						class="transition-all duration-200"
					/>
				{/each}

				<!-- Buyer nodes -->
				{#each buyerNodes as buyer, i (buyer.id)}
					{@const pos = displayPositions[i]}
					{@const r = nodeRadius(buyer.size)}
					{@const isHovered = hoveredNodeId === buyer.id}
					<!-- Invisible larger hit area -->
					<circle
						cx={pos.x}
						cy={pos.y}
						r={r + 6}
						fill="transparent"
						class="cursor-pointer"
						onmouseenter={(e) => handleNodeHover(e, buyer.id)}
						onmouseleave={handleNodeLeave}
						onclick={() => handleBuyerClick(buyer.id)}
					/>
					<!-- Visible node -->
					<circle
						cx={pos.x}
						cy={pos.y}
						r={r}
						fill={isHovered ? '#B85C38' : '#78716c'}
						stroke={isHovered ? '#B85C38' : '#57534e'}
						stroke-width="1.5"
						opacity={hoveredNodeId ? (isHovered ? 1 : 0.35) : 1}
						class="cursor-pointer transition-all duration-200"
						onmouseenter={(e) => handleNodeHover(e, buyer.id)}
						onmouseleave={handleNodeLeave}
						onclick={() => handleBuyerClick(buyer.id)}
					/>
					<!-- Label -->
					<text
						x={pos.x}
						y={pos.y + r + 14}
						text-anchor="middle"
						class="cursor-pointer fill-base-content/60 font-mono text-[10px] transition-colors"
						onmouseenter={(e) => handleNodeHover(e, buyer.id)}
						onmouseleave={handleNodeLeave}
						onclick={() => handleBuyerClick(buyer.id)}
					>
						{truncatedLabel(buyer.label)}
					</text>
				{/each}

				<!-- Center supplier node -->
				<circle
					cx={CX}
					cy={CY}
					r={28}
					fill="#B85C38"
					stroke="#9a4a2e"
					stroke-width="2"
				/>
				<text
					x={CX}
					y={CY}
					text-anchor="middle"
					dominant-baseline="central"
					class="fill-white font-display text-sm font-black"
				>
										Доставчик
				</text>
				<text
					x={CX}
					y={CY + 42}
					text-anchor="middle"
					class="fill-base-content/70 font-mono text-[10px]"
				>
					{format(profile.total_wins)} договора
				</text>

				<!-- Tooltip -->
				{#if hoveredNodeId}
					{@const hovered = buyerNodes.find((n) => n.id === hoveredNodeId)}
					{@const edge = edgeMap.get(hoveredNodeId)}
					{#if hovered}
						{@const tooltipW = 200}
						{@const tooltipH = 60}
						{@const tx = Math.min(Math.max(tooltipX, tooltipW / 2 + 4), SVG_SIZE - tooltipW / 2 - 4)}
						{@const ty = tooltipY - tooltipH - 16 > 0 ? tooltipY - tooltipH - 8 : tooltipY + 20}
						<g transform="translate({tx - tooltipW / 2}, {ty})">
							<rect
								width={tooltipW}
								height={tooltipH}
								rx="4"
								fill="#1c1917"
								stroke="#44403c"
								stroke-width="1"
								opacity="0.95"
							/>
							<text x={8} y={18} class="fill-white font-mono text-[11px] font-medium">
								{truncatedLabel(hovered.label)}
							</text>
							<text x={8} y={34} class="fill-base-content/50 font-mono text-[10px]">
								{edge?.label ?? '—'} · {formatCurrency(hovered.size, 'BGN')}
							</text>
							<text x={8} y={48} class="fill-base-content/40 font-mono text-[9px]">
								Кликни за профил на купувач
							</text>
						</g>
					{/if}
				{/if}
			</svg>

			<!-- Legend -->
			<div class="mt-4 flex flex-wrap justify-center gap-6 font-mono text-[10px] text-base-content/40">
				<span class="flex items-center gap-1.5">
					<span class="inline-block h-2 w-2 rounded-full bg-[#B85C38]"></span>
					Големина = обща стойност
				</span>
				<span class="flex items-center gap-1.5">
					<span class="inline-block h-0.5 w-5 bg-[#78716c]"></span>
					Дебелина = брой договори
				</span>
				<span class="flex items-center gap-1.5 text-base-content/25">Кликни възел → профил</span>
			</div>
		</div>
	</section>
{/if}

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

<!-- Top buyers -->
{#if profile.top_buyers.length > 0}
	<section class="border-b border-base-content/15 py-8 md:py-10">
		<p class="mb-6 font-mono text-[10px] tracking-[.22em] text-base-content/40 uppercase">
			Топ купувачи
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
							>Купувач</th
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
					{#each profile.top_buyers as b, i (i)}
						<tr class="group">
							<td class="py-3 pr-3 font-mono text-xs text-base-content/30">{i + 1}</td>
							<td class="py-3 pr-4">
								{#if b.eik}
									<a
										href={resolve(`/buyers/${b.eik}`)}
										class="font-medium text-base-content transition-colors group-hover:text-[#B85C38] hover:underline"
									>
										{b.name ?? b.eik}
									</a>
								{:else}
									<span class="text-base-content/60">{b.name ?? '—'}</span>
								{/if}
							</td>
							<td class="py-3 pr-4 text-right font-mono text-xs text-base-content/70">{b.wins}</td>
							<td class="py-3 pr-4 text-right font-mono text-xs text-base-content/70"
								>{formatCurrency(b.total_value, 'BGN')}</td
							>
							<td class="py-3 pl-4">
								<div class="flex items-center gap-2">
									<div class="h-2 w-20 overflow-hidden rounded-sm bg-base-200">
										<div
											class="h-full rounded-sm bg-[#B85C38]/70"
											style="width: {b.pct_by_count}%"
										></div>
									</div>
									<span class="font-mono text-[10px] text-base-content/50"
										>{formatPercent(b.pct_by_count)}</span
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
		href={resolve(`/contracts?supplier_eik=${profile.eik}`)}
		class="btn rounded-sm px-5 font-mono text-xs tracking-wider btn-neutral"
	>
		Всички договори →
	</a>
	<a
		href={resolve(`/anomalies?supplier_eik=${profile.eik}`)}
		class="btn rounded-sm px-5 font-mono text-xs tracking-wider btn-outline"
	>
		Аномалии за този доставчик
	</a>
</div>

<style>
	.font-display {
		font-family: 'Playfair Display', Georgia, serif;
	}
</style>
