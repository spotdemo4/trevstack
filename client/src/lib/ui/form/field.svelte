<script lang="ts">
	import type { WithElementRef } from 'bits-ui';
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';
	import { setFormContext } from './context.svelte';

	type Props = WithElementRef<HTMLAttributes<HTMLDivElement>> & {
		name?: string;
	};
	let { ref = $bindable(null), class: className, name, children, ...restProps }: Props = $props();

	const uid = $props.id();

	if (name) {
		setFormContext(uid, name);
	}
</script>

<div bind:this={ref} class={cn('flex flex-col gap-1', className)} {...restProps}>
	{@render children?.()}
</div>
