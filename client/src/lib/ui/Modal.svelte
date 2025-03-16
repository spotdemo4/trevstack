<script lang="ts">
	import { Dialog } from 'bits-ui';
	import { fade } from 'svelte/transition';
	import type { Snippet } from 'svelte';

	let {
		trigger,
		content,
		open = $bindable(false)
	}: { trigger: Snippet; content: Snippet; open?: boolean } = $props();
</script>

<Dialog.Root bind:open>
	<Dialog.Trigger>
		{@render trigger()}
	</Dialog.Trigger>
	<Dialog.Portal>
		<Dialog.Overlay forceMount>
			{#snippet child({ props, open })}
				{#if open}
					<div
						{...props}
						transition:fade={{
							duration: 100
						}}
					>
						<div class="fixed inset-0 z-50 mt-[50px] bg-black/50 transition-all"></div>
					</div>
				{/if}
			{/snippet}
		</Dialog.Overlay>
		<Dialog.Content forceMount>
			{#snippet child({ props, open })}
				{#if open}
					<div
						{...props}
						transition:fade={{
							duration: 100
						}}
					>
						<div
							class="bg-mantle border-surface-0 fixed inset-0 left-[50%] top-[50%] z-50 size-fit w-96 -translate-x-1/2 -translate-y-1/2 transform overflow-y-auto rounded-xl border pb-1 drop-shadow-md"
						>
							{@render content()}
						</div>
					</div>
				{/if}
			{/snippet}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
