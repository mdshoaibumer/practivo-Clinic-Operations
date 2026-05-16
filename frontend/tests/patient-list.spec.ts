import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './helpers';

test.describe('Patient List Navigation', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.click('nav >> text=Patients');
    await expect(page).toHaveURL(/\/patients/);
  });

  test('should display patient list with table columns', async ({ page }) => {
    await expect(page.locator('th:has-text("Name")')).toBeVisible();
    await expect(page.locator('th:has-text("Phone")')).toBeVisible();
    await expect(page.locator('th:has-text("Gender")')).toBeVisible();
    await expect(page.locator('th:has-text("Age")')).toBeVisible();
    await expect(page.locator('th:has-text("City")')).toBeVisible();
  });

  test('should show fixed patient from mock', async ({ page }) => {
    await expect(page.locator('text=Fixed Patient')).toBeVisible();
    await expect(page.locator('text=1234567890')).toBeVisible();
  });

  test('should navigate to patient detail on row click', async ({ page }) => {
    await page.click('tr:has-text("Fixed Patient")');
    await expect(page).toHaveURL(/\/patients\/p-fixed/);
    await expect(page.locator('h1:has-text("John Doe")')).toBeVisible();
  });

  test('should navigate to patient detail via View button', async ({ page }) => {
    await page.click('button:has-text("View")');
    await expect(page).toHaveURL(/\/patients\/p-fixed/);
  });

  test('should show empty state when search has no results', async ({ page }) => {
    const searchInput = page.locator('input[placeholder="Search by name or phone..."]');
    await searchInput.fill('nonexistent-name-xyz');
    // Wait for debounce
    await page.waitForTimeout(400);
    await expect(page.locator('text=No patients found matching your search')).toBeVisible();
  });
});
