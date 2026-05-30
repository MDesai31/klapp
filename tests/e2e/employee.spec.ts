import { test, expect } from "@playwright/test";

test("employee logs in and sees My Jobs", async ({ page }) => {
  await page.goto("/login");
  await page.fill('input[name="email"]', "manthan@klaus.test");
  await page.fill('input[name="password"]', "employee12345");
  await page.click('button[type="submit"]');
  await expect(page).toHaveURL(/\/jobs/);
  await expect(page.getByRole("heading", { name: "My Jobs" })).toBeVisible();
});
