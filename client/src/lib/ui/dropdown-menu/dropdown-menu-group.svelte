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
	class={cn('border-b border-surface first:pt-0 last:pb-0 last:border-none', className)}
	{...restProps}
/>
