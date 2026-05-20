import Theme from "$lib/theme";
import { AppToaster } from "$lib/toast";
import { ExternalLink } from "lucide-solid";
import { type Component, children, type JSX } from "solid-js";

import Github from "./Github";
import Navbar from "./Navbar";
import NavLink from "./NavLink";

type LayoutProps = {
  children?: JSX.Element;
};

const Layout: Component<LayoutProps> = (props) => {
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
            <Theme.Swap />
            <Github />
          </div>
        </div>
      </header>
      <main class="grow">{resolved()}</main>
      <AppToaster />
    </div>
  );
};

export default Layout;
