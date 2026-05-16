import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { readFileSync } from "fs";
import { resolve } from "path";

// Read version from wails.json so the developer only needs to edit one file.
const wailsConfig = JSON.parse(
  readFileSync(resolve(__dirname, "../wails.json"), "utf-8")
);
const appVersion: string = wailsConfig.version ?? "unknown";

export default defineConfig({
  plugins: [react()],
  define: {
    __APP_VERSION__: JSON.stringify(appVersion),
  },
  server: {
    host: "127.0.0.1",
    port: 34115,
    strictPort: true,
  },
  preview: {
    host: "127.0.0.1",
    port: 34115,
    strictPort: true,
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
});
