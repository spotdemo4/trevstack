<script lang="ts">
	import { pushState } from '$app/navigation';
	import { Dialog as DialogPrimitive } from 'bits-ui';

	let { open = $bindable(), onOpenChange, ...restProps }: DialogPrimitive.RootProps = $props();
</script>

<svelte:window
	onpopstate={() => {
		if (open) {
			open = false;
		}
	}}
/>

<DialogPrimitive.Root
	onOpenChange={(v) => {
		if (v) {
			pushState('', '#dialog');
		} else {
			history.back();
		}

		onOpenChange?.(v);
	}}
	bind:open
	{...restProps}
/>
