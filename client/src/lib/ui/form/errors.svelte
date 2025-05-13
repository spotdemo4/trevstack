<script lang="ts">
	import type { Violation } from '@bufbuild/protovalidate';
	import type { ConnectError } from '@connectrpc/connect';
	import type { WithElementRef, WithoutChildren } from 'bits-ui';
	import type { HTMLAttributes } from 'svelte/elements';
	import { cn } from '$lib/utils';
	import { getFormContext } from './context.svelte';

	type Props = WithoutChildren<WithElementRef<HTMLAttributes<HTMLDivElement>>> & {
		errors?: Violation[] | ConnectError;
	};
	let {
		ref = $bindable(null),
		class: className,
		errors = $bindable(),
		...restProps
	}: Props = $props();

	const item = getFormContext();
</script>

<div bind:this={ref} class={cn('text-red text-sm', className)} {...restProps}>
	{#if errors && Array.isArray(errors)}
		{#each errors as error (error)}
			<label for={item?.id}>{error.message}</label>
		{/each}
	{:else if errors}
		<span>{errors.message}</span>
	{/if}
</div>
