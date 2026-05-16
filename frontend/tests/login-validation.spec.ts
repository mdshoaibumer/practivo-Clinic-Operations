import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './helpers';

test.describe('Login Form Validation', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.evaluate(() => { sessionStorage.setItem('_setupComplete', 'true'); });
    await page.goto('/login');
  });

  test('should show validation error for short username', async ({ page }) => {
    await page.fill('input[id="username"]', 'ab');
    await page.fill('input[id="password"]', 'password123');
    await page.click('button:has-text("Sign In")');
    await expect(page.locator('text=Username must be at least 3 characters')).toBeVisible();
  });

  test('should show validation error for short password', async ({ page }) => {
    await page.fill('input[id="username"]', 'admin');
    await page.fill('input[id="password"]', '12345');
    await page.click('button:has-text("Sign In")');
    await expect(page.locator('text=Password must be at least 6 characters')).toBeVisible();
  });

  test('should show validation errors for empty form', async ({ page }) => {
    await page.click('button:has-text("Sign In")');
    await expect(page.locator('p.text-red-500').first()).toBeVisible();
  });
});
