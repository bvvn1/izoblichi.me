export function oneOf<T>(value: string | null, options: T[], fallback: T): T {
	return options.includes(value as T) ? (value as T) : fallback;
}
