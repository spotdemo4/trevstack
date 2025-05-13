<script lang="ts">
	import { Button } from '$lib/ui/button';
	import { Card } from '$lib/ui/card';
	import * as Form from '$lib/ui/form';
	import { Input } from '$lib/ui/input';
	import * as Tabs from '$lib/ui/tabs';
	import { ConnectError } from '@connectrpc/connect';
	import { Fingerprint } from '@lucide/svelte';
	import { startAuthentication } from '@simplewebauthn/browser';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { AuthService } from '$lib/connect/user/v1/auth_pb';
	import { coolForm } from '$lib/coolforms';
	import { AuthClient } from '$lib/transport';
	import { toast } from 'svelte-sonner';

	let tabValue = $state('login');

	async function redirect() {
		localStorage.setItem('username', login.input.username);
		login.input.password = '';

		if (page.url.searchParams.has('redir')) {
			const uri = decodeURIComponent(page.url.searchParams.get('redir')!);
			await goto(uri);
			return;
		}

		await goto('/');
	}

	const login = coolForm(AuthClient, AuthService.method.login, {
		init: {
			username: localStorage.getItem('username') ?? ''
		},
		reset: false,
		onResult: () => {
			redirect();
		}
	});

	const signup = coolForm(AuthClient, AuthService.method.signUp, {
		onResult: () => {
			tabValue = 'login';
			toast.success('Successfully created account, please log in');
		}
	});

	async function passkeyLogin() {
		const optionsJSON = JSON.parse(
			(
				await AuthClient.beginPasskeyLogin({
					username: login.input.username
				})
			).optionsJson
		).publicKey;
		const attResp = await startAuthentication({ optionsJSON });

		try {
			await AuthClient.finishPasskeyLogin({
				username: login.input.username,
				attestation: JSON.stringify(attResp)
			});
		} catch (e) {
			if (!(e instanceof ConnectError)) {
				throw e;
			}

			toast.error(e.message);
			return;
		}

		await redirect();
	}
</script>

<div class="flex h-dvh flex-col items-center justify-center">
	<Tabs.Root bind:value={tabValue} class="min-w-full px-2 sm:min-w-sm">
		<Tabs.List class="w-full">
			<Tabs.Trigger value="login">Log In</Tabs.Trigger>
			<Tabs.Trigger value="signup">Sign Up</Tabs.Trigger>
		</Tabs.List>
		<Tabs.Content value="login">
			<Card>
				<form use:login.impair class="flex flex-col gap-2">
					<Form.Field name="username">
						<Form.Label />
						<Input bind:value={login.input.username} autocomplete="username webauthn" />
						<Form.Errors bind:errors={login.errors.username} />
					</Form.Field>

					<Form.Field name="password">
						<Form.Label />
						<Input type="password" bind:value={login.input.password} />
						<Form.Errors bind:errors={login.errors.password} />
					</Form.Field>

					<Form.Errors bind:errors={login.errors.form} />

					<div class="flex gap-1">
						<Button type="submit" loading={login.loading()} class="grow">Submit</Button>
						{#if login.input.username}
							<Button type="button" onclick={passkeyLogin} class="min-w-18"><Fingerprint /></Button>
						{/if}
					</div>
				</form>
			</Card>
		</Tabs.Content>
		<Tabs.Content value="signup">
			<Card>
				<form use:signup.impair class="flex flex-col gap-2">
					<Form.Field name="username">
						<Form.Label />
						<Input bind:value={signup.input.username} />
						<Form.Errors bind:errors={signup.errors.username} />
					</Form.Field>

					<Form.Field name="password">
						<Form.Label />
						<Input type="password" bind:value={signup.input.password} />
						<Form.Errors bind:errors={signup.errors.password} />
					</Form.Field>

					<Form.Field name="confirm_password">
						<Form.Label />
						<Input type="password" bind:value={signup.input.confirmPassword} />
						<Form.Errors bind:errors={signup.errors.confirmPassword} />
					</Form.Field>

					<Form.Errors bind:errors={signup.errors.form} />

					<Button type="submit" loading={signup.loading()}>Submit</Button>
				</form>
			</Card>
		</Tabs.Content>
	</Tabs.Root>
</div>
