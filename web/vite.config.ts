import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

// Dev server runs on 8081 because the Go API's default CORS_ORIGINS already
// allows http://localhost:8081 — no server change needed to test against it.
export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    port: 8081,
    host: true,
    strictPort: false,
  },
});
