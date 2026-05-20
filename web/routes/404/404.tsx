import { A } from "@solidjs/router";
import type { Component } from "solid-js";

export const NotFound: Component = () => {
  return (
    <div class="flex h-body flex-col items-center justify-center gap-2">
      <span class="font-mono text-6xl font-bold text-ctp-overlay0">404</span>
      <p class="text-ctp-subtext0">Page not found</p>
      <A href="/" class="mt-2 text-sm text-ctp-blue hover:underline">
        Go home
      </A>
    </div>
  );
};
