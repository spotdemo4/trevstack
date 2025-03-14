<script lang="ts">
	import { ItemClient } from '$lib/transport';
	import { Plus, Trash, Pencil } from '@lucide/svelte';
	import { timestampFromDate, timestampDate } from '@bufbuild/protobuf/wkt';
	import { Dialog, Button } from 'bits-ui';
	import { fade } from 'svelte/transition';
	import { toast } from 'svelte-sonner';
	import { ConnectError } from '@connectrpc/connect';
	import Modal from '$lib/ui/Modal.svelte';
	import { SvelteMap } from 'svelte/reactivity';

	// Config
	let limit: number = $state(10);
	let offset: number = $state(0);
	let start = $state(new Date(new Date().setDate(new Date().getDate() - 1)));
	let end = $state(new Date());
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
			start: timestampFromDate(start),
			end: timestampFromDate(end),
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

<div
	class="border-surface-0 bg-mantle mx-4 mt-2 overflow-x-auto rounded border-x border-t drop-shadow-md"
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
				</tr>
			{:then items}
				{#each items as item}
					<tr class="border-surface-0 border-b">
						<td class="px-6 py-3">
							{item.added ? timestampDate(item.added).toLocaleString() : ''}
						</td>
						<td class="px-6 py-3">{item.name}</td>
						<td class="px-6 py-3">{item.description}</td>
						<td class="px-6 py-3">{item.price}</td>
						<td class="px-6 py-3">{item.quantity}</td>
						<td class="pr-2">
							<div class="flex gap-2">
								<Modal bind:open={
                                    () => editsOpen.has(item.id!) ? editsOpen.get(item.id!)! : editsOpen.set(item.id!, false) && editsOpen.get(item.id!)!,
                                    (value) => editsOpen.set(item.id!, value)
                                }>
									{#snippet trigger()}
										<button
											class="bg-text text-crust hover:brightness-120 block cursor-pointer rounded p-2 drop-shadow-md"
										>
											<Pencil />
										</button>
									{/snippet}

									{#snippet content()}
										<h1 class="border-surface-0 border-b py-3 text-center text-xl font-bold">
											Edit {item.name}
										</h1>
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
                                                        editsOpen.set(item.id, false)
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
													<input
														id="name"
														name="name"
														type="text"
														class="border-surface-0 rounded border p-2 text-sm"
														value={item.name}
													/>
												</div>
												<div class="flex flex-col gap-1">
													<label for="description" class="text-sm">Description</label>
													<input
														id="description"
														name="description"
														type="text"
														class="border-surface-0 rounded border p-2 text-sm"
														value={item.description}
													/>
												</div>
												<div class="flex flex-col gap-1">
													<label for="price" class="text-sm">Price</label>
													<input
														id="price"
														name="price"
														type="number"
														class="border-surface-0 rounded border p-2 text-sm"
														value={item.price}
													/>
												</div>
												<div class="flex flex-col gap-1">
													<label for="quantity" class="text-sm">Quantity</label>
													<input
														id="quantity"
														name="quantity"
														type="number"
														class="border-surface-0 rounded border p-2 text-sm"
														value={item.quantity}
													/>
												</div>
												<Button.Root
													type="submit"
													class="bg-sky text-crust hover:brightness-120 w-20 cursor-pointer rounded p-2 px-4 text-sm transition-all"
												>
													Submit
												</Button.Root>
											</div>
										</form>
									{/snippet}
								</Modal>

								<Modal bind:open={
                                    () => deletesOpen.has(item.id!) ? deletesOpen.get(item.id!)! : deletesOpen.set(item.id!, false) && deletesOpen.get(item.id!)!,
                                    (value) => deletesOpen.set(item.id!, value)
                                }>
									{#snippet trigger()}
										<button
											class="bg-red text-crust hover:brightness-120 block cursor-pointer rounded p-2 drop-shadow-md"
										>
											<Trash />
										</button>
									{/snippet}

									{#snippet content()}
										<h1 class="border-surface-0 border-b py-3 text-center text-xl font-bold">
											Delete {item.name}
										</h1>
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
												<span class="text-center">Are you sure you want to delete "{item.name}"?</span
												>
												<div class="flex justify-center gap-4">
													<Button.Root
														type="submit"
														class="bg-sky text-crust hover:brightness-120 cursor-pointer rounded p-2 px-4 text-sm transition-all"
													>
														Confirm
													</Button.Root>
												</div>
											</div>
										</form>
									{/snippet}
								</Modal>
							</div>
						</td>
					</tr>
				{/each}
			{/await}
		</tbody>
	</table>
</div>

<div class="mx-4 mt-1 flex justify-end">
	<Modal bind:open={addedOpen}>
		{#snippet trigger()}
			<button
				class="bg-sky text-crust hover:brightness-120 cursor-pointer rounded p-2 px-4 drop-shadow-md"
			>
				<Plus />
			</button>
		{/snippet}

		{#snippet content()}
			<h1 class="border-surface-0 border-b py-3 text-center text-xl font-bold">Add Item</h1>
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
						<input
							id="name"
							name="name"
							type="text"
							class="border-surface-0 rounded border p-2 text-sm"
						/>
					</div>
					<div class="flex flex-col gap-1">
						<label for="description" class="text-sm">Description</label>
						<input
							id="description"
							name="description"
							type="text"
							class="border-surface-0 rounded border p-2 text-sm"
						/>
					</div>
					<div class="flex flex-col gap-1">
						<label for="price" class="text-sm">Price</label>
						<input
							id="price"
							name="price"
							type="number"
							class="border-surface-0 rounded border p-2 text-sm"
						/>
					</div>
					<div class="flex flex-col gap-1">
						<label for="quantity" class="text-sm">Quantity</label>
						<input
							id="quantity"
							name="quantity"
							type="number"
							class="border-surface-0 rounded border p-2 text-sm"
						/>
					</div>
					<Button.Root
						type="submit"
						class="bg-sky text-crust hover:brightness-120 w-fit cursor-pointer rounded p-2 px-4 text-sm transition-all"
					>
						Submit
					</Button.Root>
				</div>
			</form>
		{/snippet}
	</Modal>
</div>
