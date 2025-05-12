<script lang="ts">
	import { cn } from '$lib/utils';
	import { DropdownMenu as DropdownMenuPrimitive } from 'bits-ui';

	let {
		ref = $bindable(null),
		class: className,
		...restProps
	}: DropdownMenuPrimitive.TriggerProps = $props();
</script>

<DropdownMenuPrimitive.Trigger
	bind:ref
	data-slot="dropdown-menu-trigger"
	class={cn(
		'flex cursor-pointer items-center gap-1 transition-all',

		// Focus
		'focus-visible:outline-accent focus-visible:outline-2 focus-visible:outline-offset-2',
		className
	)}
	{...restProps}
/>
