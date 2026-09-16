import { AuthClient } from "$lib/connect";
import { Code } from "@connectrpc/connect";
import type { JSX } from "@solidjs/web";
import { Effect } from "effect";
import { type Component, Loading, Show, createSignal, onSettled } from "solid-js";

type AuthGateProps = {
  children?: JSX.Element;
};

export const AuthGate: Component<AuthGateProps> = (props) => {
  const [ready, setReady] = createSignal(false);

  onSettled(() => {
    if (isPublicPath(normalizePath(window.location.pathname))) {
      setReady(true);
      return;
    }

    Effect.runFork(
      AuthClient.checkSession({}).pipe(
        Effect.match({
          onSuccess: () => setReady(true),
          onFailure: (error) => {
            if (error.code !== Code.Unauthenticated && error.code !== Code.PermissionDenied) {
              setReady(true);
            }
          },
        }),
      ),
    );
  });

  return (
    <Show when={ready()} fallback={<SessionLoading />}>
      <Loading fallback={<SessionLoading />}>{props.children}</Loading>
    </Show>
  );
};

const SessionLoading: Component = () => (
  <div class="flex h-body items-center justify-center text-sm text-ctp-subtext0">
    Checking session…
  </div>
);

function isPublicPath(path: string) {
  return path === "/auth" || path === "/403";
}

function normalizePath(path: string) {
  return path.toLowerCase().replace(/\/+$/, "") || "/";
}
