<script lang="ts">
	import {
		LayoutGrid,
		Settings,
		LogOut,
		Menu,
		LayoutList,
		Book,
		House,
		type Icon as IconType
	} from '@lucide/svelte';
	import { NavigationMenu, Popover, Separator, Dialog } from 'bits-ui';
	import { fade, fly, slide } from 'svelte/transition';
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';
	import { AuthClient, UserClient } from '$lib/transport';
	import { page } from '$app/state';
	import { cn } from '$lib/utils';
	import { userState } from '$lib/sharedState.svelte';
	import Avatar from '$lib/ui/Avatar.svelte';

	let { children } = $props();

	UserClient.getUser({}).then((res) => {
		userState.user = res.user;
	});

	let sidebarOpen = $state(false);
	let popupOpen = $state(false);

	type MenuItem = {
		name: string;
		href: string;
		icon: typeof IconType;
	};
	const menuItems: MenuItem[] = [
		{
			name: 'Home',
			href: '/',
			icon: House
		},
		{
			name: 'Items',
			href: '/items/',
			icon: LayoutList
		},
		{
			name: 'Docs',
			href: '/docs/',
			icon: Book
		}
	];

	async function logout() {
		await AuthClient.logout({});
		await goto('/auth');
		toast.success('logged out successfully');
		userState.user = undefined;

		if (sidebarOpen) {
			sidebarOpen = false;
		}
	}
</script>

<header
	class="border-surface-0 bg-mantle fixed z-50 flex h-[50px] w-full items-center justify-between border-b p-2 px-6 drop-shadow-md"
>
	<div class="flex items-center gap-4">
		<Dialog.Root bind:open={sidebarOpen}>
			<Dialog.Trigger class="hover:bg-surface-0 cursor-pointer rounded p-1 px-3 transition-all">
				<Menu />
			</Dialog.Trigger>
			<Dialog.Portal>
				<Dialog.Overlay forceMount>
					{#snippet child({ props, open })}
						{#if open}
							<div
								{...props}
								transition:fade={{
									duration: 150
								}}
							>
								<div class="fixed inset-0 z-50 mt-[50px] bg-black/50"></div>
							</div>
						{/if}
					{/snippet}
				</Dialog.Overlay>
				<Dialog.Content forceMount>
					{#snippet child({ props, open })}
						{#if open}
							<div
								class="bg-mantle border-surface-0 fixed inset-0 z-50 mt-[50px] flex w-60 flex-col justify-between border-r drop-shadow-md"
								{...props}
								transition:slide={{
									axis: 'x'
								}}
							>
								<NavigationMenu.Root orientation="vertical">
									<NavigationMenu.List
										class="flex w-full flex-col gap-2 overflow-x-hidden overflow-y-auto p-2"
									>
										{#each menuItems as item (item.name)}
											{@const Icon = item.icon}
											<NavigationMenu.Item>
												<NavigationMenu.Link
													class={cn(
														'hover:bg-surface-0 flex gap-2 rounded-lg p-2 whitespace-nowrap transition-all select-none',
														page.url.pathname === item.href && 'bg-surface-0'
													)}
													href={item.href}
													onSelect={() => {
														if (sidebarOpen) {
															sidebarOpen = false;
														}
													}}
												>
													<Icon />
													<span>{item.name}</span>
												</NavigationMenu.Link>
											</NavigationMenu.Item>
										{/each}
									</NavigationMenu.List>
								</NavigationMenu.Root>

								<div class="border-surface-0 flex flex-col gap-2 border-t p-2">
									<a
										href="/settings"
										class="hover:bg-surface-0 flex items-center gap-2 rounded-lg p-2 transition-all select-none"
										onclick={() => {
											if (sidebarOpen) {
												sidebarOpen = false;
											}
										}}
									>
										<Settings />
										<span>Settings</span>
									</a>

									<button
										class="hover:bg-surface-0 flex w-full cursor-pointer items-center gap-2 rounded-lg p-2 whitespace-nowrap transition-all"
										onclick={logout}
									>
										<LogOut size="20" />
										Log out
									</button>
								</div>
							</div>
						{/if}
					{/snippet}
				</Dialog.Content>
			</Dialog.Portal>
		</Dialog.Root>

		<a href="/" class="flex items-center gap-2 text-2xl font-bold tracking-wider select-none">
			TrevStack
			<LayoutGrid />
		</a>
	</div>

	<NavigationMenu.Root class="hidden md:block">
		<NavigationMenu.List class="flex gap-2">
			{#each menuItems as item (item.name)}
				<NavigationMenu.Item>
					<NavigationMenu.Link
						class={cn(
							'hover:bg-surface-0 flex gap-2 rounded-lg p-1 px-2 transition-all select-none',
							page.url.pathname === item.href && 'bg-surface-0'
						)}
						href={item.href}
					>
						<span>{item.name}</span>
					</NavigationMenu.Link>
				</NavigationMenu.Item>
			{/each}
		</NavigationMenu.List>
		<NavigationMenu.Viewport class="absolute" />
	</NavigationMenu.Root>

	<Popover.Root bind:open={popupOpen}>
		<Popover.Trigger
			class="outline-surface-2 bg-text text-crust h-9 w-9 cursor-pointer rounded-full text-sm outline outline-offset-2 transition-all hover:brightness-120"
		>
			<Avatar />
		</Popover.Trigger>
		<Popover.Content forceMount>
			{#snippet child({ wrapperProps, props, open })}
				{#if open}
					<div {...wrapperProps}>
						<div
							class="bg-mantle border-surface-0 m-1 rounded border drop-shadow-md transition-all"
							{...props}
							transition:fly
						>
							<a
								class="hover:bg-surface-0 flex items-center gap-1 p-3 px-4 text-sm"
								href="/settings"
								onclick={() => {
									if (popupOpen) {
										popupOpen = false;
									}
								}}
							>
								<Settings size="20" />
								Settings
							</a>
							<Separator.Root class="bg-surface-0 h-px" />
							<button
								class="hover:bg-surface-0 flex w-full cursor-pointer items-center gap-1 p-3 px-4 text-sm transition-all"
								onclick={logout}
							>
								<LogOut size="20" />
								Log out
							</button>
						</div>
					</div>
				{/if}
			{/snippet}
		</Popover.Content>
	</Popover.Root>
</header>

<div class="overflow-auto pt-[50px]">
	{@render children()}
</div>
