import { Button } from "$lib/button";
import { Drawer, useDrawer } from "$lib/drawer";
import { ExternalLink, GitBranch, Menu } from "$lib/icon";
import { createMediaQuery } from "$lib/media-query";
import type { JSX } from "@solidjs/web";
import { children, type Component, Show } from "solid-js";

import { version } from "../package.json";
import { NavLink } from "./nav-link";
import { Navbar } from "./navbar";
import { ThemeSwitch } from "./theme-switch";
import { UserMenu } from "./user-menu";

type LayoutProps = {
  children?: JSX.Element;
};

type NavigationProps = {
  mobile?: boolean;
  onNavigate?: JSX.EventHandlerUnion<HTMLAnchorElement, MouseEvent>;
};

const NavigationLinks: Component<NavigationProps> = (props) => (
  <>
    <NavLink
      end
      href="/"
      onClick={props.onNavigate}
      class={props.mobile ? "rounded-lg px-3 py-3" : undefined}
      activeClass={props.mobile ? "bg-ctp-surface0" : undefined}
    >
      Home
    </NavLink>
    <NavLink
      href="/numbers"
      onClick={props.onNavigate}
      class={props.mobile ? "rounded-lg px-3 py-3" : undefined}
      activeClass={props.mobile ? "bg-ctp-surface0" : undefined}
    >
      Numbers
    </NavLink>
    <NavLink
      href="/metrics"
      onClick={props.onNavigate}
      class={props.mobile ? "rounded-lg px-3 py-3" : undefined}
      activeClass={props.mobile ? "bg-ctp-surface0" : undefined}
    >
      Metrics
    </NavLink>
    <NavLink
      as="a"
      href="/docs"
      target="_blank"
      onClick={props.onNavigate}
      class={props.mobile ? "items-center gap-2 rounded-lg px-3 py-3" : "items-center gap-1"}
    >
      Docs <ExternalLink size={16} />
    </NavLink>
  </>
);

const MobileNavigation: Component = () => {
  const drawer = useDrawer();
  const closeOnNavigate: JSX.EventHandlerUnion<HTMLAnchorElement, MouseEvent> = (event) => {
    const sameTab =
      event.button === 0 &&
      !event.altKey &&
      !event.ctrlKey &&
      !event.metaKey &&
      !event.shiftKey &&
      (!event.currentTarget.target || event.currentTarget.target === "_self") &&
      !event.currentTarget.hasAttribute("download");
    drawer.close(sameTab ? { immediate: true } : undefined);
  };

  return (
    <nav aria-label="Main navigation" class="flex flex-col gap-1">
      <NavigationLinks mobile onNavigate={closeOnNavigate} />
    </nav>
  );
};

const MobileActions: Component = () => {
  const drawer = useDrawer();

  return (
    <div class="flex items-center gap-2 border-t border-ctp-surface1 pt-4">
      <ThemeSwitch />
      <UserMenu onBeforeNavigate={() => drawer.close({ immediate: true })} />
      <Button.Icon
        as="a"
        href="https://trev.zip/template/stack"
        target="_blank"
        rel="noopener noreferrer"
        aria-label="Open TrevStack on GitHub"
        onClick={() => drawer.close()}
      >
        <GitBranch />
      </Button.Icon>
    </div>
  );
};

const MobileMenu: Component = () => (
  <Drawer.Root direction="right">
    <Drawer.Trigger
      aria-label="Open navigation menu"
      aria-haspopup="dialog"
      class="m-0 mr-2 h-10 w-10 justify-center self-center rounded-md bg-transparent p-0 text-ctp-text hover:bg-ctp-surface0"
    >
      <Menu />
    </Drawer.Trigger>
    <Drawer.Content>
      <div class="flex items-center justify-between">
        <Drawer.Title>Menu</Drawer.Title>
        <Drawer.CloseTrigger aria-label="Close navigation menu" />
      </div>
      <MobileNavigation />
      <MobileActions />
    </Drawer.Content>
  </Drawer.Root>
);

export const Layout: Component<LayoutProps> = (props) => {
  const resolved = children(() => props.children);
  const isDesktop = createMediaQuery("(min-width: 900px)");

  return (
    <div class="flex min-h-dvh flex-col">
      <header class="z-30 flex h-header border-b border-ctp-surface1 bg-ctp-crust min-[900px]:sticky min-[900px]:top-0">
        <Show
          when={isDesktop()}
          fallback={
            <div class="flex h-full w-full items-center justify-between">
              <img src="/icons/icon.svg" class="ml-4 h-6" alt="TrevStack" />
              <MobileMenu />
            </div>
          }
        >
          <h1 class="hidden shrink-0 cursor-default items-center gap-2 px-4 font-mono text-lg font-semibold md:flex">
            TrevStack <img src="/icons/icon.svg" class="h-6" alt="logo" />
          </h1>
          <div class="flex h-full min-w-0 flex-1 items-center">
            <div class="h-full min-w-0 flex-1 overflow-x-auto px-4">
              <Navbar class="ml-auto w-max min-w-max">
                <NavigationLinks />
                <Navbar.Indicator />
              </Navbar>
            </div>
            <div class="flex shrink-0 items-center gap-2 pr-2 md:pr-4">
              <div class="hidden h-6 w-px bg-ctp-surface1 md:block" />
              <div class="hidden items-center gap-2 md:flex">
                <ThemeSwitch />
                <Button.Icon
                  as="a"
                  href="https://trev.zip/template/stack"
                  target="_blank"
                  rel="noopener noreferrer"
                  aria-label="Open TrevStack on GitHub"
                >
                  <GitBranch />
                </Button.Icon>
              </div>
              <UserMenu />
            </div>
          </div>
        </Show>
      </header>
      <main class="grow">{resolved()}</main>
      <footer class="flex h-footer shrink-0 items-center justify-center border-t border-ctp-surface1 bg-ctp-crust px-4 text-xs text-ctp-subtext0">
        v{version}
      </footer>
    </div>
  );
};
