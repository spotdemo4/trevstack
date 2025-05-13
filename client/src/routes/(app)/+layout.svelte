<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { PageData } from './$types';
	import * as Avatar from '$lib/ui/avatar';
	import { Button } from '$lib/ui/button';
	import * as DropdownMenu from '$lib/ui/dropdown-menu';
	import * as Sheet from '$lib/ui/sheet';
	import {
		Book,
		House,
		LayoutGrid,
		LayoutList,
		LogOut,
		Menu,
		Moon,
		Settings,
		Sun
	} from '@lucide/svelte';
	import { goto } from '$app/navigation';
	import { userState } from '$lib/sharedState.svelte';
	import { AuthClient, UserClient } from '$lib/transport';
	import { toggleMode } from 'mode-watcher';
	import { toast } from 'svelte-sonner';
	import pkg from '../../../package.json' with { type: 'json' };

	interface Props {
		data: PageData | undefined;
		children?: Snippet;
	}
	let { data = $bindable(), children }: Props = $props();

	UserClient.getUser({}).then((res) => {
		userState.user = res.user;
	});

	let sidebarOpen = $state(false);

	async function logout() {
		await AuthClient.logout({});
		await goto('/auth');
		toast.success('successfully logged out');
		userState.user = undefined;

		if (sidebarOpen) {
			sidebarOpen = false;
		}
	}
</script>

<div class="flex min-h-dvh flex-col justify-between gap-2">
	<header
		class="bg-mantle border-surface text-text sticky top-0 z-50 flex h-14 w-full items-center justify-between border-b px-4 py-4 lg:px-10"
	>
		<!-- Left -->
		<a href="/" class="flex items-center gap-2 text-2xl font-bold tracking-wider select-none">
			TrevStack
			<LayoutGrid />
		</a>

		<!-- Center -->
		<div class="bg-crust hidden items-center gap-2 rounded-md p-1 lg:flex">
			<Button variant="ghost" class="hover:bg-based" href="/">
				<House />
				Home
			</Button>
			<Button variant="ghost" class="hover:bg-based" href="/items">
				<LayoutList />
				Items
			</Button>
			<Button variant="ghost" class="hover:bg-based" href="/docs">
				<Book />
				Docs
			</Button>
		</div>

		<!-- Right -->
		<div class="flex items-center gap-4">
			<DropdownMenu.Root>
				<DropdownMenu.Trigger class="hidden lg:flex">
					{#snippet child({ props })}
						<Avatar.Root {...props}>
							{#if userState.user?.profilePictureId}
								<Avatar.Image
									class="hover:brightness-125"
									src={`/file/${userState.user.profilePictureId}`}
									alt={`${userState.user?.username}'s avatar`}
								/>
							{/if}
							<Avatar.Fallback class="hover:bg-surface-1"
								>{userState.user?.username.substring(0, 2).toUpperCase()}</Avatar.Fallback
							>
						</Avatar.Root>
					{/snippet}
				</DropdownMenu.Trigger>
				<DropdownMenu.Content>
					<DropdownMenu.Item
						class="hidden dark:flex"
						onclick={() => {
							toggleMode();
						}}
					>
						<Sun />
						Light Mode
					</DropdownMenu.Item>
					<DropdownMenu.Item
						class="flex dark:hidden"
						onclick={() => {
							toggleMode();
						}}
					>
						<Moon />
						Dark Mode
					</DropdownMenu.Item>
					<DropdownMenu.Link class="flex" href="/settings">
						<Settings />
						Settings
					</DropdownMenu.Link>
					<DropdownMenu.Item
						class="flex"
						onclick={() => {
							logout();
						}}
					>
						<LogOut />
						Log Out
					</DropdownMenu.Item>
				</DropdownMenu.Content>
			</DropdownMenu.Root>

			<Sheet.Root bind:open={sidebarOpen}>
				<Sheet.Trigger>
					{#snippet child({ props })}
						<Button variant="outline" {...props}>
							<Menu />
						</Button>
					{/snippet}
				</Sheet.Trigger>
				<Sheet.Content class="flex flex-col justify-between overflow-auto pt-12">
					<div class="flex flex-col gap-1">
						<Button
							variant="outline"
							class="w-full"
							href="/"
							onclick={() => {
								sidebarOpen = false;
							}}
						>
							<House />
							Home
						</Button>
						<Button
							variant="outline"
							class="w-full"
							href="/items"
							onclick={() => {
								sidebarOpen = false;
							}}
						>
							<LayoutList />
							Items
						</Button>
						<Button
							variant="outline"
							class="w-full"
							href="/docs"
							onclick={() => {
								sidebarOpen = false;
							}}
						>
							<Book />
							Docs
						</Button>
					</div>
					<div class="flex flex-col gap-1">
						<span class="text-text font-bold">Settings</span>
						<Button
							variant="outline"
							class="hidden w-full dark:flex"
							onclick={() => {
								toggleMode();
							}}
						>
							<Sun />
							Light Mode
						</Button>
						<Button
							variant="outline"
							class="flex w-full dark:hidden"
							onclick={() => {
								toggleMode();
							}}
						>
							<Moon />
							Dark Mode
						</Button>
						<Button
							variant="outline"
							class="flex w-full"
							href="/settings"
							onclick={() => {
								sidebarOpen = false;
							}}
						>
							<Settings />
							Settings
						</Button>
						<Button
							variant="outline"
							class="flex w-full"
							onclick={() => {
								logout();
							}}
						>
							<LogOut />
							Log Out
						</Button>
					</div>
				</Sheet.Content>
			</Sheet.Root>
		</div>
	</header>

	<div class="grow">
		{@render children?.()}
	</div>

	<footer class="border-surface text-subtext bg-mantle flex justify-center border-t py-1 text-xs">
		v. {pkg.version}
	</footer>
</div>
