<script lang="ts">
	import { cn } from '$lib/utils';
	import { X } from '@lucide/svelte';

	let {
		name,
		value = $bindable(''),
		type = 'text',
		placeholder,
		className,
		onchange
	}: {
		name?: string;
		value?: string | number;
		type?: string;
		placeholder?: string;
		className?: string;
		onchange?: (e: Event) => void;
	} = $props();
</script>

<div
	class={cn(
		'border-surface-0 hover:border-surface-2 flex items-center justify-between gap-1 rounded border p-0 drop-shadow-md transition-all',
		className
	)}
>
	<input
		id={name}
		{name}
		{type}
		{placeholder}
		class="focus:outline-sky grow rounded-l p-2 text-sm transition-all focus:outline focus:outline-offset-1"
		bind:value
		{onchange}
	/>
	<button
		class="text-overlay-2 hover:bg-surface-0 focus:outline-sky cursor-pointer rounded-r p-2 transition-all focus:outline focus:outline-offset-1"
		type="button"
		onclick={(e) => {
			if (value) {
				value = '';
				if (onchange) {
					onchange(e);
				}
			}
		}}
	>
		<X size="20" />
	</button>
</div>
