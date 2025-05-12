<script lang="ts">
	import { DropdownMenu as DropdownMenuPrimitive } from 'bits-ui';
	import Item from './dropdown-menu-item.svelte';

	let {
		ref = $bindable(null),
		class: className,
		href,
		children,
		...restProps
	}: DropdownMenuPrimitive.ItemProps & {
		href?: string;
	} = $props();
</script>

<Item bind:ref class={className} {...restProps}>
	{#snippet child({ props })}
		<a {...props} {href}>
			{@render children?.()}
		</a>
	{/snippet}
</Item>
