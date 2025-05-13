<script lang="ts">
	import * as Avatar from '$lib/ui/avatar';
	import { Button } from '$lib/ui/button';
	import { Card } from '$lib/ui/card';
	import * as Dialog from '$lib/ui/dialog';
	import * as Form from '$lib/ui/form';
	import { Input } from '$lib/ui/input';
	import { Separator } from '$lib/ui/separator';
	import { ConnectError } from '@connectrpc/connect';
	import { startRegistration } from '@simplewebauthn/browser';
	import { UserService } from '$lib/connect/user/v1/user_pb';
	import { coolForm, newState } from '$lib/coolforms';
	import { userState } from '$lib/sharedState.svelte';
	import { UserClient } from '$lib/transport';
	import { toast } from 'svelte-sonner';

	const updatePassword = coolForm(UserClient, UserService.method.updatePassword, {
		reset: true,
		onResult: () => {
			toast.success('Successfully updated password');
		}
	});

	async function createPasskey() {
		const optionsJSON = JSON.parse(
			(await UserClient.beginPasskeyRegistration({})).optionsJson
		).publicKey;
		const attResp = await startRegistration({ optionsJSON });

		try {
			await UserClient.finishPasskeyRegistration({
				attestation: JSON.stringify(attResp)
			});
		} catch (e) {
			if (!(e instanceof ConnectError)) {
				throw e;
			}

			toast.error(e.message);
			return;
		}

		toast.success('Added new passkey');
	}
</script>

{#snippet apiKeyModal()}
	{@const s = newState({ open: false, key: '' })}
	{@const apikey = coolForm(UserClient, UserService.method.getAPIKey, {
		onResult: (result) => {
			s.key = result.key;
		}
	})}

	<Dialog.Root bind:open={s.open}>
		<Dialog.Trigger>
			{#snippet child({ props })}
				<Button {...props}>Generate API Key</Button>
			{/snippet}
		</Dialog.Trigger>
		<Dialog.Content>
			{#if !s.key}
				<Dialog.Header>
					<Dialog.Title>Generate API Key</Dialog.Title>
				</Dialog.Header>
				<form use:apikey.impair class="flex flex-col gap-2">
					<Form.Field name="password">
						<Form.Label />
						<Input type="password" bind:value={apikey.input.password} />
						<Form.Errors bind:errors={apikey.errors.password} />
					</Form.Field>

					<Form.Field name="confirm_password">
						<Form.Label />
						<Input type="password" bind:value={apikey.input.confirmPassword} />
						<Form.Errors bind:errors={apikey.errors.confirmPassword} />
					</Form.Field>

					<Form.Errors bind:errors={apikey.errors.form} />

					<Button type="submit" loading={apikey.loading()}>Submit</Button>
				</form>
			{:else}
				<Dialog.Header>
					<Dialog.Title>API Key</Dialog.Title>
				</Dialog.Header>
				<span class="break-all">{s.key}</span>
			{/if}
		</Dialog.Content>
	</Dialog.Root>
{/snippet}

{#snippet profilePictureModal()}
	{@const s = newState({ open: false })}
	{@const updatepp = coolForm(UserClient, UserService.method.updateProfilePicture, {
		onSubmit: async (form, input) => {
			const file = form.get('file');

			if (file instanceof File) {
				input.fileName = file.name;
				input.data = await file.bytes();
			}

			return input;
		},
		onResult: () => {
			s.open = false;
		}
	})}

	<Dialog.Root bind:open={s.open}>
		<Dialog.Trigger>
			{#snippet child({ props })}
				<Button {...props}>Change Profile Picture</Button>
			{/snippet}
		</Dialog.Trigger>
		<Dialog.Content>
			<Dialog.Header>
				<Dialog.Title>Change Profile Picture</Dialog.Title>
			</Dialog.Header>
			<form use:updatepp.impair class="flex flex-col gap-2">
				<Form.Field name="file">
					<Form.Label />
					<Input type="file" />
					<Form.Errors bind:errors={updatepp.errors.data} />
				</Form.Field>

				<Form.Errors bind:errors={updatepp.errors.form} />

				<Button type="submit" loading={updatepp.loading()}>Submit</Button>
			</form>
		</Dialog.Content>
	</Dialog.Root>
{/snippet}

<div class="h-body flex">
	<div class="m-auto flex w-96 flex-col gap-4 p-4">
		<div class="flex items-center justify-center gap-4">
			<Avatar.Root class="size-12">
				{#if userState.user?.profilePictureId}
					<Avatar.Image
						src={`/file/${userState.user.profilePictureId}`}
						alt={`${userState.user?.username}'s avatar`}
					/>
				{/if}
				<Avatar.Fallback class="text-md"
					>{userState.user?.username.substring(0, 2).toUpperCase()}</Avatar.Fallback
				>
			</Avatar.Root>
			<h1 class="overflow-x-hidden text-2xl font-medium">{userState.user?.username}</h1>
		</div>

		<Separator />

		<div class="flex flex-wrap justify-center gap-2">
			{@render apiKeyModal()}

			{@render profilePictureModal()}

			<Button onclick={createPasskey}>Register Device</Button>
		</div>

		<Card>
			<form use:updatePassword.impair class="flex flex-col gap-2">
				<Form.Field name="old_password">
					<Form.Label />
					<Input type="password" bind:value={updatePassword.input.oldPassword} />
					<Form.Errors bind:errors={updatePassword.errors.oldPassword} />
				</Form.Field>

				<Form.Field name="new_password">
					<Form.Label />
					<Input type="password" bind:value={updatePassword.input.newPassword} />
					<Form.Errors bind:errors={updatePassword.errors.newPassword} />
				</Form.Field>

				<Form.Field name="confirm_password">
					<Form.Label />
					<Input type="password" bind:value={updatePassword.input.confirmPassword} />
					<Form.Errors bind:errors={updatePassword.errors.confirmPassword} />
				</Form.Field>

				<Form.Errors bind:errors={updatePassword.errors.form} />

				<Button type="submit" loading={updatePassword.loading()}>Submit</Button>
			</form>
		</Card>
	</div>
</div>
