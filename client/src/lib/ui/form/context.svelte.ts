import { getContext, hasContext, setContext } from 'svelte';

type Item = {
	id: string;
	name: string;
};

const key = 'form';
export function setFormContext(id: string, name: string) {
	const item = getFormContext();
	if (!item) {
		const item: Item = $state({
			id,
			name
		});
		setContext(key, item);

		return;
	}

	item.id = id;
	item.name = name;
}

export function getFormContext() {
	if (!hasContext(key)) {
		return null;
	}

	return getContext(key) as Item;
}
