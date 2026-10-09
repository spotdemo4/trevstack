import type { Component } from "solid-js";

export const NotFound: Component = () => {
  return (
    <div class="flex h-body flex-col items-center justify-center gap-2">
      <span class="font-mono text-6xl font-bold text-foreground-faint">404</span>
      <p class="text-foreground-subtle">Page not found</p>
      <a href="/" class="mt-2 text-sm text-link hover:underline">
        Go home
      </a>
    </div>
  );
};
