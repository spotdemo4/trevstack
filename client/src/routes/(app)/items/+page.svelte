<script lang="ts">
	import { ItemClient } from '$lib/transport';
	import { Plus, Trash, Pencil } from '@lucide/svelte';
	import { timestampFromDate, timestampDate } from '@bufbuild/protobuf/wkt';
	import { toast } from 'svelte-sonner';
	import { ConnectError } from '@connectrpc/connect';
	import Modal from '$lib/ui/Modal.svelte';
	import Button from '$lib/ui/Button.svelte';
	import DateRangePicker from '$lib/ui/DateRangePicker.svelte';
	import Input from '$lib/ui/Input.svelte';
	import Select from '$lib/ui/Select.svelte';
	import { SvelteMap } from 'svelte/reactivity';
	import type { Item } from '$lib/services/item/v1/item_pb';
	import Pagination from '$lib/ui/Pagination.svelte';

	// Config
	let limit: number = $state(10);
	let offset: number = $state(0);
	let start: Date | undefined = $state();
	let end: Date | undefined = $state();
	let filter = $state('');

	// Items
	let items = $state(getItems());
	let count: number = $state(0);

	// Open
	let addedOpen = $state(false);
	let deletesOpen: SvelteMap<number, boolean> = new SvelteMap();
	let editsOpen: SvelteMap<number, boolean> = new SvelteMap();

	async function getItems() {
		return await ItemClient.getItems({
			limit: limit,
			offset: offset,
			start: start ? timestampFromDate(start) : undefined,
			end: end ? timestampFromDate(end) : undefined,
			filter: filter
		}).then((resp) => {
			count = Number(resp.count);
			return resp.items;
		});
	}
	async function updateItems() {
		let i = getItems();
		i.then(() => {
			items = i;
		});
	}
</script>

<div class="mx-4 my-2 flex flex-wrap items-center justify-center gap-2">
	<Input bind:value={filter} className="bg-mantle" placeholder="Filter" onchange={updateItems} />
	<Select
		items={[
			{
				label: '10 Items',
				value: '10'
			},
			{
				label: '25 Items',
				value: '25'
			},
			{
				label: '100 Items',
				value: '100'
			},
			{
				label: '250 Items',
				value: '250'
			}
		]}
		placeholder="Items per page"
		onchange={() => {
			offset = 0;
			updateItems();
		}}
		defaultValue="10"
		bind:value={() => limit.toString(), (v) => (limit = parseInt(v))}
	/>
	<DateRangePicker bind:start bind:end onchange={updateItems} />
</div>

{#snippet editModal(item: Item)}
	<Modal
		bind:open={
			() =>
				editsOpen.has(item.id!)
					? editsOpen.get(item.id!)!
					: editsOpen.set(item.id!, false) && editsOpen.get(item.id!)!,
			(value) => editsOpen.set(item.id!, value)
		}
	>
		{#snippet trigger(props)}
			<Button {...props} className="bg-text">
				<Pencil />
			</Button>
		{/snippet}

		{#snippet title()}
			Edit '{item.name}'
		{/snippet}

		{#snippet content()}
			<form
				onsubmit={async (e) => {
					e.preventDefault();
					const form = e.target as HTMLFormElement;
					const formData = new FormData(form);
					const name = formData.get('name')?.toString();
					const description = formData.get('description')?.toString();
					const price = formData.get('price')?.toString();
					const quantity = formData.get('quantity')?.toString();

					try {
						const response = await ItemClient.updateItem({
							item: {
								id: item.id,
								name: name,
								description: description,
								price: parseFloat(price ?? '0'),
								quantity: parseInt(quantity ?? '0')
							}
						});

						if (response.item && item.id) {
							toast.success(`item "${name}" saved`);
							editsOpen.set(item.id, false);
							await updateItems();
						}
					} catch (err) {
						const error = ConnectError.from(err);
						toast.error(error.rawMessage);
					}
				}}
			>
				<div class="flex flex-col gap-4 p-3">
					<div class="flex flex-col gap-1">
						<label for="name" class="text-sm">Name</label>
						<Input name="name" type="text" value={item.name} />
					</div>
					<div class="flex flex-col gap-1">
						<label for="description" class="text-sm">Description</label>
						<Input name="description" type="text" value={item.description} />
					</div>
					<div class="flex flex-col gap-1">
						<label for="price" class="text-sm">Price</label>
						<Input name="price" type="number" value={item.price} />
					</div>
					<div class="flex flex-col gap-1">
						<label for="quantity" class="text-sm">Quantity</label>
						<Input name="quantity" type="number" value={item.quantity} />
					</div>
					<Button type="submit">Submit</Button>
				</div>
			</form>
		{/snippet}
	</Modal>
{/snippet}

{#snippet deleteModal(item: Item)}
	<Modal
		bind:open={
			() =>
				deletesOpen.has(item.id!)
					? deletesOpen.get(item.id!)!
					: deletesOpen.set(item.id!, false) && deletesOpen.get(item.id!)!,
			(value) => deletesOpen.set(item.id!, value)
		}
	>
		{#snippet trigger(props)}
			<Button {...props} className="bg-red">
				<Trash />
			</Button>
		{/snippet}

		{#snippet title()}
			Delete '{item.name}'
		{/snippet}

		{#snippet content()}
			<form
				onsubmit={async (e) => {
					e.preventDefault();

					try {
						await ItemClient.deleteItem({
							id: item.id
						});

						toast.success(`item "${item.name}" deleted`);
						deletesOpen.set(item.id!, false);
						await updateItems();
					} catch (err) {
						const error = ConnectError.from(err);
						toast.error(error.rawMessage);
					}
				}}
			>
				<div class="flex flex-col gap-4 p-3">
					<span class="text-center">Are you sure you want to delete "{item.name}"?</span>
					<div class="flex justify-center gap-4">
						<Button type="submit">Submit</Button>
					</div>
				</div>
			</form>
		{/snippet}
	</Modal>
{/snippet}

<div
	class="border-surface-0 bg-mantle mx-4 my-2 hidden overflow-x-auto rounded border-x border-t drop-shadow-md sm:block"
>
	<table class="w-full table-auto border-collapse text-left rtl:text-right">
		<thead>
			<tr class="border-surface-0 border-b">
				<th scope="col" class="text-subtext-0 px-6 py-3 font-normal">Added</th>
				<th scope="col" class="text-subtext-0 px-6 py-3 font-normal">Name</th>
				<th scope="col" class="text-subtext-0 px-6 py-3 font-normal">Description</th>
				<th scope="col" class="text-subtext-0 px-6 py-3 font-normal">Price</th>
				<th scope="col" class="text-subtext-0 px-6 py-3 font-normal">Quantity</th>
				<th class="w-0"></th>
			</tr>
		</thead>
		<tbody>
			{#await items}
				<tr class="border-surface-0 border-b">
					<td class="px-6 py-3"><div class="bg-surface-2 m-2 h-3 animate-pulse rounded"></div></td>
					<td class="px-6 py-3"><div class="bg-surface-2 m-2 h-3 animate-pulse rounded"></div></td>
					<td class="px-6 py-3"><div class="bg-surface-2 m-2 h-3 animate-pulse rounded"></div></td>
					<td class="px-6 py-3"><div class="bg-surface-2 m-2 h-3 animate-pulse rounded"></div></td>
					<td class="px-6 py-3"><div class="bg-surface-2 m-2 h-3 animate-pulse rounded"></div></td>
					<td class="w-8"></td>
				</tr>
			{:then items}
				{#each items as item (item.id)}
					<tr class="border-surface-0 border-b">
						<td class="px-6 py-3">
							{item.added ? timestampDate(item.added).toLocaleString() : ''}
						</td>
						<td class="px-6 py-3">{item.name}</td>
						<td class="px-6 py-3">{item.description}</td>
						<td class="px-6 py-3">${item.price}</td>
						<td class="px-6 py-3">{item.quantity}</td>
						<td class="pr-2">
							<div class="flex gap-2">
								{@render editModal(item)}
								{@render deleteModal(item)}
							</div>
						</td>
					</tr>
				{/each}
			{/await}
		</tbody>
	</table>
</div>

<div class="flex flex-wrap justify-center gap-2 px-4 sm:hidden">
	{#await items}
		<div
			class="border-surface-0 bg-mantle flex w-full flex-wrap gap-6 rounded border p-5 drop-shadow-md"
		>
			<div class="flex flex-col">
				<span class="text-subtext-0 text-sm">Added</span>
				<div class="bg-surface-2 m-2 h-3 animate-pulse rounded"></div>
			</div>
			<div class="flex flex-col">
				<span class="text-subtext-0 text-sm">Name</span>
				<div class="bg-surface-2 m-2 h-3 animate-pulse rounded"></div>
			</div>
			<div class="flex flex-col">
				<span class="text-subtext-0 text-sm">Description</span>
				<div class="bg-surface-2 m-2 h-3 animate-pulse rounded"></div>
			</div>
			<div class="flex flex-col">
				<span class="text-subtext-0 text-sm">Price</span>
				<div class="bg-surface-2 m-2 h-3 animate-pulse rounded"></div>
			</div>
			<div class="flex flex-col">
				<span class="text-subtext-0 text-sm">Quantity</span>
				<div class="bg-surface-2 m-2 h-3 animate-pulse rounded"></div>
			</div>
		</div>
	{:then items}
		{#each items as item (item.id)}
			<div
				class="border-surface-0 bg-mantle flex w-full flex-wrap gap-6 rounded border p-5 drop-shadow-md"
			>
				<div class="flex flex-col">
					<span class="text-subtext-0 text-sm">Added</span>
					<span class="truncate"
						>{item.added ? timestampDate(item.added).toLocaleString() : ''}</span
					>
				</div>
				<div class="flex flex-col">
					<span class="text-subtext-0 text-sm">Name</span>
					<span class="truncate">{item.name}</span>
				</div>
				<div class="flex flex-col">
					<span class="text-subtext-0 text-sm">Description</span>
					<span class="truncate">{item.description}</span>
				</div>
				<div class="flex flex-col">
					<span class="text-subtext-0 text-sm">Price</span>
					<span class="truncate">${item.price}</span>
				</div>
				<div class="flex flex-col">
					<span class="text-subtext-0 text-sm">Quantity</span>
					<span class="truncate">{item.quantity}</span>
				</div>
				<div class="ml-auto flex justify-end gap-2">
					{@render editModal(item)}
					{@render deleteModal(item)}
				</div>
			</div>
		{/each}
	{/await}
</div>

<div class="mx-4 mt-2 mb-4 flex justify-end sm:mt-1">
	<Modal bind:open={addedOpen}>
		{#snippet trigger(props)}
			<Button {...props} className="bg-sky">
				<Plus />
			</Button>
		{/snippet}

		{#snippet title()}
			Add Item
		{/snippet}

		{#snippet content()}
			<form
				onsubmit={async (e) => {
					e.preventDefault();
					const form = e.target as HTMLFormElement;
					const formData = new FormData(form);
					const name = formData.get('name')?.toString();
					const description = formData.get('description')?.toString();
					const price = formData.get('price')?.toString();
					const quantity = formData.get('quantity')?.toString();

					try {
						const response = await ItemClient.createItem({
							item: {
								name: name,
								description: description,
								price: parseFloat(price ?? '0'),
								quantity: parseInt(quantity ?? '0')
							}
						});

						if (response.item) {
							form.reset();
							toast.success(`item "${name}" added`);
							addedOpen = false;
							await updateItems();
						}
					} catch (err) {
						const error = ConnectError.from(err);
						toast.error(error.rawMessage);
					}
				}}
			>
				<div class="flex flex-col gap-4 p-3">
					<div class="flex flex-col gap-1">
						<label for="name" class="text-sm">Name</label>
						<Input name="name" type="text" />
					</div>
					<div class="flex flex-col gap-1">
						<label for="description" class="text-sm">Description</label>
						<Input name="description" type="text" />
					</div>
					<div class="flex flex-col gap-1">
						<label for="price" class="text-sm">Price</label>
						<Input name="price" type="text" />
					</div>
					<div class="flex flex-col gap-1">
						<label for="quantity" class="text-sm">Quantity</label>
						<Input name="quantity" type="text" />
					</div>
					<Button type="submit">Submit</Button>
				</div>
			</form>
		{/snippet}
	</Modal>
</div>

<div class="py-4">
	<Pagination bind:count bind:limit bind:offset onchange={updateItems} />
</div>
