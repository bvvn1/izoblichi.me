const BGN_TO_EUR = 1.95583;

const formatter = new Intl.NumberFormat('bg-BG', {
	notation: 'compact',
	maximumFractionDigits: 1
});
const eurFormatter = new Intl.NumberFormat('bg-BG', {
	style: 'currency',
	currency: 'EUR',
	minimumFractionDigits: 0,
	maximumFractionDigits: 0
});
const percentFormatter = new Intl.NumberFormat('bg-BG', {
	style: 'percent',
	minimumFractionDigits: 1,
	maximumFractionDigits: 1
});

export const format = (n: number | null) => (n == null ? '—' : formatter.format(n));
export const formatCurrency = (n: number | null, currency: string | null) => {
	if (n == null) return '—';
	if (currency === 'BGN') return eurFormatter.format(n / BGN_TO_EUR);
	if (currency === 'EUR') return eurFormatter.format(n);
	return formatter.format(n) + ' ' + (currency || '');
};
export const formatDate = (d: string | null) => {
	if (!d) return '—';
	const date = new Date(d);
	return date.toLocaleDateString('bg-BG', { year: 'numeric', month: 'short', day: 'numeric' });
};
// export const formatDateNoLetters = (d: string | null) => {
// 	if (!d) return '—';
// 	const date = new Date(d);
// 	return date.toLocaleDateString('bg-BG', { year: 'numeric', month: '2-digit', day: '2-digit' });
// };
export const formatPercent = (n: number | null) => {
	if (n == null) return '-';
	return percentFormatter.format(n / 100);
};
