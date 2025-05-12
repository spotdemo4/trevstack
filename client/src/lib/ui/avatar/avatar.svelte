<script lang="ts">
	import { Avatar as AvatarPrimitive } from 'bits-ui';
	import { cn } from '$lib/utils.js';

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
		'outline-surface-1 relative flex size-9 shrink-0 overflow-hidden rounded-full outline outline-offset-2',
		className
	)}
	{...restProps}
/>
