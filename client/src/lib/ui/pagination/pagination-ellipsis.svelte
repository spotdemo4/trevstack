<script lang="ts">
	import type { WithElementRef, WithoutChildren } from 'bits-ui';
	import type { HTMLAttributes } from 'svelte/elements';
	import Ellipsis from '@lucide/svelte/icons/ellipsis';
	import { cn } from '$lib/utils.js';

	let {
		ref = $bindable(null),
		class: className,
		...restProps
	}: WithoutChildren<WithElementRef<HTMLAttributes<HTMLSpanElement>>> = $props();
</script>

<span
	bind:this={ref}
	aria-hidden="true"
	data-slot="pagination-ellipsis"
	class={cn('text-text flex size-9 items-center justify-center', className)}
	{...restProps}
>
	<Ellipsis class="size-4" />
	<span class="sr-only">More pages</span>
</span>
