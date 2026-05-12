import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import { VitePWA } from "vite-plugin-pwa";
import solidPlugin from "vite-plugin-solid";

export default defineConfig({
  plugins: [
    solidPlugin(),
    tailwindcss(),
    VitePWA({
      registerType: "autoUpdate",
      manifest: {
        name: "TrevStack",
        short_name: "TrevStack",
        description: "TrevStack web client",
        theme_color: "#04a5e5",
        background_color: "#eff1f5",
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
});
