<script lang="ts">
	import { cn } from '$lib/utils';
	import { Check, ChevronsDown, ChevronsUp, ChevronsUpDown, X } from '@lucide/svelte';
	import { Select } from 'bits-ui';

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
			class="focus:outline-sky data-placeholder:text-overlay-0 gap-2 inline-flex grow cursor-pointer select-none items-center justify-between rounded-l py-2 pl-2 text-sm transition-colors focus:outline focus:outline-offset-1"
			aria-label={placeholder}
		>
			{selectedLabel}
			<ChevronsUpDown class="text-overlay-0" size="20" />
		</Select.Trigger>
		<Select.Portal>
			<Select.Content
				class="focus-override border-surface-0 bg-mantle shadow-popover data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 outline-hidden z-50 select-none rounded border p-1"
				sideOffset={10}
			>
				<Select.ScrollUpButton class="flex w-full items-center justify-center">
					<ChevronsUp size="20" />
				</Select.ScrollUpButton>
				<Select.Viewport class="p-1">
					{#each items as item, i (i + item.value)}
						<Select.Item
							class="data-disabled:cursor-not-allowed data-highlighted:bg-surface-0 outline-hidden data-disabled:opacity-50 flex h-10 w-full cursor-pointer select-none items-center gap-4 rounded px-5 py-3 text-sm capitalize"
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
