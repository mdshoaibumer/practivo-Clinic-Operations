import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './helpers';

test.describe('Billing - Invoice Form Advanced', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.click('nav >> text=Billing');
    await page.click('button:has-text("New Invoice")');
    await expect(page.getByRole('heading', { name: 'Create Invoice' })).toBeVisible();
  });

  test('should add multiple line items', async ({ page }) => {
    // First item is auto-added when form opens
    await page.fill('input[placeholder="Description"]', 'Root Canal');
    await page.fill('input[placeholder="₹ Price"]', '5000');

    // Add second item
    await page.click('button:has-text("+ Add Item")');
    const descriptions = page.locator('input[placeholder="Description"]');
    await descriptions.nth(1).fill('Cleaning');
    const prices = page.locator('input[placeholder="₹ Price"]');
    await prices.nth(1).fill('1000');

    // Both items should be visible
    await expect(descriptions).toHaveCount(2);
  });

  test('should remove a line item', async ({ page }) => {
    // Add second item
    await page.click('button:has-text("+ Add Item")');
    const descriptions = page.locator('input[placeholder="Description"]');
    await expect(descriptions).toHaveCount(2);

    // Remove second item - the trash buttons are in each grid row
    const trashButtons = page.locator('div.col-span-1 button');
    await trashButtons.last().click();

    await expect(descriptions).toHaveCount(1);
  });

  test('should auto-fill from treatment selection', async ({ page }) => {
    // Select a treatment from the dropdown
    const treatmentSelect = page.locator('select').nth(1); // Second select (first is patient)
    await treatmentSelect.selectOption({ index: 1 }); // Root Canal

    // Description should auto-fill
    const description = page.locator('input[placeholder="Description"]').first();
    await expect(description).toHaveValue('Root Canal');
  });

  test('should fill tooth number', async ({ page }) => {
    await page.fill('input[placeholder="Description"]', 'Root Canal');
    await page.fill('input[placeholder="₹ Price"]', '5000');
    await page.fill('input[placeholder="Tooth"]', '14');

    await expect(page.locator('input[placeholder="Tooth"]').first()).toHaveValue('14');
  });

  test('should show error when no patient selected', async ({ page }) => {
    await page.fill('input[placeholder="Description"]', 'Root Canal');
    await page.fill('input[placeholder="₹ Price"]', '5000');
    await page.click('button:has-text("Create Invoice")');

    await expect(page.locator('text=Please select a patient')).toBeVisible();
  });

  test('should show error when no items added', async ({ page }) => {
    // Select patient
    await page.locator('select').first().selectOption({ index: 1 });

    // Remove the auto-added item
    await page.locator('div.col-span-1 button').click();

    await page.click('button:has-text("Create Invoice")');
    await expect(page.locator('text=Add at least one item')).toBeVisible();
  });

  test('should handle quantity greater than 1', async ({ page }) => {
    await page.fill('input[placeholder="Description"]', 'Composite Filling');
    await page.fill('input[placeholder="₹ Price"]', '2000');

    // Change quantity
    const qtyInput = page.locator('input[type="number"][min="1"]');
    await qtyInput.clear();
    await qtyInput.fill('3');

    await expect(qtyInput).toHaveValue('3');
  });
});
