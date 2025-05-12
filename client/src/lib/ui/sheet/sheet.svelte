<script lang="ts">
	import { pushState } from '$app/navigation';
	import { Dialog as SheetPrimitive } from 'bits-ui';

	let { open = $bindable(), onOpenChange, ...restProps }: SheetPrimitive.RootProps = $props();
</script>

<svelte:window
	onpopstate={() => {
		if (open) {
			open = false;
		}
	}}
/>

<SheetPrimitive.Root
	onOpenChange={(v) => {
		if (v) {
			pushState('', '#sheet');
		} else {
			history.back();
		}

		onOpenChange?.(v);
	}}
	bind:open
	{...restProps}
/>
