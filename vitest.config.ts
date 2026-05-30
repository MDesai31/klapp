import { defineConfig } from "vitest/config";
import path from "node:path";

export default defineConfig({
  test: {
    environment: "node",
    include: ["tests/unit/**/*.test.ts", "tests/integration/**/*.test.ts"],
    setupFiles: ["tests/setup.ts"],
    sequence: { concurrent: false },
    maxWorkers: 1,
    minWorkers: 1,
  },
  resolve: {
    alias: { "@": path.resolve(__dirname, "src") },
  },
});
