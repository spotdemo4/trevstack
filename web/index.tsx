import { createRouter } from "@solidjs/router";
import { render } from "@solidjs/web";
import { lazy } from "solid-js";

import "./index.css";
import { Layout } from "./layout/layout";

const Home = lazy(() => import("./routes/home"));
const Numbers = lazy(() => import("./routes/numbers"));
const Metrics = lazy(() => import("./routes/metrics"));
const NotFound = lazy(() => import("./routes/404"));

const Router = createRouter({
  routes: [
    { path: "/", component: Home },
    { path: "/numbers", component: Numbers },
    { path: "/metrics", component: Metrics },
    { path: "*404", component: NotFound },
  ],
});

const wrapper = document.getElementById("app");

if (!wrapper) {
  throw new Error("Wrapper div not found");
}

render(() => <Router>{(props) => <Layout>{props.children}</Layout>}</Router>, wrapper);
