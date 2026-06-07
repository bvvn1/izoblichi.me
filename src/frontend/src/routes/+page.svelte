<script lang="ts">
	import { resolve } from '$app/paths';
	import { format, formatCurrency } from '$lib/formatting';
	import type { PageData } from './$types';
	import Search from '$lib/Search.svelte';
	import { Tween } from 'svelte/motion';
	import { cubicOut, cubicInOut } from 'svelte/easing';
	import { onMount } from 'svelte';

	let { data }: { data: PageData } = $props();

	const contractsTween = $state(new Tween(0, { duration: 2500, easing: cubicOut }));
	const valueTween = $state(new Tween(0, { duration: 3000, easing: cubicInOut }));
	const participantsTween = $state(new Tween(0, { duration: 2500, easing: cubicOut }));

	onMount(() => {
		contractsTween.set(data.stats.total_contracts);
		valueTween.set(data.stats.total_value_bgn ?? 0);
		participantsTween.set(data.stats.unique_buyers + data.stats.unique_suppliers);
	});
</script>

<!-- Hero -->
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

<!-- Tools -->
<section class="border-b border-base-content/15 py-12 md:py-14">
	<p class="mb-7 font-mono text-[10px] tracking-[.22em] text-base-content/40 uppercase">
		Инструменти
	</p>
	<div class="grid grid-cols-1 border-t border-l border-base-content/15 md:grid-cols-2">
		<a
			href={resolve('/contracts')}
			class="group block border-r border-b border-base-content/15 p-6 transition-colors hover:bg-base-200/50 lg:p-7"
		>
			<div class="mb-3 text-base-content/40 transition-colors group-hover:text-[#B85C38]">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					fill="none"
					viewBox="0 0 24 24"
					stroke-width="1.5"
					stroke="currentColor"
					class="h-5 w-5"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z"
					/>
				</svg>
			</div>
			<p
				class="font-display mb-1.5 text-base font-bold transition-colors group-hover:text-[#B85C38]"
			>
				Договори
			</p>
			<p class="text-sm leading-relaxed text-base-content/55">
				Пълнотекстово търсене по заглавие, купувач, доставчик, година и стойност.
			</p>
		</a>

		<a
			href={resolve('/parties')}
			class="group block border-r border-b border-base-content/15 p-6 transition-colors hover:bg-base-200/50 lg:p-7"
		>
			<div class="mb-3 text-base-content/40 transition-colors group-hover:text-[#B85C38]">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					fill="none"
					viewBox="0 0 24 24"
					stroke-width="1.5"
					stroke="currentColor"
					class="h-5 w-5"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M3.75 21h16.5M4.5 3h15M5.25 3v18m13.5-18v18M9 6.75h1.5m-1.5 3h1.5m-1.5 3h1.5m3-6H15m-1.5 3H15m-1.5 3H15M9 21v-3.375c0-.621.504-1.125 1.125-1.125h3.75c.621 0 1.125.504 1.125 1.125V21"
					/>
				</svg>
			</div>
			<p
				class="font-display mb-1.5 text-base font-bold transition-colors group-hover:text-[#B85C38]"
			>
				Участници
			</p>
			<p class="text-sm leading-relaxed text-base-content/55">
				Профил на всеки купувач и доставчик — история, топ контрагенти и концентрация.
			</p>
		</a>

		<a
			href={resolve('/anomalies')}
			class="group block border-r border-b border-base-content/15 p-6 transition-colors hover:bg-base-200/50 lg:p-7"
		>
			<div class="mb-3 text-base-content/40 transition-colors group-hover:text-[#B85C38]">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					fill="none"
					viewBox="0 0 24 24"
					stroke-width="1.5"
					stroke="currentColor"
					class="h-5 w-5"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z"
					/>
				</svg>
			</div>
			<p
				class="font-display mb-1.5 text-base font-bold transition-colors group-hover:text-[#B85C38]"
			>
				Аномалии
			</p>
			<p class="text-sm leading-relaxed text-base-content/55">
				Автоматично засичане на прагове, поръчки без конкуренция и доминиращи доставчици.
			</p>
		</a>

		<a
			href={resolve('/map')}
			class="group block border-r border-b border-base-content/15 p-6 transition-colors hover:bg-base-200/50 lg:p-7"
		>
			<div class="mb-3 text-base-content/40 transition-colors group-hover:text-[#B85C38]">
				<svg
					xmlns="http://www.w3.org/2000/svg"
					fill="none"
					viewBox="0 0 24 24"
					stroke-width="1.5"
					stroke="currentColor"
					class="h-5 w-5"
				>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M15 10.5a3 3 0 1 1-6 0 3 3 0 0 1 6 0Z"
					/>
					<path
						stroke-linecap="round"
						stroke-linejoin="round"
						d="M19.5 10.5c0 7.142-7.5 11.25-7.5 11.25S4.5 17.642 4.5 10.5a7.5 7.5 0 1 1 15 0Z"
					/>
				</svg>
			</div>
			<p
				class="font-display mb-1.5 text-base font-bold transition-colors group-hover:text-[#B85C38]"
			>
				Карта
			</p>
			<p class="text-sm leading-relaxed text-base-content/55">
				Географска визуализация на купувачите — разходи и рисков индекс по градове.
			</p>
		</a>
	</div>
</section>

<!-- Anomaly criteria -->
<section class="border-b border-base-content/15 py-12 md:py-14">
	<p class="mb-2 font-mono text-[10px] tracking-[.22em] text-base-content/40 uppercase">
		Методология
	</p>
	<h2 class="font-display mb-8 text-xl font-bold md:text-2xl">Как засичаме аномалиите</h2>

	<div
		class="grid grid-cols-1 gap-px overflow-hidden rounded-sm border border-base-content/10 bg-base-content/10 md:grid-cols-2"
	>
		<div class="bg-base-100 p-6 lg:p-7">
			<div class="mb-4 flex items-start gap-3">
				<div class="mt-0.5 shrink-0 text-[#B85C38]">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="h-4 w-4"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M15.75 15.75 21 21m-4.5-9a6.75 6.75 0 1 1-13.5 0 6.75 6.75 0 0 1 13.5 0Z"
						/>
					</svg>
				</div>
				<div>
					<p class="mb-0.5 text-sm font-semibold">Близо до прага</p>
					<p class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">
						Near threshold
					</p>
				</div>
			</div>
			<p class="mb-3 text-sm leading-relaxed text-base-content/60">
				Договорна стойност в рамките на <strong class="text-base-content">5%</strong> под прага за открита
				процедура — признак за умишлено занижаване, за да се избегне конкурс.
			</p>
			<div class="space-y-1">
				<div class="flex items-center justify-between rounded-sm bg-base-200/60 px-3 py-1.5">
					<span class="font-mono text-[10px] text-base-content/50 uppercase">Стоки / Услуги</span>
					<span class="font-mono text-[10px] font-medium">34 000 – 35 791 €</span>
				</div>
				<div class="flex items-center justify-between rounded-sm bg-base-200/60 px-3 py-1.5">
					<span class="font-mono text-[10px] text-base-content/50 uppercase">Строителство</span>
					<span class="font-mono text-[10px] font-medium">128 000 – 135 002 €</span>
				</div>
			</div>
		</div>

		<div class="bg-base-100 p-6 lg:p-7">
			<div class="mb-4 flex items-start gap-3">
				<div class="mt-0.5 shrink-0 text-[#B85C38]">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="h-4 w-4"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z"
						/>
					</svg>
				</div>
				<div>
					<p class="mb-0.5 text-sm font-semibold">Без конкуренция</p>
					<p class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">No bid</p>
				</div>
			</div>
			<p class="mb-3 text-sm leading-relaxed text-base-content/60">
				Поръчката е спечелена с <strong class="text-base-content">≤ 1 подадена оферта</strong> — без реална
				конкуренция. Системното повторение е по-силен сигнал от единичния случай.
			</p>
			<div class="flex items-center justify-between rounded-sm bg-base-200/60 px-3 py-1.5">
				<span class="font-mono text-[10px] text-base-content/50 uppercase">Условие</span>
				<span class="font-mono text-[10px] font-medium">bid_count ≤ 1</span>
			</div>
		</div>

		<div class="bg-base-100 p-6 lg:p-7">
			<div class="mb-4 flex items-start gap-3">
				<div class="mt-0.5 shrink-0 text-[#B85C38]">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="h-4 w-4"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M7.5 21 3 16.5m0 0L7.5 12M3 16.5h13.5m0-13.5L21 7.5m0 0L16.5 12M21 7.5H7.5"
						/>
					</svg>
				</div>
				<div>
					<p class="mb-0.5 text-sm font-semibold">Доминиращ доставчик</p>
					<p class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">
						Dominance
					</p>
				</div>
			</div>
			<p class="mb-3 text-sm leading-relaxed text-base-content/60">
				Един доставчик печели <strong class="text-base-content">≥ 80%</strong> от всички договори на даден
				купувач при минимум 5 спечелени — признак на зависимост или скрито договаряне.
			</p>
			<div class="space-y-1">
				<div class="flex items-center justify-between rounded-sm bg-base-200/60 px-3 py-1.5">
					<span class="font-mono text-[10px] text-base-content/50 uppercase">Дял по брой</span>
					<span class="font-mono text-[10px] font-medium">≥ 80%</span>
				</div>
				<div class="flex items-center justify-between rounded-sm bg-base-200/60 px-3 py-1.5">
					<span class="font-mono text-[10px] text-base-content/50 uppercase">Мин. договори</span>
					<span class="font-mono text-[10px] font-medium">≥ 5</span>
				</div>
			</div>
		</div>

		<div class="bg-base-100 p-6 lg:p-7">
			<div class="mb-4 flex items-start gap-3">
				<div class="mt-0.5 shrink-0 text-[#B85C38]">
					<svg
						xmlns="http://www.w3.org/2000/svg"
						fill="none"
						viewBox="0 0 24 24"
						stroke-width="1.5"
						stroke="currentColor"
						class="h-4 w-4"
					>
						<path
							stroke-linecap="round"
							stroke-linejoin="round"
							d="M16.023 9.348h4.992v-.001M2.985 19.644v-4.992m0 0h4.992m-4.993 0 3.181 3.183a8.25 8.25 0 0 0 13.803-3.7M4.031 9.865a8.25 8.25 0 0 1 13.803-3.7l3.181 3.182m0-4.991v4.99"
						/>
					</svg>
				</div>
				<div>
					<p class="mb-0.5 text-sm font-semibold">Системно повторение</p>
					<p class="font-mono text-[10px] tracking-wider text-base-content/40 uppercase">
						Repeated award
					</p>
				</div>
			</div>
			<p class="mb-3 text-sm leading-relaxed text-base-content/60">
				Двойка купувач–доставчик с трайна шарка: многогодишна история, множество договора и <strong
					class="text-base-content">≥ 50% без конкуренция</strong
				>.
			</p>
			<div class="space-y-1">
				<div class="flex items-center justify-between rounded-sm bg-base-200/60 px-3 py-1.5">
					<span class="font-mono text-[10px] text-base-content/50 uppercase">Активни години</span>
					<span class="font-mono text-[10px] font-medium">≥ 2</span>
				</div>
				<div class="flex items-center justify-between rounded-sm bg-base-200/60 px-3 py-1.5">
					<span class="font-mono text-[10px] text-base-content/50 uppercase">Общо договори</span>
					<span class="font-mono text-[10px] font-medium">≥ 3</span>
				</div>
				<div class="flex items-center justify-between rounded-sm bg-base-200/60 px-3 py-1.5">
					<span class="font-mono text-[10px] text-base-content/50 uppercase">Без конкуренция</span>
					<span class="font-mono text-[10px] font-medium">≥ 50%</span>
				</div>
			</div>
		</div>
	</div>
</section>

<!-- CTA -->
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
	.font-display {
		font-family: 'Playfair Display', Georgia, serif;
	}
</style>
