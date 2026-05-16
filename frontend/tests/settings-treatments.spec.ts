import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './helpers';

test.describe('Settings - Treatment Management', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.click('nav >> text=Settings');
    await expect(page).toHaveURL(/\/settings/);
    await page.click('button:has-text("Treatments")');
    await expect(page.locator('text=Treatment Catalog')).toBeVisible();
  });

  test('should display existing treatments', async ({ page }) => {
    await expect(page.locator('text=Root Canal')).toBeVisible();
    await expect(page.locator('text=Cleaning')).toBeVisible();
  });

  test('should open add treatment form', async ({ page }) => {
    await page.click('button:has-text("Add")');
    await expect(page.locator('input[placeholder="Name *"]')).toBeVisible();
    await expect(page.locator('input[placeholder="Code"]')).toBeVisible();
    await expect(page.locator('input[placeholder="Price (₹)"]')).toBeVisible();
    await expect(page.locator('input[placeholder="Category"]')).toBeVisible();
  });

  test('should add a new treatment', async ({ page }) => {
    await page.click('button:has-text("Add")');
    await page.fill('input[placeholder="Name *"]', 'Tooth Extraction');
    await page.fill('input[placeholder="Code"]', 'EXT');
    await page.fill('input[placeholder="Price (₹)"]', '2000');
    await page.fill('input[placeholder="Category"]', 'Surgery');
    await page.click('button:has-text("Save")');

    // Form should close and new treatment should appear
    await expect(page.locator('input[placeholder="Name *"]')).not.toBeVisible();
    await expect(page.locator('text=Tooth Extraction')).toBeVisible();
  });

  test('should cancel adding a treatment', async ({ page }) => {
    await page.click('button:has-text("Add")');
    await expect(page.locator('input[placeholder="Name *"]')).toBeVisible();
    await page.locator('button:has-text("Cancel")').last().click();
    await expect(page.locator('input[placeholder="Name *"]')).not.toBeVisible();
  });
});
