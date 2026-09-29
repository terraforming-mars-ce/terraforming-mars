import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import path from "path";
import { assetWatcher } from "./scripts/assets/vite.ts";

export default defineConfig({
  plugins: [react(), assetWatcher()],
  server: {
    port: 3000,
    open: true,
    proxy: {
      "/api": {
        target: "http://localhost:3001",
        changeOrigin: true,
        secure: false,
      },
      "/ws": {
        target: "http://localhost:3001",
        ws: true,
        changeOrigin: true,
      },
    },
  },
  resolve: {
    alias: {
      "@": path.resolve(import.meta.dirname, "./src"),
    },
  },
  build: {
    outDir: "build",
    sourcemap: true,
    rolldownOptions: {
      output: {
        codeSplitting: {
          groups: [
            {
              name: "react-vendor",
              test: /node_modules\/(react|react-dom|scheduler)\//,
              priority: 20,
            },
            {
              name: "three-vendor",
              test: /node_modules\/(three|@react-three\/[^/]+)\//,
              priority: 10,
            },
          ],
        },
      },
    },
  },
  optimizeDeps: {
    include: ["three", "@react-three/fiber", "@react-three/drei"],
  },
});
