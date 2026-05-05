/* @refresh reload */

import { Route, Router } from "@solidjs/router";
import { render } from "solid-js/web";

import "./index.css";
import Layout from "./layout/Layout";
import Home from "./routes/home";
import Metrics from "./routes/metrics";
import Numbers from "./routes/numbers";

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
    </Router>
  ),
  wrapper,
);
