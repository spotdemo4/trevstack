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
	import { fly, slide } from 'svelte/transition';
	import { toast } from 'svelte-sonner';
	import { goto } from '$app/navigation';
	import { AuthClient } from '$lib/transport';
	import { page } from '$app/state';
	import { cn } from '$lib/utils';
	let { children } = $props();

	const username = localStorage.getItem('username');
	let sidebarOpen = $state(false);

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
		localStorage.removeItem('username');
		await goto('/auth');
		toast.success('logged out successfully');
	}
</script>

<header
	class="border-surface-0 bg-mantle fixed flex h-[50px] w-full items-center justify-between border-b p-2 px-6 drop-shadow-md"
>
	<div class="flex items-center gap-4">
		<Dialog.Root bind:open={sidebarOpen}>
			<Dialog.Trigger class="hover:bg-surface-0 cursor-pointer rounded p-1 px-3 transition-all">
				<Menu />
			</Dialog.Trigger>
			<Dialog.Portal>
				<Dialog.Overlay class="fixed inset-0 z-50 mt-[50px] bg-black/50" />
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
									<NavigationMenu.List class="flex w-full flex-col gap-2 overflow-y-scroll p-2">
										{#each menuItems as item}
											{@const Icon = item.icon}
											<NavigationMenu.Item>
												<NavigationMenu.Link
													class={cn(
														'hover:bg-surface-0 flex select-none gap-2 whitespace-nowrap rounded-lg p-2 transition-all',
														page.url.pathname === item.href && 'bg-surface-0'
													)}
													href={item.href}
													onclick={() => {
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
										class="hover:bg-surface-0 flex select-none items-center gap-2 rounded-lg p-2 transition-all"
									>
										<Settings />
										<span>Settings</span>
									</a>

									<button
										class="hover:bg-surface-0 flex w-full cursor-pointer items-center gap-2 whitespace-nowrap rounded-lg p-2 transition-all"
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

		<a href="/" class="flex select-none items-center gap-2 text-2xl font-bold tracking-wider">
			TrevStack
			<LayoutGrid />
		</a>
	</div>

	<NavigationMenu.Root class="hidden md:block">
		<NavigationMenu.List class="flex gap-2">
			{#each menuItems as item}
				<NavigationMenu.Item>
					<NavigationMenu.Link
						class={cn(
							'hover:bg-surface-0 flex select-none gap-2 rounded-lg p-1 px-2 transition-all',
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

	<Popover.Root>
		<Popover.Trigger
			class="border-surface-2 hover:bg-surface-0 cursor-pointer rounded border p-1 px-4 text-sm transition-all"
		>
			{username}
		</Popover.Trigger>
		<Popover.Content forceMount>
			{#snippet child({ wrapperProps, props, open })}
				{#if open}
					<div {...wrapperProps}>
						<div
							class="bg-mantle border-surface-0 z-50 mt-1 rounded border drop-shadow-md transition-all"
							{...props}
							transition:fly
						>
							<a
								class="hover:bg-surface-0 flex items-center gap-1 p-3 px-4 text-sm"
								href="/settings"
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

<div class="pt-[50px]">
	{@render children()}
</div>
