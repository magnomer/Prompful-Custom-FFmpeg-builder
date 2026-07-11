import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { readFileSync } from "fs";
import { resolve } from "path";

// version.json at the repo root is the single source of truth for the app
// version (UI, exe metadata, and the CLI all derive from it). Edit that file
// only; build.ps1 mirrors it into wails.json for the Wails/NSIS metadata.
const versionConfig = JSON.parse(
  readFileSync(resolve(__dirname, "../version.json"), "utf-8")
);
const programVersion: string = versionConfig.version ?? "unknown";

export default defineConfig({
  base: "./",
  plugins: [react()],
  define: {
    __PROGRAM_VERSION__: JSON.stringify(programVersion),
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
