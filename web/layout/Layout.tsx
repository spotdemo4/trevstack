import logo from "$assets/logo.svg";
import { Navbar } from "$lib/navbar";
import Theme from "$lib/theme";
import { AppToaster } from "$lib/toast";
import { ExternalLink } from "lucide-solid";
import { type Component, children, type JSX } from "solid-js";

import Github from "./Github";
import LayoutNavLink from "./LayoutNavLink";

type LayoutProps = {
  children?: JSX.Element;
};

const Layout: Component<LayoutProps> = (props) => {
  const resolved = children(() => props.children);

  return (
    <div class="flex min-h-dvh flex-col">
      <header class="sticky top-0 z-30 flex h-header justify-center border-b border-ctp-surface1 bg-ctp-crust md:justify-between">
        <h1 class="hidden cursor-default items-center gap-2 px-4 font-mono text-lg font-semibold md:flex">
          TrevStack <img src={logo} class="h-6" alt="logo" />
        </h1>
        <div class="flex h-full items-center gap-4 overflow-x-auto px-4">
          <Navbar>
            <LayoutNavLink end href="/">
              Home
            </LayoutNavLink>
            <LayoutNavLink href="/numbers">Numbers</LayoutNavLink>
            <LayoutNavLink href="/metrics">Metrics</LayoutNavLink>
            <LayoutNavLink as="a" href="/docs" target="_blank" class="items-center gap-1">
              Docs <ExternalLink size={16} />
            </LayoutNavLink>
            <Navbar.Indicator />
          </Navbar>
          <div class="h-6 w-px bg-ctp-surface1" />
          <div class="flex items-center gap-2">
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
