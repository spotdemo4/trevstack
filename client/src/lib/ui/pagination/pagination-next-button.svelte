<script lang="ts">
	import { buttonVariants } from '$lib/ui/button/index.js';
	import ChevronRight from '@lucide/svelte/icons/chevron-right';
	import { cn } from '$lib/utils.js';
	import { Pagination as PaginationPrimitive } from 'bits-ui';

	let {
		ref = $bindable(null),
		class: className,
		children,
		...restProps
	}: PaginationPrimitive.NextButtonProps = $props();
</script>

{#snippet Fallback()}
	<span>Next</span>
	<ChevronRight class="size-4" />
{/snippet}

<!-- TODO: Fix this error: Expression produces a union type that is too complex to represent. Note: Removing `Fallback` in children={children || Fallback} fixes, makes you wonder how/why `Fallback` is causing this. -->
<PaginationPrimitive.NextButton
	bind:ref
	aria-label="Go to next page"
	class={cn(
		buttonVariants({
			size: 'default',
			variant: 'ghost',
			class: 'gap-1 px-2.5 sm:pr-2.5'
		}),
		className
	)}
	children={children || Fallback}
	{...restProps}
/>
