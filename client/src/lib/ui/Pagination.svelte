<script lang="ts">
	import { cn } from '$lib/utils';
	import { ChevronLeft, ChevronRight } from '@lucide/svelte';
	import { Pagination } from 'bits-ui';
	import { pushState, replaceState } from '$app/navigation';
	import { onMount } from 'svelte';

	let {
		count = $bindable(),
		limit = $bindable(),
		offset = $bindable(0),
		className,
		onchange
	}: {
		count: number;
		limit: number;
		offset?: number;
		className?: string;
		onchange?: (e: number) => void;
	} = $props();

	let page: number = $state(1);

	onMount(() => {
		replaceState('', `${page}`);
	});
</script>

<svelte:window
	onpopstate={(e) => {
		const lastPage: number = Number(e.state['sveltekit:states']);
		if (!isNaN(lastPage)) {
			page = lastPage;
			offset = (lastPage - 1) * limit;
			window.scrollTo(0, 0);
			onchange?.(lastPage);
		}
	}}
/>

{#key count && limit}
	<Pagination.Root
		{count}
		bind:page
		perPage={limit}
		onPageChange={(e) => {
			offset = (e - 1) * limit;
			window.scrollTo(0, 0);
			pushState('', `${e}`);
			onchange?.(e);
		}}
	>
		{#snippet children({ pages, range })}
			<div class={cn('mb-2 flex items-center justify-center gap-2', className)}>
				<Pagination.PrevButton
					class="hover:bg-surface-0 disabled:text-overlay-0 inline-flex cursor-pointer items-center justify-center rounded p-2 transition-all disabled:cursor-not-allowed hover:disabled:bg-transparent"
				>
					<ChevronLeft />
				</Pagination.PrevButton>
				<div class="flex items-center gap-2">
					{#each pages as page (page.key)}
						{#if page.type === 'ellipsis'}
							<div class="font-medium select-none">...</div>
						{:else}
							<Pagination.Page
								{page}
								class="hover:bg-surface-0 data-selected:bg-surface-0 data-selected:text-background inline-flex size-10 cursor-pointer items-center justify-center rounded bg-transparent font-medium transition-all select-none disabled:cursor-not-allowed disabled:opacity-50 hover:disabled:bg-transparent"
							>
								{page.value}
							</Pagination.Page>
						{/if}
					{/each}
				</div>
				<Pagination.NextButton
					class="hover:bg-surface-0 disabled:text-overlay-0 inline-flex cursor-pointer items-center justify-center rounded p-2 transition-all disabled:cursor-not-allowed hover:disabled:bg-transparent"
				>
					<ChevronRight />
				</Pagination.NextButton>
			</div>
			<p class="text-overlay-2 text-center text-sm">
				Showing {range.start} - {range.end}
			</p>
		{/snippet}
	</Pagination.Root>
{/key}
