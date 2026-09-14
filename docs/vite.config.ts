import { defineConfig } from "vite";

export default defineConfig({
  base: "/docs/",
  build: {
    assetsInlineLimit: 0,
    target: "esnext",
  },
});
