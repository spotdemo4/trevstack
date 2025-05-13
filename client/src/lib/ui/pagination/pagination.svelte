<script lang="ts">
	import { pushState } from '$app/navigation';
	import { cn } from '$lib/utils.js';
	import { Pagination as PaginationPrimitive } from 'bits-ui';

	let {
		ref = $bindable(null),
		class: className,
		count = 0,
		perPage = 10,
		page = $bindable(1),
		siblingCount = 1,
		onPageChange,
		...restProps
	}: PaginationPrimitive.RootProps = $props();
</script>

<svelte:window
	onpopstate={(state) => {
		const sks = state.state['sveltekit:states'] as object | string;
		if (typeof sks === 'string' && sks.includes('#pagination-')) {
			page = Number(sks.split('#pagination-')[1]);
			onPageChange?.(page);
		} else {
			page = 1;
			onPageChange?.(page);
		}
	}}
/>

<PaginationPrimitive.Root
	bind:ref
	bind:page
	onPageChange={(p) => {
		pushState('', `#pagination-${p}`);
		onPageChange?.(p);
	}}
	role="navigation"
	aria-label="pagination"
	data-slot="pagination"
	class={cn('mx-auto flex w-full justify-center', className)}
	{count}
	{perPage}
	{siblingCount}
	{...restProps}
/>
