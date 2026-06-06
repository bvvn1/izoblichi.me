<script lang="ts">
	import { resolve } from '$app/paths';

	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	const formatter = new Intl.NumberFormat('bg-BG', {
		notation: 'compact',
		maximumFractionDigits: 1
	});
	const bgnFormatter = new Intl.NumberFormat('bg-BG', {
		notation: 'compact',
		maximumFractionDigits: 1,
		style: 'currency',
		currency: 'BGN'
	});

	const format = (n: number | null) => (n == null ? '—' : formatter.format(n));
	const formatBgn = (n: number | null) => (n == null ? '—' : bgnFormatter.format(n));
</script>

<section class="border-b border-base-content/15 py-16 md:py-24">
	<p class="text-l mb-6 font-mono font-semibold tracking-[.2em] text-base-content/40 uppercase">
		изобличи.ме
	</p>

	<h1
		class="font-display mb-6 max-w-2xl text-[clamp(38px,7vw,68px)] leading-[1.1] font-black tracking-tight"
	>
		Проследи
		<span class="font-serif font-normal text-base-content/40 italic underline">обществените</span>
		пари.
	</h1>

	<p class="balancing mb-8 max-w-xl text-base leading-relaxed text-base-content/60 md:text-lg">
		Търсете, филтрирайте и анализирайте договорите, сключени от държавните институции. Откривайте
		аномалии, разплитайте връзките между купувачи и доставчици и изисквайте прозрачност.
	</p>

	<div class="flex flex-wrap gap-3">
		<a
			href={resolve('/contracts')}
			class="btn rounded-sm px-5 font-mono text-xs tracking-wider transition-all btn-neutral hover:gap-3"
		>
			Разгледай договорите →
		</a>
		<a
			href={resolve('/anomalies')}
			class="btn rounded-sm px-5 font-mono text-xs tracking-wider btn-outline"
		>
			Виж аномалиите
		</a>
	</div>
</section>

<div
	class="grid grid-cols-1 divide-y divide-base-content/15 border-b border-base-content/15 md:grid-cols-3 md:divide-x md:divide-y-0"
>
	<div class="py-6 pr-6 md:py-8">
		<p class="font-display mb-2 text-4xl font-black tracking-tight lg:text-5xl">
			<span class="text-[#B85C38]">{format(data.stats.total_contracts)}</span>
		</p>
		<p class="font-mono text-[10px] tracking-widest text-base-content/50 uppercase md:text-xs">
			Индексирани договори
		</p>
	</div>

	<div class="py-6 pr-6 md:py-8 md:pl-8">
		<p class="font-display mb-2 text-4xl font-black tracking-tight lg:text-5xl">
			<span class="text-[#B85C38]">{formatBgn(data.stats.total_value_bgn)}</span>
		</p>
		<p class="font-mono text-[10px] tracking-widest text-base-content/50 uppercase md:text-xs">
			Обща стойност
		</p>
	</div>

	<div class="py-6 md:py-8 md:pl-8">
		<p class="font-display mb-2 text-4xl font-black tracking-tight lg:text-5xl">
			<span class="text-[#B85C38]"
				>{format(data.stats.unique_buyers + data.stats.unique_suppliers)}</span
			>
		</p>
		<p class="font-mono text-[10px] tracking-widest text-base-content/50 uppercase md:text-xs">
			Уникални участници
		</p>
	</div>
</div>

<section class="border-b border-base-content/15 py-12 md:py-16">
	<p class="mb-8 font-mono text-xs tracking-[.2em] text-base-content/40 uppercase">
		Възможности на платформата
	</p>

	<div class="grid grid-cols-1 border-t border-l border-base-content/15 md:grid-cols-2">
		{#each [{ icon: '🔍', title: 'Търсене на договори', desc: 'Пълнотекстово търсене в реално време по заглавия, ключови думи, години и филтри за участници.', href: '/contracts' }, { icon: '🏛', title: 'Профили на купувачи', desc: 'Всеки възложител (министерства, общини, ведомства) с пълната си история на обществените поръчки.', href: '/buyers' }, { icon: '💼', title: 'Профили на доставчици', desc: 'Проследете кои компании печелят договори — разбивка по министерства, сектори и години.', href: '/suppliers' }, { icon: '⚠️', title: 'Сигнали за аномалии', desc: 'Автоматично засичане на прагове, поръчки с един кандидат и съмнителна концентрация на средства.', href: '/anomalies' }] as const as feature (feature)}
			<a
				href={resolve(feature.href)}
				class="group block border-r border-b border-base-content/15 p-6 transition-all duration-200 hover:bg-base-200/50 lg:p-8"
			>
				<div class="mb-4 inline-block text-2xl transition-transform group-hover:scale-110">
					{feature.icon}
				</div>
				<p
					class="font-display mb-2 text-lg font-bold text-base-content transition-colors group-hover:text-[#B85C38]"
				>
					{feature.title}
				</p>
				<p class="text-sm leading-relaxed text-base-content/60">{feature.desc}</p>
			</a>
		{/each}
	</div>
</section>

<div
	class="to-neutral-focus relative my-12 overflow-hidden rounded-sm bg-gradient-to-br from-neutral p-8 text-neutral-content shadow-lg md:p-12"
>
	<div
		class="pointer-events-none absolute -right-10 -bottom-10 h-40 w-40 rounded-full bg-[#B85C38]/10 blur-3xl"
	></div>

	<p
		class="mb-4 flex items-center gap-2 font-mono text-xs tracking-[.2em] text-neutral-content/50 uppercase"
	>
		<span class="h-2 w-2 animate-pulse rounded-full bg-[#B85C38]"></span>
		Последен анализ на данните
	</p>

	<h2 class="font-display mb-4 max-w-xl text-3xl leading-tight font-bold md:text-4xl">
		Засечени са аномалии в последните договори.
	</h2>

	<p class="mb-6 max-w-xl text-sm leading-relaxed text-neutral-content/75 md:text-base">
		Алгоритъмът ни изолира случаи на необичайно висока концентрация на доставчици, договори на ръба
		на прага за директно възлагане и системно явяване на единствен кандидат.
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
