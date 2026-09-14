import { createApiReference } from "@scalar/api-reference";

import "@scalar/api-reference/style.css";

import openapi from "./openapi.yaml?url";

createApiReference("#app", {
  agent: {
    disabled: true,
  },
  telemetry: false,
  url: openapi,
  withDefaultFonts: false,
});
