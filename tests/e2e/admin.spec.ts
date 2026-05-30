import { test, expect } from "@playwright/test";

test("admin logs in and can open All Jobs", async ({ page }) => {
  await page.goto("/login");
  await page.fill('input[name="email"]', "admin@klaus.test");
  await page.fill('input[name="password"]', "admin12345");
  await page.click('button[type="submit"]');
  await expect(page).toHaveURL(/\/admin|\/jobs/);
  await page.goto("/admin");
  await expect(page.getByRole("heading", { name: "All Jobs" })).toBeVisible();
});
