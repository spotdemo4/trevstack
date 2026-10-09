import type { Component } from "solid-js";

export const Forbidden: Component = () => {
  return (
    <div class="flex h-body flex-col items-center justify-center gap-2">
      <span class="font-mono text-6xl font-bold text-foreground-faint">403</span>
      <p class="text-foreground-subtle">Your session is invalid</p>
      <a href="/auth" class="mt-2 text-sm text-link hover:underline">
        Sign in again
      </a>
    </div>
  );
};
