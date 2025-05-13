<script lang="ts">
	import type { Item } from '$lib/connect/item/v1/item_pb';
	import { Button } from '$lib/ui/button';
	import { Card } from '$lib/ui/card';
	import { DateRangePicker } from '$lib/ui/daterangepicker';
	import * as Dialog from '$lib/ui/dialog';
	import * as Form from '$lib/ui/form';
	import { Input } from '$lib/ui/input';
	import { Pager } from '$lib/ui/pager';
	import * as Select from '$lib/ui/select';
	import { Skeleton } from '$lib/ui/skeleton';
	import * as Table from '$lib/ui/table';
	import { timestampDate, timestampFromDate } from '@bufbuild/protobuf/wkt';
	import { LoaderCircle, Pencil, Plus, Trash } from '@lucide/svelte';
	import { ItemService } from '$lib/connect/item/v1/item_pb';
	import { coolForm, newState } from '$lib/coolforms';
	import { ItemClient } from '$lib/transport';
	import { cn } from '$lib/utils';

	const get = coolForm(ItemClient, ItemService.method.getItems, {
		init: {
			limit: 10,
			offset: 0
		},
		start: true,
		reset: false
	});
</script>

<div class="mx-4 my-2 flex flex-wrap items-center justify-center gap-2">
	<Input
		bind:value={get.input.filter}
		class="bg-based max-w-sm"
		placeholder="Filter"
		onchange={() => {
			get.submit();
		}}
	/>

	<Select.Root
		type="single"
		bind:value={() => get.input.limit?.toString(), (limit) => (get.input.limit = Number(limit))}
		onValueChange={() => {
			get.submit();
		}}
	>
		<Select.Trigger class="w-32">
			{get.input.limit} Items
		</Select.Trigger>
		<Select.Content>
			<Select.Item value="10">10 Items</Select.Item>
			<Select.Item value="25">25 Items</Select.Item>
			<Select.Item value="100">100 Items</Select.Item>
			<Select.Item value="250">250 Items</Select.Item>
		</Select.Content>
	</Select.Root>

	<DateRangePicker
		start={get.input.start ? timestampDate(get.input.start) : undefined}
		end={get.input.end ? timestampDate(get.input.end) : undefined}
		onchange={(start, end) => {
			get.input.start = timestampFromDate(start);
			get.input.end = timestampFromDate(end);
			get.submit();
		}}
	/>

	<LoaderCircle class={cn('invisible animate-spin', get.loading() && 'visible')} />
</div>

{#snippet editModal(item: Item)}
	{@const s = newState({ open: false })}
	{@const update = coolForm(ItemClient, ItemService.method.updateItem, {
		init: {
			id: item.id,
			name: item.name,
			description: item.description,
			price: item.price,
			quantity: item.quantity
		},
		onResult: () => {
			get.submit();
			s.open = false;
		}
	})}

	<Dialog.Root bind:open={s.open}>
		<Dialog.Trigger>
			{#snippet child({ props })}
				<Button {...props}>
					<Pencil />
				</Button>
			{/snippet}
		</Dialog.Trigger>
		<Dialog.Content>
			<Dialog.Header>
				<Dialog.Title>Edit '{item.name}'</Dialog.Title>
			</Dialog.Header>
			<form use:update.impair class="flex flex-col gap-2">
				<Input type="hidden" bind:value={update.input.id} />

				<Form.Field name="name">
					<Form.Label />
					<Input bind:value={update.input.name} />
					<Form.Errors bind:errors={update.errors.name} />
				</Form.Field>

				<Form.Field name="description">
					<Form.Label />
					<Input bind:value={update.input.description} />
					<Form.Errors bind:errors={update.errors.description} />
				</Form.Field>

				<Form.Field name="price">
					<Form.Label />
					<Input type="number" bind:value={update.input.price} />
					<Form.Errors bind:errors={update.errors.price} />
				</Form.Field>

				<Form.Field name="quantity">
					<Form.Label />
					<Input type="number" bind:value={update.input.quantity} />
					<Form.Errors bind:errors={update.errors.quantity} />
				</Form.Field>

				<Form.Errors bind:errors={update.errors.form} />

				<Button type="submit" loading={update.loading()}>Submit</Button>
			</form>
		</Dialog.Content>
	</Dialog.Root>
{/snippet}

{#snippet deleteModal(item: Item)}
	{@const s = newState({ open: false })}
	{@const remove = coolForm(ItemClient, ItemService.method.deleteItem, {
		init: {
			id: item.id
		},
		onResult: () => {
			get.submit();
			s.open = false;
		}
	})}

	<Dialog.Root bind:open={s.open}>
		<Dialog.Trigger>
			{#snippet child({ props })}
				<Button variant="red" {...props}>
					<Trash />
				</Button>
			{/snippet}
		</Dialog.Trigger>
		<Dialog.Content>
			<Dialog.Header>
				<Dialog.Title>Delete '{item.name}'</Dialog.Title>
				<Dialog.Description>Are you sure you want to delete {item.name}?</Dialog.Description>
			</Dialog.Header>
			<form use:remove.impair class="flex flex-col gap-2">
				<Input type="hidden" bind:value={remove.input.id} />

				<Button type="submit" loading={remove.loading()}>Delete</Button>
			</form>
		</Dialog.Content>
	</Dialog.Root>
{/snippet}

{#snippet createModal()}
	{@const s = newState({ open: false })}
	{@const create = coolForm(ItemClient, ItemService.method.createItem, {
		onResult: () => {
			get.submit();
			s.open = false;
		}
	})}

	<Dialog.Root bind:open={s.open}>
		<Dialog.Trigger>
			{#snippet child({ props })}
				<Button class="w-14" {...props}>
					<Plus />
				</Button>
			{/snippet}
		</Dialog.Trigger>
		<Dialog.Content>
			<Dialog.Header>
				<Dialog.Title>Add item</Dialog.Title>
			</Dialog.Header>
			<form use:create.impair class="flex flex-col gap-2">
				<Form.Field name="name">
					<Form.Label />
					<Input bind:value={create.input.name} />
					<Form.Errors bind:errors={create.errors.name} />
				</Form.Field>

				<Form.Field name="description">
					<Form.Label />
					<Input bind:value={create.input.description} />
					<Form.Errors bind:errors={create.errors.description} />
				</Form.Field>

				<Form.Field name="price">
					<Form.Label />
					<Input type="number" bind:value={create.input.price} />
					<Form.Errors bind:errors={create.errors.price} />
				</Form.Field>

				<Form.Field name="quantity">
					<Form.Label />
					<Input type="number" bind:value={create.input.quantity} />
					<Form.Errors bind:errors={create.errors.quantity} />
				</Form.Field>

				<Form.Errors bind:errors={create.errors.form} />

				<Button type="submit" loading={create.loading()}>Submit</Button>
			</form>
		</Dialog.Content>
	</Dialog.Root>
{/snippet}

<div class="hidden px-4 sm:block">
	<Table.Root>
		<Table.Header>
			<Table.Row>
				<Table.Head>Added</Table.Head>
				<Table.Head>Name</Table.Head>
				<Table.Head>Description</Table.Head>
				<Table.Head>Price</Table.Head>
				<Table.Head>Quantity</Table.Head>
				<Table.Head></Table.Head>
			</Table.Row>
		</Table.Header>
		<Table.Body>
			{#if get.loading() && get.output.items.length == 0}
				<Table.Row>
					<Table.Cell><Skeleton class="h-5 w-full" /></Table.Cell>
					<Table.Cell><Skeleton class="h-5 w-full" /></Table.Cell>
					<Table.Cell><Skeleton class="h-5 w-full" /></Table.Cell>
					<Table.Cell><Skeleton class="h-5 w-full" /></Table.Cell>
					<Table.Cell><Skeleton class="h-5 w-full" /></Table.Cell>
					<Table.Cell class="w-0"></Table.Cell>
				</Table.Row>
			{:else}
				{#each get.output.items as item (item.id)}
					<Table.Row>
						<Table.Cell>{item.added ? timestampDate(item.added).toLocaleString() : ''}</Table.Cell>
						<Table.Cell>{item.name}</Table.Cell>
						<Table.Cell>{item.description}</Table.Cell>
						<Table.Cell>{item.price}</Table.Cell>
						<Table.Cell>{item.quantity}</Table.Cell>
						<Table.Cell class="w-0">
							{@render editModal(item)}
							{@render deleteModal(item)}
						</Table.Cell>
					</Table.Row>
				{/each}
			{/if}
		</Table.Body>
	</Table.Root>
</div>

<div class="flex flex-wrap justify-center gap-2 px-4 sm:hidden">
	{#if get.output.items.length > 0}
		{#each get.output.items as item (item.id)}
			<Card class="flex flex-wrap gap-4">
				<div class="flex flex-col">
					<span class="text-subtext text-sm">Added</span>
					<span class="truncate"
						>{item.added ? timestampDate(item.added).toLocaleString() : ''}</span
					>
				</div>
				<div class="flex flex-col">
					<span class="text-subtext text-sm">Name</span>
					<span class="truncate">{item.name}</span>
				</div>
				<div class="flex flex-col">
					<span class="text-subtext text-sm">Description</span>
					<span class="truncate">{item.description}</span>
				</div>
				<div class="flex flex-col">
					<span class="text-subtext text-sm">Price</span>
					<span class="truncate">${item.price}</span>
				</div>
				<div class="flex flex-col">
					<span class="text-subtext text-sm">Quantity</span>
					<span class="truncate">{item.quantity}</span>
				</div>
				<div class="ml-auto flex justify-end gap-1 self-end">
					{@render editModal(item)}
					{@render deleteModal(item)}
				</div>
			</Card>
		{/each}
	{/if}
</div>

<div class="mx-4 my-2 flex justify-end">
	{@render createModal()}
</div>

<div class="py-4">
	<Pager
		count={Number(get.output.count)}
		limit={get.input.limit}
		bind:offset={get.input.offset}
		onsubmit={() => {
			get.submit();
		}}
	/>
</div>
