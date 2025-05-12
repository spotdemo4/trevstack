<script lang="ts">
	import { cn } from '$lib/utils';
	import { DropdownMenu as DropdownMenuPrimitive } from 'bits-ui';

	let {
		ref = $bindable(null),
		class: className,
		...restProps
	}: DropdownMenuPrimitive.GroupProps = $props();
</script>

<DropdownMenuPrimitive.Group
	bind:ref
	data-slot="dropdown-menu-group"
	class={cn('border-surface border-b first:pt-0 last:border-none last:pb-0', className)}
	{...restProps}
/>
