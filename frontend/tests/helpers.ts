import { Page } from '@playwright/test';

/**
 * Marks setup as complete and logs in as admin.
 * Call at the start of any test that needs an authenticated session.
 */
export async function loginAsAdmin(page: Page) {
  await page.goto('/');
  await page.evaluate(() => {
    sessionStorage.setItem('_setupComplete', 'true');
  });
  await page.goto('/login');
  await page.fill('input[id="username"]', 'admin');
  await page.fill('input[id="password"]', 'password123');
  await page.click('button:has-text("Sign In")');
  await page.waitForURL(/\/dashboard/);
}
