import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests/e2e",
  fullyParallel: false,
  retries: process.env.CI ? 2 : 0,
  reporter: "list",
  use: {
    baseURL: "http://localhost:3100",
    trace: "retain-on-failure",
  },
  projects: [
    { name: "desktop", use: { ...devices["Desktop Chrome"] } },
    { name: "mobile", use: { ...devices["Pixel 7"] } },
  ],
  webServer: {
    command: "DEMO_MODE=true pnpm build && cp -R public .next/standalone/apps/web/public && mkdir -p .next/standalone/apps/web/.next && cp -R .next/static .next/standalone/apps/web/.next/static && DEMO_MODE=true HOSTNAME=localhost PORT=3100 node .next/standalone/apps/web/server.js",
    url: "http://localhost:3100",
    reuseExistingServer: !process.env.CI,
    timeout: 120_000,
  },
});
