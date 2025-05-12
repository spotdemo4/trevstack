export function newState<T>(s: T) {
	const state = $state(s);

	return state;
}
