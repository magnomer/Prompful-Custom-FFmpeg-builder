import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { readFileSync } from "fs";
import { resolve } from "path";

// The UI reads its release version from the repository's version manifest.
// A Go test verifies that the compiled and installer versions remain equal.
const versionConfig = JSON.parse(
  readFileSync(resolve(__dirname, "../version.json"), "utf-8")
);
const programVersion: string = versionConfig["current-version"] ?? "unknown";

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
