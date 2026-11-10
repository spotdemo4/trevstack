import { getSignInPath, isPublicPath, session } from "$lib/auth";
import { PageLoading } from "$lib/icon";
import { registerUpdates } from "$lib/pwa";
import { AppToaster } from "$lib/toast";
import { createRouter, useLocation, useNavigate } from "@solidjs/router";
import { type JSX, render } from "@solidjs/web";
import { type Component, Loading, Show, createEffect, lazy, onCleanup } from "solid-js";

import "./index.css";
import { Layout } from "./layout/layout";
import { NetworkStatus } from "./layout/network-status";

const Home = lazy(() => import("./routes/home"));
const Forms = lazy(() => import("./routes/forms"));
const Tables = lazy(() => import("./routes/tables"));
const Metrics = lazy(() => import("./routes/metrics"));
const Watch = lazy(() => import("./routes/watch"));
const Track = lazy(() => import("./routes/track"));
const Share = lazy(() => import("./routes/share"));
const Auth = lazy(() => import("./routes/auth"));
const Forbidden = lazy(() => import("./routes/403"));
const NotFound = lazy(() => import("./routes/404"));

const Router = createRouter({
  routes: [
    { path: "/", component: Home },
    { path: "/forms", component: Forms },
    { path: "/tables", component: Tables },
    { path: "/metrics", component: Metrics },
    { path: "/watch", component: Watch },
    { path: "/track", component: Track },
    { path: "/share", component: Share },
    { path: "/auth", component: Auth },
    { path: "/403", component: Forbidden },
    { path: "*404", component: NotFound },
  ],
});

const App: Component<{ children?: JSX.Element }> = (props) => {
  const location = useLocation();
  const navigate = useNavigate();
  const isPublic = () => isPublicPath(location.pathname);
  const hasSession = () => session.claims() !== null;

  createEffect(
    () =>
      isPublic() || hasSession()
        ? null
        : getSignInPath(`${location.pathname}${location.search}${location.hash}`),
    (target) => {
      if (target) navigate(target, { replace: true });
    },
  );

  return (
    <>
      <Show
        when={isPublic()}
        fallback={
          <Show when={hasSession()}>
            <Layout>
              <Loading fallback={<PageLoading />}>{props.children}</Loading>
            </Layout>
          </Show>
        }
      >
        <main class="min-h-viewport" style={{ "--spacing-body": "var(--spacing-viewport)" }}>
          <Loading fallback={<PageLoading />}>{props.children}</Loading>
        </main>
      </Show>
      <NetworkStatus />
      <AppToaster />
    </>
  );
};

const wrapper = document.getElementById("app");

if (!wrapper) {
  throw new Error("Wrapper div not found");
}

const cleanupSession = session.initialize();
registerUpdates();

render(() => {
  onCleanup(cleanupSession);
  return <Router>{(props) => <App>{props.children}</App>}</Router>;
}, wrapper);
