import solidPlugin from "@solidjs/vite-plugin";
import tailwindcss from "@tailwindcss/vite";
import { VitePWA } from "vite-plugin-pwa";
import { defineConfig } from "vitest/config";

declare module "vite-plugin-pwa" {
  interface ManifestOptions {
    // https://w3c.github.io/manifest/#color_scheme_dark-member
    // Remove once https://github.com/vite-pwa/vite-plugin-pwa/pull/941 is released.
    color_scheme_dark?: Partial<Pick<ManifestOptions, "theme_color" | "background_color">>;
  }
}

export default defineConfig({
  plugins: [
    solidPlugin(),
    tailwindcss(),
    VitePWA({
      // lib/pwa registers the service worker and asks before updating, so an
      // update never reloads away unsaved work.
      registerType: "prompt",
      injectRegister: false,
      manifest: {
        name: "TrevStack",
        short_name: "TrevStack",
        description: "TrevStack web client",
        theme_color: "#04a5e5",
        background_color: "#eff1f5",
        color_scheme_dark: {
          theme_color: "#89dceb",
          background_color: "#1e1e2e",
        },
        display: "standalone",
        start_url: "/",
        icons: [{ src: "/icons/icon.svg", sizes: "any", type: "image/svg+xml" }],
      },
      workbox: {
        navigateFallbackDenylist: [/^\/grpc/, /^\/docs/],
      },
    }),
  ],
  server: {
    host: true,
    port: 3000,
    proxy: {
      "/grpc": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
      "/docs": {
        target: "http://localhost:8080",
        changeOrigin: true,
      },
    },
  },
  build: {
    target: "esnext",
  },
  resolve: {
    tsconfigPaths: true,
  },
  test: {
    environment: "happy-dom",
    include: ["**/*.test.{ts,tsx}"],
    restoreMocks: true,
    unstubGlobals: true,
  },
});
