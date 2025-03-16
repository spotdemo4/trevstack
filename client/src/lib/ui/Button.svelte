<script lang="ts">
	import { Button } from 'bits-ui';
	import { cn } from '$lib/utils';
	import type { MouseEventHandler } from 'svelte/elements';
	import type { Snippet } from 'svelte';

	let {
		className,
		type,
		onclick,
		children
	}: {
		className?: string;
		type?: 'submit' | 'reset' | 'button' | null;
		onclick?: () => MouseEventHandler<HTMLButtonElement> | null | undefined;
		children?: Snippet<[]>;
	} = $props();
</script>

<Button.Root
	{type}
	class={cn(
		'bg-sky text-crust hover:brightness-120 w-fit cursor-pointer rounded p-2 px-4 text-sm font-medium transition-all',
		className
	)}
	{onclick}
>
	{@render children?.()}
</Button.Root>
