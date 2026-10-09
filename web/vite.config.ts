import { readFileSync } from "node:fs";

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

// The manifest and the pre-paint styles in index.html need literal colors
// before any CSS loads, so they are read from the theme instead of repeated.
const theme = readFileSync(new URL("theme.css", import.meta.url), "utf8");
const [lightTheme = "", darkTheme = ""] = theme.split(":root.dark");

function themeColor(block: string, name: string) {
  const value = block.match(new RegExp(`--color-${name}:\\s*([^;]+);`))?.[1]?.trim();
  if (!value) throw new Error(`theme.css does not set --color-${name} for every theme`);
  return value;
}

const themeColors = {
  THEME_LIGHT_BACKGROUND: themeColor(lightTheme, "background"),
  THEME_LIGHT_PRIMARY: themeColor(lightTheme, "primary"),
  THEME_DARK_BACKGROUND: themeColor(darkTheme, "background"),
  THEME_DARK_PRIMARY: themeColor(darkTheme, "primary"),
};

export default defineConfig({
  plugins: [
    solidPlugin(),
    {
      name: "theme-colors",
      transformIndexHtml: {
        order: "pre",
        handler: (html) =>
          html.replace(/%(THEME_\w+)%/g, (_, key: string) => {
            if (!Object.hasOwn(themeColors, key)) {
              throw new Error(`index.html uses unknown theme color ${key}`);
            }
            return themeColors[key as keyof typeof themeColors];
          }),
      },
    },
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
        theme_color: themeColors.THEME_LIGHT_PRIMARY,
        background_color: themeColors.THEME_LIGHT_BACKGROUND,
        color_scheme_dark: {
          theme_color: themeColors.THEME_DARK_PRIMARY,
          background_color: themeColors.THEME_DARK_BACKGROUND,
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
