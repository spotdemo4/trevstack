<script lang="ts">
	import type { WithElementRef } from 'bits-ui';
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';
	import { getFormContext } from './context.svelte';

	type Props = WithElementRef<HTMLAttributes<HTMLLabelElement>>;
	let { ref = $bindable(null), class: className, children, ...restProps }: Props = $props();

	const item = getFormContext();

	function formatName(name: string) {
		// Replace _ with spaces
		name = name.replace('_', ' ');

		// Capitalize first letter
		name = name.charAt(0).toUpperCase() + name.slice(1);

		return name;
	}
</script>

<label bind:this={ref} class={cn('text-sm', className)} for={item?.id} {...restProps}>
	{#if children}
		{@render children()}
	{:else if item}
		{formatName(item.name)}
	{/if}
</label>
