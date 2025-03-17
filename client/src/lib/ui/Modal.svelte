<script lang="ts">
	import { X } from '@lucide/svelte';
	import { Dialog } from 'bits-ui';
	import { fade } from 'svelte/transition';
	import type { Snippet } from 'svelte';

	let {
		trigger,
		title,
		content,
		open = $bindable(false)
	}: {
		trigger: Snippet<[Record<string, unknown>]>;
		title: Snippet;
		content: Snippet;
		open?: boolean;
	} = $props();
</script>

<Dialog.Root bind:open>
	<Dialog.Trigger>
		{#snippet child({ props })}
			{@render trigger(props)}
		{/snippet}
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
			{#snippet child({ props, open: propopen })}
				{#if propopen}
					<div
						{...props}
						transition:fade={{
							duration: 100
						}}
					>
						<div
							class="bg-mantle border-surface-0 fixed inset-0 left-[50%] top-[50%] z-50 size-fit w-96 -translate-x-1/2 -translate-y-1/2 transform overflow-y-auto rounded-xl border pb-1 drop-shadow-md"
						>
							<div class="border-surface-0 flex justify-between border-b p-2">
								<h1 class="grow truncate p-1 text-center text-xl font-bold">
									{@render title()}
								</h1>
								<button
									tabindex="-1"
									class="text-overlay-2 hover:bg-surface-0 focus:outline-sky cursor-pointer rounded p-1 transition-all focus:outline focus:outline-offset-1"
									onclick={() => {
										open = false;
									}}
								>
									<X />
								</button>
							</div>
							{@render content()}
						</div>
					</div>
				{/if}
			{/snippet}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
