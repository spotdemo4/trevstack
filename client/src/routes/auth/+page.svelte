<script lang="ts">
	import { Tabs } from 'bits-ui';
	import { cn } from '$lib/utils';
	import { AuthClient } from '$lib/transport';
	import { goto } from '$app/navigation';
	import { ConnectError } from '@connectrpc/connect';
	import { toast } from 'svelte-sonner';
	import Button from '$lib/ui/Button.svelte';

	let tab = $state('login');
</script>

<div class="flex h-screen flex-col items-center justify-center">
	<Tabs.Root bind:value={tab} class="w-[390px] p-3">
		<Tabs.List
			class="bg-mantle border-surface-0 flex w-full justify-around gap-1 rounded-lg border p-1 drop-shadow-md"
		>
			<Tabs.Trigger
				value="login"
				class={cn(
					'hover:bg-surface-0 grow cursor-pointer rounded p-2 transition-all',
					tab == 'login' && 'bg-surface-0'
				)}>Log In</Tabs.Trigger
			>
			<Tabs.Trigger
				value="signup"
				class={cn(
					'hover:bg-surface-0 grow cursor-pointer rounded p-2 transition-all',
					tab == 'signup' && 'bg-surface-0'
				)}>Sign Up</Tabs.Trigger
			>
		</Tabs.List>
		<Tabs.Content
			value="login"
			class="bg-mantle border-surface-0 mt-2 rounded-lg border p-6 drop-shadow-md"
		>
			<form
				onsubmit={async (e) => {
					e.preventDefault();
					const formData = new FormData(e.target as HTMLFormElement);
					const username = formData.get('login-username')?.toString();
					const password = formData.get('login-password')?.toString();

					try {
						const response = await AuthClient.login({
							username: username,
							password: password
						});

						if (response.token && username) {
							goto('/');
						}
					} catch (err) {
						const error = ConnectError.from(err);
						toast.error(error.rawMessage);
					}
				}}
			>
				<div class="flex flex-col gap-4">
					<div class="flex flex-col gap-1">
						<label for="login-username" class="text-sm">Username</label>
						<input
							id="login-username"
							name="login-username"
							type="text"
							class="border-surface-0 rounded border p-2 text-sm"
						/>
					</div>
					<div class="flex flex-col gap-1">
						<label for="login-password" class="text-sm">Password</label>
						<input
							id="login-password"
							name="login-password"
							type="password"
							class="border-surface-0 rounded border p-2 text-sm"
						/>
					</div>
					<Button type="submit">Submit</Button>
				</div>
			</form>
		</Tabs.Content>
		<Tabs.Content
			value="signup"
			class="bg-mantle border-surface-0 mt-2 rounded-lg border p-6 drop-shadow-md"
		>
			<form
				onsubmit={async (e) => {
					e.preventDefault();
					const form = e.target as HTMLFormElement;
					const formData = new FormData(form);

					try {
						await AuthClient.signUp({
							username: formData.get('signup-username')?.toString(),
							password: formData.get('signup-password')?.toString(),
							confirmPassword: formData.get('signup-confirm-password')?.toString()
						});

						toast.success('account created successfully, please log in');
						form.reset();
						tab = 'login';
					} catch (err) {
						const error = ConnectError.from(err);
						toast.error(error.rawMessage);
					}
				}}
			>
				<div class="flex flex-col gap-4">
					<div class="flex flex-col gap-1">
						<label for="signup-username" class="text-sm">Username</label>
						<input
							id="signup-username"
							name="signup-username"
							type="text"
							class="border-surface-0 rounded border p-2 text-sm"
						/>
					</div>
					<div class="flex flex-col gap-1">
						<label for="signup-password" class="text-sm">Password</label>
						<input
							id="signup-password"
							name="signup-password"
							type="password"
							class="border-surface-0 rounded border p-2 text-sm"
						/>
					</div>
					<div class="flex flex-col gap-1">
						<label for="signup-confirm-password" class="text-sm">Confirm Password</label>
						<input
							id="signup-confirm-password"
							name="signup-confirm-password"
							type="password"
							class="border-surface-0 rounded border p-2 text-sm"
						/>
					</div>
					<Button type="submit">Submit</Button>
				</div>
			</form>
		</Tabs.Content>
	</Tabs.Root>
</div>
