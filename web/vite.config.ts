import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";
import solidPlugin from "vite-plugin-solid";

export default defineConfig({
  plugins: [solidPlugin(), tailwindcss()],
  server: {
    port: 3000,
    proxy: {
      "/grpc": {
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
