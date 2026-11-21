import { defineConfig, loadEnv } from "vite";
import vue from "@vitejs/plugin-vue";
import { fileURLToPath, URL } from "node:url";
import { gatewayProxy } from "./server/dev-proxy.mjs";

export default defineConfig(({ mode }) => ({
  plugins: [vue()],
  resolve: { alias: { "@shared": fileURLToPath(new URL("../shared", import.meta.url)), "@": fileURLToPath(new URL("./src", import.meta.url)) } },
  server: {
    host: "127.0.0.1",
    port: 4312,
    strictPort: true,
    proxy: gatewayProxy({
      ...loadEnv(mode, process.cwd(), ""),
      ...process.env,
    }),
  },
}));
