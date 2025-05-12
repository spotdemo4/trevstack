<script lang="ts">
	import * as Pagination from '$lib/ui/pagination';

	type Props = {
		count: number;
		limit?: number;
		offset?: number;
		onsubmit?: (offset: number) => void;
	};
	let { count = $bindable(), limit = 10, offset = $bindable(0), onsubmit }: Props = $props();
</script>

<Pagination.Root
	{count}
	page={offset / limit + 1}
	perPage={limit}
	onPageChange={async (num) => {
		offset = num * limit - limit;
		scrollTo({
			top: 0,
			behavior: 'smooth'
		});
		onsubmit?.(offset);
	}}
>
	{#snippet children({ pages, currentPage })}
		<Pagination.Content>
			<Pagination.Item>
				<Pagination.PrevButton />
			</Pagination.Item>
			{#each pages as page (page.key)}
				{#if page.type === 'ellipsis'}
					<Pagination.Item class="hidden md:block">
						<Pagination.Ellipsis />
					</Pagination.Item>
				{:else}
					<Pagination.Item class="hidden md:block">
						<Pagination.Link {page} isActive={currentPage === page.value}>
							{page.value}
						</Pagination.Link>
					</Pagination.Item>
				{/if}
			{/each}
			<Pagination.Item>
				<Pagination.NextButton />
			</Pagination.Item>
		</Pagination.Content>
	{/snippet}
</Pagination.Root>
