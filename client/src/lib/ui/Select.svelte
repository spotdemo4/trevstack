<script lang="ts">
	import { cn } from '$lib/utils';
	import { Check, ChevronsDown, ChevronsUp, ChevronsUpDown, X } from '@lucide/svelte';
	import { Select } from 'bits-ui';
	import { fade } from 'svelte/transition';

	let {
		value = $bindable('10'),
		placeholder = 'Select an item',
		items = [],
		defaultValue = '',
		className,
		onchange
	}: {
		value?: string;
		placeholder?: string;
		items: { value: string; label: string; disabled?: boolean }[];
		defaultValue?: string;
		className?: string;
		onchange?: (e: string) => void;
	} = $props();

	const selectedLabel = $derived(value ? items.find((i) => i.value === value)?.label : placeholder);
</script>

<div
	class={cn(
		'border-surface-0 bg-mantle hover:border-surface-2 flex items-center justify-between rounded border p-0 drop-shadow-md transition-all',
		className
	)}
>
	<Select.Root type="single" {items} bind:value onValueChange={onchange}>
		<Select.Trigger
			class="focus:outline-sky data-placeholder:text-overlay-0 inline-flex grow cursor-pointer items-center justify-between gap-2 rounded-l py-2 pl-2 text-sm transition-colors select-none focus:outline focus:outline-offset-1"
			aria-label={placeholder}
		>
			{selectedLabel}
			<ChevronsUpDown class="text-overlay-0" size="20" />
		</Select.Trigger>
		<Select.Portal>
			<Select.Content forceMount>
				{#snippet child({ wrapperProps, props, open })}
					{#if open}
						<div {...wrapperProps}>
							<div
								{...props}
								class="border-surface-0 bg-mantle shadow-popover z-50 mt-1 rounded border p-1 outline-hidden select-none"
								transition:fade={{
									duration: 100
								}}
							>
								<Select.ScrollUpButton class="flex w-full items-center justify-center">
									<ChevronsUp size="20" />
								</Select.ScrollUpButton>
								<Select.Viewport class="p-1">
									{#each items as item, i (i + item.value)}
										<Select.Item
											class="data-highlighted:bg-surface-0 flex h-10 w-full cursor-pointer items-center gap-4 rounded px-5 py-3 text-sm capitalize outline-hidden select-none data-disabled:cursor-not-allowed data-disabled:opacity-50"
											value={item.value}
											label={item.label}
											disabled={item.disabled}
										>
											{#snippet children({ selected })}
												{item.label}
												{#if selected}
													<div class="ml-auto">
														<Check size="20" />
													</div>
												{/if}
											{/snippet}
										</Select.Item>
									{/each}
								</Select.Viewport>
								<Select.ScrollDownButton class="flex w-full items-center justify-center">
									<ChevronsDown size="20" />
								</Select.ScrollDownButton>
							</div>
						</div>
					{/if}
				{/snippet}
			</Select.Content>
		</Select.Portal>
	</Select.Root>
	<button
		class="text-overlay-2 hover:bg-surface-0 focus:outline-sky cursor-pointer rounded-r p-2 transition-all focus:outline focus:outline-offset-1"
		type="button"
		onclick={() => {
			value = defaultValue;
			onchange?.(value);
		}}
	>
		<X size="20" />
	</button>
</div>
