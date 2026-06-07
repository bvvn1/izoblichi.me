<script lang="ts">
	import { api } from '$lib/api';
	import type { AutocompleteResponse } from '$lib/types';
	import { goto } from '$app/navigation';
	import { resolve } from '$app/paths';

	let query = $state('');
	let results: AutocompleteResponse | null = $state(null);
	let loading = $state(false);
	let open = $state(false);
	let activeIndex = $state(-1);
	let inputEl: HTMLInputElement | undefined = $state();
	let timer: ReturnType<typeof setTimeout>;

	const buyerCount = $derived(results?.buyers.length ?? 0);
	const supplierCount = $derived(results?.suppliers.length ?? 0);
	const contractCount = $derived(results?.contracts.length ?? 0);
	const totalCount = $derived(buyerCount + supplierCount + contractCount);
	const showDropdown = $derived(open && query.length >= 2);
	const empty = $derived(results !== null && totalCount === 0 && !loading);

	const flat = $derived.by((): string[] => {
		if (!results) return [];
		return [
			...results.buyers.map((b) => resolve(`/buyers/${b.eik}`)),
			...results.suppliers.map((s) => resolve(`/suppliers/${s.eik}`)),
			...results.contracts.map((c) => resolve(`/contracts/${c.source}/${c.id}`))
		];
	});

	function onInput() {
		clearTimeout(timer);
		activeIndex = -1;
		if (query.length < 2) {
			results = null;
			return;
		}
		loading = true;
		timer = setTimeout(async () => {
			try {
				results = await api(fetch).autocomplete(query);
			} catch {
				results = null;
			} finally {
				loading = false;
			}
		}, 250);
	}

	function navigate(href: string) {
		open = false;
		goto(resolve(href));
	}

	function submit() {
		if (activeIndex >= 0 && flat[activeIndex]) {
			navigate(flat[activeIndex]);
			return;
		}
		if (query.trim()) {
			navigate(resolve(`/contracts?q=${encodeURIComponent(query.trim())}`));
		}
	}

	function onKeydown(e: KeyboardEvent) {
		if (e.key === 'ArrowDown') {
			e.preventDefault();
			activeIndex = Math.min(activeIndex + 1, totalCount - 1);
		} else if (e.key === 'ArrowUp') {
			e.preventDefault();
			activeIndex = Math.max(activeIndex - 1, -1);
		} else if (e.key === 'Enter') {
			e.preventDefault();
			submit();
		} else if (e.key === 'Escape') {
			open = false;
			inputEl?.blur();
		}
	}

	function clickOutside(node: HTMLElement) {
		const handler = (e: MouseEvent) => {
			if (!node.contains(e.target as Node)) open = false;
		};
		document.addEventListener('mousedown', handler);
		return { destroy: () => document.removeEventListener('mousedown', handler) };
	}
</script>

<div class="relative w-full" use:clickOutside>
	<!-- Input -->
	<div
		class="flex items-center gap-3 border-2 border-base-content bg-base-100 px-4 py-3.5 transition-shadow focus-within:shadow-[0_4px_24px_rgba(0,0,0,0.12)] md:px-5 md:py-4"
	>
		{#if loading}
			<span class="loading loading-sm shrink-0 loading-spinner opacity-40"></span>
		{:else}
			<svg
				class="h-5 w-5 shrink-0 text-base-content/40"
				fill="none"
				stroke="currentColor"
				stroke-width="2"
				viewBox="0 0 24 24"
			>
				<circle cx="11" cy="11" r="8" />
				<path d="m21 21-4.35-4.35" />
			</svg>
		{/if}

		<input
			bind:this={inputEl}
			bind:value={query}
			oninput={onInput}
			onkeydown={onKeydown}
			onfocus={() => (open = true)}
			type="text"
			placeholder="Търси институция, фирма или договор..."
			class="min-w-0 flex-1 bg-transparent text-base outline-none placeholder:text-base-content/30 md:text-lg"
			autocomplete="off"
			spellcheck="false"
		/>

		{#if query}
			<button
				onclick={() => {
					query = '';
					results = null;
					activeIndex = -1;
					inputEl?.focus();
				}}
				class="shrink-0 text-base-content/30 transition-colors hover:text-base-content"
				aria-label="Изчисти"
			>
				<svg class="h-4 w-4" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
					<path d="M18 6 6 18M6 6l12 12" />
				</svg>
			</button>
		{:else}
			<kbd
				class="hidden shrink-0 rounded border border-base-content/20 px-1.5 py-0.5 font-mono text-[10px] text-base-content/30 sm:block"
				>Enter</kbd
			>
		{/if}
	</div>

	<!-- Dropdown -->
	{#if showDropdown}
		<div
			class="absolute top-full z-50 mt-px w-full border border-t-0 border-base-content/20 bg-base-100 shadow-2xl"
		>
			{#if empty}
				<p class="px-5 py-4 font-mono text-xs text-base-content/40">
					Няма резултати за „{query}"
				</p>
			{:else}
				{#if results?.buyers.length}
					<p
						class="border-b border-base-content/10 px-5 pt-3 pb-1.5 font-mono text-[10px] tracking-widest text-base-content/40 uppercase"
					>
						Купувачи
					</p>
					{#each results.buyers as buyer, i (buyer.eik)}
						{@const idx = i}
						<button
							class="flex w-full items-center gap-3 px-5 py-2.5 text-left text-sm transition-colors hover:bg-base-200 {activeIndex ===
							idx
								? 'bg-base-200'
								: ''}"
							onmouseover={() => (activeIndex = idx)}
							onfocus={() => (activeIndex = idx)}
							onclick={() => navigate(resolve(`/buyers/${buyer.eik}`))}
						>
							<span class="shrink-0 text-base-content/40">🏛</span>
							<span class="truncate">{buyer.name}</span>
						</button>
					{/each}
				{/if}

				{#if results?.suppliers.length}
					<p
						class="border-y border-base-content/10 px-5 pt-3 pb-1.5 font-mono text-[10px] tracking-widest text-base-content/40 uppercase"
					>
						Доставчици
					</p>
					{#each results.suppliers as supplier, i (supplier.eik)}
						{@const idx = buyerCount + i}
						<button
							class="flex w-full items-center gap-3 px-5 py-2.5 text-left text-sm transition-colors hover:bg-base-200 {activeIndex ===
							idx
								? 'bg-base-200'
								: ''}"
							onmouseover={() => (activeIndex = idx)}
							onfocus={() => (activeIndex = idx)}
							onclick={() => navigate(resolve(`/suppliers/${supplier.eik}`))}
						>
							<span class="shrink-0 text-base-content/40">💼</span>
							<span class="truncate">{supplier.name}</span>
						</button>
					{/each}
				{/if}

				{#if results?.contracts.length}
					<p
						class="border-y border-base-content/10 px-5 pt-3 pb-1.5 font-mono text-[10px] tracking-widest text-base-content/40 uppercase"
					>
						Договори
					</p>
					{#each results.contracts as contract, i (contract.id)}
						{@const idx = buyerCount + supplierCount + i}
						<button
							class="flex w-full items-center gap-3 px-5 py-2.5 text-left text-sm transition-colors hover:bg-base-200 {activeIndex ===
							idx
								? 'bg-base-200'
								: ''}"
							onmouseover={() => (activeIndex = idx)}
							onfocus={() => (activeIndex = idx)}
							onclick={() => navigate(resolve(`/contracts/${contract.source}/${contract.id}`))}
						>
							<span class="shrink-0 text-base-content/40">📄</span>
							<span class="line-clamp-1">{contract.title}</span>
						</button>
					{/each}
				{/if}
			{/if}

			<!-- Footer -->
			{#if query.trim().length >= 2}
				<div class="border-t border-base-content/10 px-5 py-2.5">
					<button
						onclick={() => navigate(resolve(`/contracts?q=${encodeURIComponent(query.trim())}`))}
						class="font-mono text-xs tracking-wider text-base-content/50 transition-colors hover:text-base-content"
					>
						Виж всички договори за „{query}" →
					</button>
				</div>
			{/if}
		</div>
	{/if}
</div>
