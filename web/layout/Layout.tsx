import { Toast } from "@kobalte/core";
import { NavigationMenu } from "@kobalte/core/navigation-menu";
import { A } from "@solidjs/router";
import { CodeXml } from "lucide-solid";
import { type Component, children, type JSX } from "solid-js";
import { Portal } from "solid-js/web";
import logo from "$assets/logo.svg";

const Layout: Component<{ children?: JSX.Element }> = (props) => {
	const resolved = children(() => props.children);

	return (
		<div class="flex min-h-dvh flex-col">
			<header class="sticky top-0 z-50 flex h-header items-center justify-between border-ctp-surface1 border-b bg-ctp-crust px-4 py-3">
				<h1 class="flex cursor-default items-center gap-2 font-mono font-semibold text-lg">
					TrevStack <img src={logo} class="h-6" alt="logo" />
				</h1>
				<NavigationMenu class="flex gap-5">
					<NavigationMenu.Trigger
						as={A}
						href="/"
						class="rounded underline decoration-2 decoration-transparent underline-offset-4 transition-colors hover:decoration-ctp-overlay0 focus:outline-none"
						activeClass="!decoration-ctp-sky"
						end={true}
					>
						Home
					</NavigationMenu.Trigger>
					<NavigationMenu.Trigger
						as={A}
						href="/numbers"
						class="rounded underline decoration-2 decoration-transparent underline-offset-4 transition-colors hover:decoration-ctp-overlay0 focus:outline-none"
						activeClass="!decoration-ctp-sky"
					>
						Numbers
					</NavigationMenu.Trigger>
					<NavigationMenu.Trigger
						as="a"
						href="https://github.com/spotdemo4/trevstack"
						target="_blank"
						class="rounded text-ctp-text transition-colors hover:text-ctp-sky focus:outline-none"
					>
						<CodeXml />
					</NavigationMenu.Trigger>
					<NavigationMenu.Viewport>
						<NavigationMenu.Arrow />
					</NavigationMenu.Viewport>
				</NavigationMenu>
			</header>
			<main class="grow">{resolved()}</main>
			<Portal>
				<Toast.Region>
					<Toast.List class="fixed right-4 bottom-4 z-50 flex w-96 max-w-[calc(100vw-2rem)] flex-col gap-2 outline-none" />
				</Toast.Region>
			</Portal>
		</div>
	);
};

export default Layout;
