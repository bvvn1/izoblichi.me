<script lang="ts">
	import { resolve } from '$app/paths';
	import { format, formatCurrency } from '$lib/formatting';
	import type { PageData } from './$types';
	import Search from '$lib/Search.svelte';
	import { Tween } from 'svelte/motion';
	import { cubicOut, cubicInOut } from 'svelte/easing';
	import { onMount } from 'svelte';

	let { data }: { data: PageData } = $props();

	const contractsTween = $state(
		new Tween(0, {
			duration: 2500,
			easing: cubicOut
		})
	);

	const valueTween = $state(
		new Tween(0, {
			duration: 3000,
			easing: cubicInOut
		})
	);

	const participantsTween = $state(
		new Tween(0, {
			duration: 2500,
			easing: cubicOut
		})
	);

	onMount(() => {
		contractsTween.set(data.stats.total_contracts);
		valueTween.set(data.stats.total_value_bgn ?? 0);
		participantsTween.set(data.stats.unique_buyers + data.stats.unique_suppliers);
	});
</script>

<section class="border-b border-base-content/15 pt-14 pb-12 md:pt-20 md:pb-16">
	<p class="mb-5 font-mono text-[11px] tracking-[.22em] text-base-content/40 uppercase">
		изобличи.ме — обществени поръчки
	</p>

	<h1
		class="font-display mb-3 max-w-2xl text-[clamp(34px,6vw,60px)] leading-[1.1] font-black tracking-tight"
	>
		Правим
		<span class="font-serif font-normal text-base-content/35 italic">корупцията</span>
		<span class="mx-1 font-black text-error line-through decoration-error decoration-4">(не)</span
		>видима.
	</h1>

	<p class="mb-10 max-w-lg text-sm leading-relaxed text-base-content/50 md:text-base">
		Търсете договори, разплитайте мрежи и откривайте аномалии в публичните поръчки на България.
	</p>

	<div class="max-w-3xl">
		<Search />
		<p class="mt-3 font-mono text-[10px] tracking-widest text-base-content/30 uppercase">
			Опитайте: „Министерство на финансите", „болница", „строителство"...
		</p>
	</div>
</section>

<!-- Stats -->
<div
	class="grid grid-cols-1 divide-y divide-base-content/15 border-b border-base-content/15 md:flex md:items-center md:divide-x md:divide-y-0"
>
	<div class="flex-1 py-6 pr-6 md:py-8">
		<p class="font-display mb-2 text-3xl font-black tracking-tight sm:text-4xl xl:text-5xl">
			<span class="text-[#B85C38]">{format(Math.floor(contractsTween.current))}</span>
		</p>
		<p class="font-mono text-[10px] tracking-widest text-base-content/50 uppercase">
			Индексирани договори
		</p>
	</div>

	<div class="flex-1 py-6 pr-6 md:py-8 md:pl-8">
		<p class="font-display mb-2 text-3xl font-black tracking-tight sm:text-4xl xl:text-5xl">
			<span class="text-[#B85C38]">{formatCurrency(Math.floor(valueTween.current), 'BGN')}</span>
		</p>
		<p class="font-mono text-[10px] tracking-widest text-base-content/50 uppercase">
			Обща стойност
		</p>
	</div>

	<div class="flex-1 py-6 md:py-8 md:pl-8">
		<p class="font-display mb-2 text-3xl font-black tracking-tight sm:text-4xl xl:text-5xl">
			<span class="text-[#B85C38]">{format(Math.floor(participantsTween.current))}</span>
		</p>
		<p class="font-mono text-[10px] tracking-widest text-base-content/50 uppercase">
			Уникални участници
		</p>
	</div>
</div>

<section class="border-b border-base-content/15 py-12 md:py-14">
	<p class="mb-7 font-mono text-[10px] tracking-[.22em] text-base-content/40 uppercase">
		Инструменти
	</p>
	<div class="grid grid-cols-1 border-t border-l border-base-content/15 md:grid-cols-2">
		{#each [{ icon: '📋', title: 'Договори', desc: 'Пълнотекстово търсене по заглавие, купувач, доставчик, година и стойност.', href: '/contracts' }, { icon: '🏛', title: 'Купувачи', desc: 'Профил на всеки възложител — история, топ доставчици, концентрация и сигнали.', href: '/parties' }, { icon: '💼', title: 'Доставчици', desc: 'Кои фирми печелят и от кого — разбивка по сектори, години и институции.', href: '/parties' }, { icon: '⚠️', title: 'Аномалии', desc: 'Автоматично засичане на прагове, поръчки с един кандидат и доминиращи доставчици.', href: '/anomalies' }] as const as feature (feature)}
			<a
				href={resolve(feature.href)}
				class="group block border-r border-b border-base-content/15 p-6 transition-colors hover:bg-base-200/50 lg:p-7"
			>
				<div class="mb-3 text-xl transition-transform group-hover:scale-105">{feature.icon}</div>
				<p
					class="font-display mb-1.5 text-base font-bold transition-colors group-hover:text-[#B85C38]"
				>
					{feature.title}
				</p>
				<p class="text-sm leading-relaxed text-base-content/55">{feature.desc}</p>
			</a>
		{/each}
	</div>
</section>

<div
	class="to-neutral-focus relative my-12 overflow-hidden rounded-sm bg-linear-to-br from-neutral p-8 text-neutral-content shadow-lg md:p-12"
>
	<div
		class="pointer-events-none absolute -right-10 -bottom-10 h-40 w-40 rounded-full bg-[#B85C38]/10 blur-3xl"
	></div>
	<p
		class="mb-4 flex items-center gap-2 font-mono text-xs tracking-[.2em] text-neutral-content/50 uppercase"
	>
		<span class="h-2 w-2 animate-pulse rounded-full bg-[#B85C38]"></span>
		Последен анализ
	</p>
	<h2 class="font-display mb-3 max-w-xl text-2xl leading-tight font-bold md:text-3xl">
		Засечени са аномалии в последните договори.
	</h2>
	<p class="mb-6 max-w-xl text-sm leading-relaxed text-neutral-content/70 md:text-base">
		Необичайна концентрация на доставчици, договори на ръба на прага за директно възлагане и
		системно явяване на единствен кандидат.
	</p>
	<a
		href={resolve('/anomalies')}
		class="inline-flex items-center gap-2 border-b border-neutral-content/30 pb-1 font-mono text-xs tracking-widest uppercase transition-all duration-200 hover:gap-3 hover:border-neutral-content"
	>
		Изследвай аномалиите →
	</a>
</div>

<style>
	:global(body) {
		font-family: 'Plus Jakarta Sans', sans-serif;
	}

	.font-display {
		font-family: 'Playfair Display', Georgia, serif;
	}
</style>
