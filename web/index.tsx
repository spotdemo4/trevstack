import { getSignInPath, isPublicPath, session } from "$lib/auth";
import { PageLoading } from "$lib/icon";
import { AppToaster } from "$lib/toast";
import { createRouter, useLocation, useNavigate } from "@solidjs/router";
import { type JSX, render } from "@solidjs/web";
import { type Component, Loading, Show, createEffect, lazy, onCleanup } from "solid-js";

import "./index.css";
import { Layout } from "./layout/layout";
import { NetworkStatus } from "./layout/network-status";

const Home = lazy(() => import("./routes/home"));
const Numbers = lazy(() => import("./routes/numbers"));
const Metrics = lazy(() => import("./routes/metrics"));
const Auth = lazy(() => import("./routes/auth"));
const Forbidden = lazy(() => import("./routes/403"));
const NotFound = lazy(() => import("./routes/404"));

const Router = createRouter({
  routes: [
    { path: "/", component: Home },
    { path: "/numbers", component: Numbers },
    { path: "/metrics", component: Metrics },
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
        <main class="min-h-dvh" style={{ "--spacing-body": "100dvh" }}>
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

render(() => {
  onCleanup(cleanupSession);
  return <Router>{(props) => <App>{props.children}</App>}</Router>;
}, wrapper);
