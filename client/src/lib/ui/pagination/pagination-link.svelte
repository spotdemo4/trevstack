<script lang="ts">
	import type { Props } from '$lib/ui/button';
	import { buttonVariants } from '$lib/ui/button';
	import { cn } from '$lib/utils.js';
	import { Pagination as PaginationPrimitive } from 'bits-ui';

	let {
		ref = $bindable(null),
		class: className,
		size = 'icon',
		isActive = false,
		page,
		children,
		...restProps
	}: PaginationPrimitive.PageProps &
		Props & {
			isActive: boolean;
		} = $props();
</script>

{#snippet Fallback()}
	{page.value}
{/snippet}

<PaginationPrimitive.Page
	bind:ref
	{page}
	aria-current={isActive ? 'page' : undefined}
	data-slot="pagination-link"
	data-active={isActive}
	class={cn(
		buttonVariants({
			variant: 'ghost',
			size
		}),
		'text-text',
		isActive && 'bg-surface',
		className
	)}
	children={children || Fallback}
	{...restProps}
/>
