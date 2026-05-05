import logo from "$assets/logo.svg";
import { Navbar } from "$lib/navbar";
import Theme from "$lib/theme";
import { AppToaster } from "$lib/toast";
import { A } from "@solidjs/router";
import { type Component, children, type JSX } from "solid-js";

import Github from "./Github";

type LayoutProps = {
  children?: JSX.Element;
};

const Layout: Component<LayoutProps> = (props) => {
  const resolved = children(() => props.children);

  return (
    <div class="flex min-h-dvh flex-col">
      <header class="sticky top-0 z-30 flex h-header justify-between border-b border-ctp-surface1 bg-ctp-crust px-4">
        <h1 class="flex cursor-default items-center gap-2 font-mono text-lg font-semibold">
          TrevStack <img src={logo} class="h-6" alt="logo" />
        </h1>
        <div class="flex h-full items-center gap-4">
          <Navbar>
            <A
              end
              href="/"
              class="inline-flex px-1 text-sm text-ctp-subtext0 transition-colors hover:text-ctp-text"
              activeClass="text-ctp-text"
            >
              Home
            </A>
            <A
              href="/numbers"
              class="inline-flex px-1 text-sm text-ctp-subtext0 transition-colors hover:text-ctp-text"
              activeClass="text-ctp-text"
            >
              Numbers
            </A>
            <A
              href="/metrics"
              class="inline-flex px-1 text-sm text-ctp-subtext0 transition-colors hover:text-ctp-text"
              activeClass="text-ctp-text"
            >
              Metrics
            </A>
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
