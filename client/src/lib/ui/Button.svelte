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
		onclick?: () => void;
		children?: Snippet<[]>;
	} = $props();
</script>

<Button.Root
	{type}
	class={cn(
		'bg-sky text-crust focus:outline-sky flex w-fit cursor-pointer items-center justify-center rounded p-2 px-4 text-sm font-medium transition-all hover:brightness-120 focus:outline-2 focus:outline-offset-1',
		className
	)}
	onclick={() => {
		onclick?.();
	}}
>
	{@render children?.()}
</Button.Root>
