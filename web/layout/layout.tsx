import { Button } from "$lib/button";
import { ExternalLink, GitBranch } from "$lib/icon";
import type { JSX } from "@solidjs/web";
import { children, type Component } from "solid-js";

import { NavLink } from "./nav-link";
import { Navbar } from "./navbar";
import { ThemeSwitch } from "./theme-switch";
import { UserMenu } from "./user-menu";

type LayoutProps = {
  children?: JSX.Element;
};

export const Layout: Component<LayoutProps> = (props) => {
  const resolved = children(() => props.children);

  return (
    <div class="flex min-h-dvh flex-col">
      <header class="z-30 flex h-header border-b border-ctp-surface1 bg-ctp-crust min-[900px]:sticky min-[900px]:top-0">
        <h1 class="hidden shrink-0 cursor-default items-center gap-2 px-4 font-mono text-lg font-semibold md:flex">
          TrevStack <img src="/icons/icon.svg" class="h-6" alt="logo" />
        </h1>
        <div class="flex h-full min-w-0 flex-1 items-center">
          <div class="h-full min-w-0 flex-1 overflow-x-auto px-4">
            <Navbar class="ml-auto w-max min-w-max">
              <NavLink end href="/">
                Home
              </NavLink>
              <NavLink href="/numbers">Numbers</NavLink>
              <NavLink href="/metrics">Metrics</NavLink>
              <NavLink as="a" href="/docs" target="_blank" class="items-center gap-1">
                Docs <ExternalLink size={16} />
              </NavLink>
              <Navbar.Indicator />
            </Navbar>
          </div>
          <div class="flex shrink-0 items-center gap-2 pr-2 md:pr-4">
            <div class="hidden h-6 w-px bg-ctp-surface1 md:block" />
            <div class="hidden items-center gap-2 md:flex">
              <ThemeSwitch />
              <Button.Icon
                as="a"
                href="https://trev.zip/llc/stack"
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
      </header>
      <main class="grow">{resolved()}</main>
    </div>
  );
};
