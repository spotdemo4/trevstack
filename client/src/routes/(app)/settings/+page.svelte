<script lang="ts">
	import { UserClient } from '$lib/transport';
	import Button from '$lib/ui/Button.svelte';
	import Modal from '$lib/ui/Modal.svelte';
	import { ConnectError } from '@connectrpc/connect';
	import { Avatar, Separator } from 'bits-ui';
	import { toast } from 'svelte-sonner';

	let user = UserClient.getUser({}).then((res) => {
		return res.user;
	});
	let key = $state('');
</script>

<div class="flex h-[calc(100vh-50px)]">
	<div class="m-auto flex w-96 flex-col gap-4 p-4">
		{#await user then user}
			<div class="flex items-center justify-center gap-4">
				<div
					class="outline-surface-2 bg-text text-crust h-9 w-9 select-none rounded-full outline outline-offset-2 text-sm"
				>
					<Avatar.Root class="flex h-full w-full items-center justify-center">
						<Avatar.Image src={user?.profilePicture} alt={`${user?.username}'s avatar`} class="rounded-full" />
						<Avatar.Fallback class="font-medium uppercase"
							>{user?.username.substring(0, 2)}</Avatar.Fallback
						>
					</Avatar.Root>
				</div>
				<h1 class="overflow-x-hidden text-2xl font-medium">{user?.username}</h1>
			</div>
		{/await}

		<Separator.Root class="bg-surface-0 h-px" />

		<div class="flex justify-around gap-2">
			<Modal>
				{#snippet trigger()}
					<Button className="bg-text">Generate API Key</Button>
				{/snippet}

				{#snippet content()}
					<h1 class="border-surface-0 border-b py-3 text-center text-xl font-bold">
						Generate API Key
					</h1>
					{#if key == ''}
						<form
							onsubmit={async (e) => {
								e.preventDefault();
								const form = e.target as HTMLFormElement;
								const formData = new FormData(form);

								try {
									const response = await UserClient.getAPIKey({
										password: formData.get('password')?.toString(),
										confirmPassword: formData.get('confirm-password')?.toString()
									});

									if (response.key) {
										key = response.key;
										form.reset();
									}
								} catch (err) {
									const error = ConnectError.from(err);
									toast.error(error.rawMessage);
								}
							}}
						>
							<div class="flex flex-col gap-4 p-3">
								<div class="flex flex-col gap-1">
									<label for="password" class="text-sm">Password</label>
									<input
										id="password"
										name="password"
										type="password"
										class="border-surface-0 rounded border p-2 text-sm"
									/>
								</div>
								<div class="flex flex-col gap-1">
									<label for="confirm-password" class="text-sm">Confirm Password</label>
									<input
										id="confirm-password"
										name="confirm-password"
										type="password"
										class="border-surface-0 rounded border p-2 text-sm"
									/>
								</div>
								<Button type="submit">Submit</Button>
							</div>
						</form>
					{:else}
						<div class="p-3">
							<span class="text-wrap break-all">{key}</span>
						</div>
					{/if}
				{/snippet}
			</Modal>

			<Modal>
				{#snippet trigger()}
					<Button className="bg-text">Change Profile Picture</Button>
				{/snippet}

				{#snippet content()}
					<h1 class="border-surface-0 border-b py-3 text-center text-xl font-bold">
						Change Profile Picture
					</h1>
					<form
						onsubmit={async (e) => {
							e.preventDefault();
							const form = e.target as HTMLFormElement;

							let fileInput = document.getElementById('file') as HTMLInputElement;
							let file = fileInput.files?.[0];

							if (!file) {
								toast.error('No file selected');
								return;
							}

							const data = await file.bytes();

							try {
								const response = await UserClient.updateProfilePicture({
									fileName: file.name,
									data: data,
								});

								if (response.user) {
									toast.success('Profile picture updated');
									form.reset();

								}
							} catch (err) {
								const error = ConnectError.from(err);
								toast.error(error.rawMessage);
							}
						}}
					>
						<div class="flex flex-col gap-4 p-3">
							<div class="flex flex-col gap-1">
								<label for="file" class="text-sm">Profile Picture</label>
								<input
									id="file"
									name="file"
									type="file"
									class="border-surface-0 rounded border p-2 text-sm"
								/>
							</div>
							<Button type="submit">Submit</Button>
						</div>
					</form>
				{/snippet}
			</Modal>
		</div>

		<form
			onsubmit={async (e) => {
				e.preventDefault();
				const form = e.target as HTMLFormElement;
				const formData = new FormData(form);

				try {
					await UserClient.updatePassword({
						oldPassword: formData.get('old-password')?.toString(),
						newPassword: formData.get('new-password')?.toString(),
						confirmPassword: formData.get('confirm-password')?.toString()
					});

					toast.success('password updated successfully');
					form.reset();
				} catch (err) {
					const error = ConnectError.from(err);
					toast.error(error.rawMessage);
				}
			}}
			class="bg-mantle border-surface-0 rounded border p-4 drop-shadow-md"
		>
			<div class="flex flex-col gap-4">
				<div class="flex flex-col gap-1">
					<label for="old-password" class="text-sm">Old Password</label>
					<input
						id="old-password"
						name="old-password"
						type="password"
						class="border-surface-0 rounded border p-2 text-sm"
					/>
				</div>
				<div class="flex flex-col gap-1">
					<label for="new-password" class="text-sm">New Password</label>
					<input
						id="new-password"
						name="new-password"
						type="password"
						class="border-surface-0 rounded border p-2 text-sm"
					/>
				</div>
				<div class="flex flex-col gap-1">
					<label for="confirm-password" class="text-sm">Confirm New Password</label>
					<input
						id="confirm-password"
						name="confirm-password"
						type="password"
						class="border-surface-0 rounded border p-2 text-sm"
					/>
				</div>
				<Button type="submit">Submit</Button>
			</div>
		</form>
	</div>
</div>
