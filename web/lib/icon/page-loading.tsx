import { type Component, onCleanup } from "solid-js";

import { LoaderCircle } from "./icon";

export const PageLoading: Component = () => {
  let indicator: HTMLSpanElement | undefined;

  onCleanup(() => {
    if (!indicator?.isConnected || window.matchMedia("(prefers-reduced-motion: reduce)").matches) {
      return;
    }

    // Solid removes the fallback immediately; retain only its icon for the exit animation.
    const { top, left, width, height } = indicator.getBoundingClientRect();
    const exiting = indicator.cloneNode(true) as HTMLSpanElement;
    exiting.setAttribute("aria-hidden", "true");
    Object.assign(exiting.style, {
      position: "fixed",
      top: `${top}px`,
      left: `${left}px`,
      width: `${width}px`,
      height: `${height}px`,
      pointerEvents: "none",
      zIndex: "50",
    });
    document.body.append(exiting);

    const animation = exiting.animate(
      [
        { opacity: 1, transform: "scale(1)" },
        { opacity: 0, transform: "scale(0.8)" },
      ],
      { duration: 150, easing: "ease-in", fill: "forwards" },
    );
    const remove = () => exiting.remove();
    void animation.finished.then(remove, remove);
  });

  return (
    <div class="flex h-body items-center justify-center" role="status" aria-label="Loading">
      <span ref={(element) => (indicator = element)} class="inline-flex text-ctp-subtext0">
        <LoaderCircle class="animate-spin motion-reduce:animate-none" />
      </span>
    </div>
  );
};
