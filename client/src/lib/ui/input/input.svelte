<script lang="ts">
	import type { WithElementRef } from 'bits-ui';
	import type { HTMLInputAttributes, HTMLInputTypeAttribute } from 'svelte/elements';
	import { cn } from '$lib/utils.js';
	import { getFormContext } from '../form/context.svelte';

	type InputType = Exclude<HTMLInputTypeAttribute, 'file'>;

	type Props = WithElementRef<
		Omit<HTMLInputAttributes, 'type'> &
			({ type: 'file'; files?: FileList } | { type?: InputType; files?: undefined })
	>;

	let {
		ref = $bindable(null),
		value = $bindable(),
		type,
		files = $bindable(),
		class: className,
		id,
		name,
		...restProps
	}: Props = $props();

	const item = getFormContext();
	if (item && !id) {
		id = item.id;
	}
	if (item && !name) {
		name = item.name;
	}
</script>

{#if type === 'file'}
	<input
		{id}
		{name}
		bind:this={ref}
		class={cn(
			'border-surface-1 file:bg-surface hover:border-overlay placeholder:text-subtext text-text flex h-9 w-full min-w-0 cursor-pointer rounded-md border text-sm font-medium shadow-xs transition-all file:mr-2 file:px-3 file:py-2 md:text-sm',

			// Focus
			'focus-visible:outline-accent focus-visible:outline-2 focus-visible:outline-offset-2',

			// Disabled
			'disabled:cursor-not-allowed disabled:opacity-50',

			className
		)}
		type="file"
		bind:files
		bind:value
		{...restProps}
	/>
{:else}
	<input
		{id}
		{name}
		bind:this={ref}
		class={cn(
			'border-surface-1 hover:border-overlay placeholder:text-subtext text-text flex h-9 w-full min-w-0 rounded-md border px-3 py-1 shadow-xs transition-all md:text-sm',

			// Focus
			'focus-visible:outline-accent focus-visible:outline-2 focus-visible:outline-offset-2',

			// Disabled
			'disabled:cursor-not-allowed disabled:opacity-50',

			className
		)}
		{type}
		bind:value
		{...restProps}
	/>
{/if}
