import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './scenarios',
  timeout: 30000,
  retries: 1,
  use: {
    baseURL: 'http://localhost:8080',
    headless: true,
    extraHTTPHeaders: {
      'Content-Type': 'application/json',
    },
  },
  webServer: {
    command: 'JWT_SECRET=dev-secret ADMIN_PASSWORD=admin123 go run ../cmd/server',
    port: 8080,
    reuseExistingServer: true,
    cwd: '..',
  },
});
