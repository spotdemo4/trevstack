<script lang="ts">
	import type { ToasterProps as SonnerProps } from 'svelte-sonner';
	import { Toaster as Sonner } from 'svelte-sonner';

	let { ...restProps }: SonnerProps = $props();
</script>

<Sonner
	toastOptions={{
		unstyled: true,
		classes: {
			toast:
				'bg-based text-sm flex gap-2 px-4 border border-surface-1 p-2 rounded-md items-center text-text w-full min-h-12',
			title: 'text-text',
			description: 'text-subtext text-xs',
			actionButton: 'bg-accent',
			cancelButton: 'bg-red',
			closeButton: 'bg-green'
		}
	}}
	{...restProps}
/>
