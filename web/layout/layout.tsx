import { Button } from "$lib/button";
import { AppToaster } from "$lib/toast";
import { ExternalLink } from "lucide-solid";
import { GitBranch } from "lucide-solid";
import { type Component, children, type JSX } from "solid-js";

import { NavLink } from "./nav-link";
import { Navbar } from "./navbar";
import { ThemeSwitch } from "./theme-switch";

type LayoutProps = {
  children?: JSX.Element;
};

export const Layout: Component<LayoutProps> = (props) => {
  const resolved = children(() => props.children);

  return (
    <div class="flex min-h-dvh flex-col">
      <header class="sticky top-0 z-30 flex h-header justify-center border-b border-ctp-surface1 bg-ctp-crust md:justify-between">
        <h1 class="hidden cursor-default items-center gap-2 px-4 font-mono text-lg font-semibold md:flex">
          TrevStack <img src="/icons/icon.svg" class="h-6" alt="logo" />
        </h1>
        <div class="flex h-full items-center gap-4 overflow-x-auto px-4">
          <Navbar>
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
          <div class="hidden h-6 w-px bg-ctp-surface1 md:block" />
          <div class="hidden items-center gap-2 md:flex">
            <ThemeSwitch />
            <Button.Icon
              as="a"
              href="https://trev.zip/trev/stack"
              target="_blank"
              rel="noopener noreferrer"
              aria-label="Open TrevStack on GitHub"
            >
              <GitBranch />
            </Button.Icon>
          </div>
        </div>
      </header>
      <main class="grow">{resolved()}</main>
      <AppToaster />
    </div>
  );
};
