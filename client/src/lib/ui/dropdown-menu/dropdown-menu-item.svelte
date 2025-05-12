<script lang="ts">
	import { cn } from '$lib/utils.js';
	import { DropdownMenu as DropdownMenuPrimitive } from 'bits-ui';

	let {
		ref = $bindable(null),
		class: className,
		...restProps
	}: DropdownMenuPrimitive.ItemProps = $props();
</script>

<DropdownMenuPrimitive.Item
	bind:ref
	data-slot="dropdown-menu-item"
	class={cn(
		'focus:bg-surface text-text relative flex cursor-pointer items-center gap-2 px-4 py-2 text-sm outline-hidden transition-all select-none',

		// Disabled
		'data-[disabled]:pointer-events-none data-[disabled]:opacity-50',

		// Images
		"[&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",

		className
	)}
	{...restProps}
/>
