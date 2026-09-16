import type { Component } from "solid-js";

export const Forbidden: Component = () => {
  return (
    <div class="flex h-body flex-col items-center justify-center gap-2">
      <span class="font-mono text-6xl font-bold text-ctp-overlay0">403</span>
      <p class="text-ctp-subtext0">Your session is invalid</p>
      <a href="/auth" class="mt-2 text-sm text-ctp-blue hover:underline">
        Sign in again
      </a>
    </div>
  );
};
