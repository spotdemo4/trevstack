<script lang="ts">
	import { cn } from '$lib/utils.js';
	import { Avatar as AvatarPrimitive } from 'bits-ui';

	let {
		ref = $bindable(null),
		class: className,
		...restProps
	}: AvatarPrimitive.RootProps = $props();
</script>

<AvatarPrimitive.Root
	bind:ref
	data-slot="avatar"
	class={cn(
		'outline-surface-1 relative flex size-9 shrink-0 overflow-hidden rounded-full shadow-xs outline outline-offset-2',

		className
	)}
	{...restProps}
/>
