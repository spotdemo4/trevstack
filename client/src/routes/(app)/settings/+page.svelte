<script lang="ts">
	import { UserClient } from '$lib/transport';
	import Button from '$lib/ui/Button.svelte';
	import Modal from '$lib/ui/Modal.svelte';
	import Input from '$lib/ui/Input.svelte';
	import { ConnectError } from '@connectrpc/connect';
	import { Separator } from 'bits-ui';
	import { toast } from 'svelte-sonner';
	import { userState } from '$lib/sharedState.svelte';
	import { createPasskey } from '$lib/webauthn';
	import { page } from '$app/state';
	import Avatar from '$lib/ui/Avatar.svelte';

	let openChangeProfilePicture = $state(false);
	let key = $state('');
</script>

<div class="flex h-[calc(100vh-50px)]">
	<div class="m-auto flex w-96 flex-col gap-4 p-4">
		<div class="flex items-center justify-center gap-4">
			<div
				class="outline-surface-2 bg-text text-crust h-9 w-9 select-none rounded-full text-sm outline outline-offset-2"
			>
				<Avatar />
			</div>
			<h1 class="overflow-x-hidden text-2xl font-medium">{userState.user?.username}</h1>
		</div>

		<Separator.Root class="bg-surface-0 h-px" />

		<div class="flex flex-wrap justify-around gap-2">
			<Modal>
				{#snippet trigger(props)}
					<Button {...props} className="bg-text">Generate API Key</Button>
				{/snippet}

				{#snippet title()}
					Generate API Key
				{/snippet}

				{#snippet content()}
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
									<Input name="password" type="password" />
								</div>
								<div class="flex flex-col gap-1">
									<label for="confirm-password" class="text-sm">Confirm Password</label>
									<Input name="confirm-password" type="password" />
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

			<Modal bind:open={openChangeProfilePicture}>
				{#snippet trigger(props)}
					<Button {...props} className="bg-text">Change Profile Picture</Button>
				{/snippet}

				{#snippet title()}
					Change Profile Picture
				{/snippet}

				{#snippet content()}
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
									data: data
								});

								if (response.user) {
									toast.success('Profile picture updated');
									form.reset();
									openChangeProfilePicture = false;
									userState.user = response.user;
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
								<Input name="file" type="file" />
							</div>
							<Button type="submit">Submit</Button>
						</div>
					</form>
				{/snippet}
			</Modal>

			<Button
				className="bg-text"
				onclick={async () => {
					if (userState.user) {
						await createPasskey(userState.user.username, userState.user.id, "what");

					}
				}}>Register Device</Button
			>
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
					<Input name="old-password" type="password" />
				</div>
				<div class="flex flex-col gap-1">
					<label for="new-password" class="text-sm">New Password</label>
					<Input name="new-password" type="password" />
				</div>
				<div class="flex flex-col gap-1">
					<label for="confirm-password" class="text-sm">Confirm New Password</label>
					<Input name="confirm-password" type="password" />
				</div>
				<Button type="submit">Submit</Button>
			</div>
		</form>
	</div>
</div>
