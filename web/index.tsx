/* @refresh reload */

import { Route, Router } from "@solidjs/router";
import { render } from "solid-js/web";

import Home from "./routes/home";

import "./index.css";

const wrapper = document.getElementById("app");

if (!wrapper) {
	throw new Error("Wrapper div not found");
}

render(
	() => (
		<Router>
			<Route path="/" component={Home} />
		</Router>
	),
	wrapper,
);
