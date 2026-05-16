import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './helpers';

test.describe('Reports', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.click('nav >> text=Reports');
    await expect(page).toHaveURL(/\/reports/);
  });

  test('should display reports page with daily report by default', async ({ page }) => {
    await expect(page.locator('h1:has-text("Reports")')).toBeVisible();

    // Daily Report button should be active/default
    await expect(page.locator('button:has-text("Daily Report")')).toBeVisible();
    await expect(page.locator('button:has-text("Monthly Report")')).toBeVisible();

    // Daily report content should load from mock
    await expect(page.locator('text=Daily Collection Report')).toBeVisible();
    await expect(page.locator('text=Total Collection')).toBeVisible();
  });

  test('should display payment details in daily report', async ({ page }) => {
    // Payment table from mock data
    await expect(page.locator('text=Payment Details')).toBeVisible();
    await expect(page.locator('text=TEST-0001')).toBeVisible();
    await expect(page.locator('text=John Doe')).toBeVisible();
    await expect(page.locator('text=Jane Smith')).toBeVisible();
  });

  test('should switch to monthly report', async ({ page }) => {
    await page.click('button:has-text("Monthly Report")');

    // Monthly report content should load
    await expect(page.locator('h2:has-text("Monthly Report")')).toBeVisible();
    await expect(page.locator('text=Total Revenue')).toBeVisible();
    await expect(page.locator('text=Total Invoiced')).toBeVisible();
    await expect(page.locator('text=Outstanding')).toBeVisible();
  });

  test('should switch back to daily report', async ({ page }) => {
    await page.click('button:has-text("Monthly Report")');
    await expect(page.locator('text=Total Revenue')).toBeVisible();

    await page.click('button:has-text("Daily Report")');
    await expect(page.locator('text=Daily Collection Report')).toBeVisible();
  });
});
