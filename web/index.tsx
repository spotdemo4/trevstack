/* @refresh reload */

import { Route, Router } from "@solidjs/router";
import { lazy } from "solid-js";
import { render } from "solid-js/web";

import "./index.css";
import Layout from "./layout/layout";

const Home = lazy(() => import("./routes/home"));
const Numbers = lazy(() => import("./routes/numbers"));
const Metrics = lazy(() => import("./routes/metrics"));
const NotFound = lazy(() => import("./routes/404"));

const wrapper = document.getElementById("app");

if (!wrapper) {
  throw new Error("Wrapper div not found");
}

render(
  () => (
    <Router root={Layout}>
      <Route path="/" component={Home} />
      <Route path="/numbers" component={Numbers} />
      <Route path="/metrics" component={Metrics} />
      <Route path="*404" component={NotFound} />
    </Router>
  ),
  wrapper,
);
