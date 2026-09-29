import { toaster } from "$lib/toast";
import { registerSW } from "virtual:pwa-register";

// How often an open app checks for a new version.
const CHECK_INTERVAL_MS = 15 * 60 * 1000;
// A failed page load only reloads once in this window, so a chunk that is
// really missing shows its error instead of reloading forever.
const RELOAD_GUARD_MS = 10_000;
const RELOAD_KEY = "stack.chunk-reload";

let updateToastId: number | undefined;

// registerUpdates registers the service worker and checks for new versions
// periodically and whenever the app comes back into view, since an
// installed app can stay open for days. A waiting version is offered in a toast
// rather than reloaded automatically, so a half-finished form is never lost;
// dismissing the toast means "later".
export function registerUpdates() {
  // A tab still running an older build asks for page chunks that newer
  // deploys removed. The user is navigating anyway, so load the new build.
  window.addEventListener("vite:preloadError", (event) => {
    try {
      const last = Number(sessionStorage.getItem(RELOAD_KEY) ?? 0);
      if (Date.now() - last < RELOAD_GUARD_MS) return;
      sessionStorage.setItem(RELOAD_KEY, String(Date.now()));
    } catch {
      // Without storage, reloading could loop, so let the error show.
      return;
    }
    event.preventDefault();
    window.location.reload();
  });

  const applyUpdate = registerSW({
    immediate: true,
    onNeedRefresh: () => {
      if (updateToastId !== undefined) return;
      updateToastId = toaster.info({
        title: "Update available",
        description: "A new version of this app is available.",
        duration: 0,
        action: { label: "Reload", onClick: () => void applyUpdate(true) },
      });
    },
    onRegisteredSW: (_url, registration) => {
      if (!registration) return;
      const check = () => {
        if (document.visibilityState !== "visible" || !navigator.onLine) return;
        if (registration.installing) return;
        // A failed check, such as a server restart, is retried next time.
        registration.update().catch(() => undefined);
      };
      setInterval(check, CHECK_INTERVAL_MS);
      document.addEventListener("visibilitychange", check);
    },
  });
}
