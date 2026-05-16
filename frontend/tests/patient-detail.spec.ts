import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './helpers';

test.describe('Patient Detail', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/patients/p-fixed');
  });

  test('should display patient personal information', async ({ page }) => {
    await expect(page.locator('h1:has-text("John Doe")')).toBeVisible();
    await expect(page.locator('text=Personal Information')).toBeVisible();

    // Check key fields from mock
    await expect(page.locator('text=9876543210')).toBeVisible();
    await expect(page.locator('text=john@example.com')).toBeVisible();
    await expect(page.locator('text=male')).toBeVisible();
    await expect(page.locator('text=Mumbai')).toBeVisible();
  });

  test('should display medical information', async ({ page }) => {
    await expect(page.locator('text=Medical Information')).toBeVisible();
  });

  test('should display treatment history', async ({ page }) => {
    await expect(page.locator('text=Treatment History')).toBeVisible();
    // Mock returns Root Canal and Cleaning
    await expect(page.locator('text=Root Canal')).toBeVisible();
    await expect(page.locator('text=Cleaning')).toBeVisible();
    // Tooth number
    await expect(page.locator('text=Tooth: 14')).toBeVisible();
  });

  test('should have quick action buttons', async ({ page }) => {
    await expect(page.locator('button:has-text("Create Invoice")')).toBeVisible();
    await expect(page.locator('button:has-text("Book Appointment")')).toBeVisible();
  });

  test('should navigate to create invoice', async ({ page }) => {
    await page.click('button:has-text("Create Invoice")');
    await expect(page).toHaveURL(/\/billing/);
  });

  test('should navigate back to patients list', async ({ page }) => {
    await page.click('button:has(svg.lucide-arrow-left)');
    await expect(page).toHaveURL(/\/patients$/);
  });
});
