import logo from "$assets/logo.svg";
import { Navbar } from "$lib/navbar";
import { AppToaster } from "$lib/toast";
import { A } from "@solidjs/router";
import { type Component, children, type JSX } from "solid-js";

const Layout: Component<{ children?: JSX.Element }> = (props) => {
  const resolved = children(() => props.children);

  return (
    <div class="flex min-h-dvh flex-col">
      <header class="h-header border-ctp-surface1 bg-ctp-crust sticky top-0 z-50 flex justify-between border-b px-4">
        <h1 class="flex cursor-default items-center gap-2 font-mono text-lg font-semibold">
          TrevStack <img src={logo} class="h-6" alt="logo" />
        </h1>
        <Navbar>
          <A
            end
            href="/"
            class="text-ctp-subtext0 hover:text-ctp-text inline-flex px-1 text-sm transition-colors"
            activeClass="text-ctp-text"
          >
            Home
          </A>
          <A
            href="/numbers"
            class="text-ctp-subtext0 hover:text-ctp-text inline-flex px-1 text-sm transition-colors"
            activeClass="text-ctp-text"
          >
            Numbers
          </A>
          <Navbar.Indicator />
        </Navbar>
      </header>
      <main class="grow">{resolved()}</main>
      <AppToaster />
    </div>
  );
};

export default Layout;
