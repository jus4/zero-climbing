import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig, loadEnv } from "vite";

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  const apiTarget = env.API_PROXY_TARGET || "http://localhost:8082";

  return {
    plugins: [react(), tailwindcss()],
    server: {
      proxy: {
        "/api": { target: apiTarget, changeOrigin: true },
        "/proxy": { target: apiTarget, changeOrigin: true },
      },
    },
  };
});
