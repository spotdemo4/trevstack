import { toaster } from "$lib/toast";
import { type Component, onSettled } from "solid-js";

export const NetworkStatus: Component = () => {
  let offlineToastId: number | undefined;
  let wasOffline = false;

  const handleOffline = () => {
    wasOffline = true;
    if (offlineToastId !== undefined) return;

    offlineToastId = toaster.warning({
      title: "Offline",
      description: "You’re viewing cached app content. Live data and changes may be unavailable.",
      duration: 0,
    });
  };

  const handleOnline = () => {
    if (!wasOffline) return;

    wasOffline = false;
    if (offlineToastId !== undefined) {
      toaster.dismiss(offlineToastId);
      offlineToastId = undefined;
    }

    toaster.success({
      title: "Back online",
      description: "Your connection has been restored.",
    });
  };

  onSettled(() => {
    window.addEventListener("offline", handleOffline);
    window.addEventListener("online", handleOnline);

    if (!navigator.onLine) handleOffline();

    return () => {
      window.removeEventListener("offline", handleOffline);
      window.removeEventListener("online", handleOnline);
      if (offlineToastId !== undefined) toaster.dismiss(offlineToastId);
    };
  });

  return null;
};
