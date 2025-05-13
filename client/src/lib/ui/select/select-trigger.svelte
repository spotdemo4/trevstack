<script lang="ts">
	import type { WithoutChild } from 'bits-ui';
	import ChevronDown from '@lucide/svelte/icons/chevron-down';
	import { cn } from '$lib/utils.js';
	import { Select as SelectPrimitive } from 'bits-ui';

	let {
		ref = $bindable(null),
		class: className,
		children,
		size = 'default',
		...restProps
	}: WithoutChild<SelectPrimitive.TriggerProps> & {
		size?: 'sm' | 'default';
	} = $props();
</script>

<SelectPrimitive.Trigger
	bind:ref
	data-slot="select-trigger"
	data-size={size}
	class={cn(
		'border-surface-1 bg-based hover:border-overlay text-text flex w-full cursor-pointer flex-row-reverse items-center justify-between gap-2 rounded-md border px-3 py-2 whitespace-nowrap shadow-xs transition-all md:text-sm',

		// Focus
		'focus-visible:outline-accent focus-visible:outline-2 focus-visible:outline-offset-2',

		// Disabled
		'disabled:cursor-not-allowed disabled:opacity-50',

		// Data
		'data-[size=default]:h-9 data-[size=sm]:h-8 *:data-[slot=select-value]:line-clamp-1 *:data-[slot=select-value]:flex *:data-[slot=select-value]:items-center *:data-[slot=select-value]:gap-2',

		// Image
		"[&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
		className
	)}
	{...restProps}
>
	<ChevronDown class="size-4 opacity-50" />
	{@render children?.()}
</SelectPrimitive.Trigger>
