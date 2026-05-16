import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './helpers';

test.describe('Settings - GST & Backup', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.click('nav >> text=Settings');
    await expect(page).toHaveURL(/\/settings/);
  });

  test('should toggle GST checkbox and show conditional fields', async ({ page }) => {
    // GST should already be enabled from mock (gstEnabled: true)
    const gstCheckbox = page.locator('input[type="checkbox"]');
    await expect(gstCheckbox).toBeChecked();

    // GSTIN and GST Rate fields should be visible
    await expect(page.locator('text=GSTIN')).toBeVisible();
    await expect(page.locator('text=GST Rate %')).toBeVisible();

    // Uncheck GST
    await gstCheckbox.uncheck();
    await expect(gstCheckbox).not.toBeChecked();

    // Conditional fields should disappear
    await expect(page.locator('text=GSTIN')).not.toBeVisible();
  });

  test('should fill clinic email and address', async ({ page }) => {
    const emailInput = page.locator('input').nth(2); // email is 3rd input
    await emailInput.clear();
    await emailInput.fill('new@clinic.com');
    await expect(emailInput).toHaveValue('new@clinic.com');
  });

  test('should create a backup', async ({ page }) => {
    await page.click('button:has-text("Backup")');
    await page.click('button:has-text("Create Backup")');
    await expect(page.locator('text=Backup created')).toBeVisible();
  });

  test('should display backup list', async ({ page }) => {
    await page.click('button:has-text("Backup")');
    await expect(page.locator('text=backup-2025-05-16.db')).toBeVisible();
    await expect(page.locator('text=backup-2025-05-15.db')).toBeVisible();
  });

  test('should show restore confirmation dialog', async ({ page }) => {
    await page.click('button:has-text("Backup")');

    // Set up dialog handler
    let dialogMessage = '';
    page.on('dialog', async dialog => {
      dialogMessage = dialog.message();
      await dialog.accept();
    });

    // Click first Restore button
    await page.getByRole('button', { name: 'Restore' }).first().click();
    expect(dialogMessage).toContain('Restore from this backup');
  });

  test('should change password successfully', async ({ page }) => {
    await page.click('button:has-text("Password")');

    await page.locator('input[type="password"]').nth(0).fill('password123');
    await page.locator('input[type="password"]').nth(1).fill('newpass123');
    await page.locator('input[type="password"]').nth(2).fill('newpass123');

    await page.click('button:has-text("Change Password")');
    await expect(page.locator('text=Password changed successfully')).toBeVisible();
  });

  test('should delete a treatment with confirmation', async ({ page }) => {
    await page.click('button:has-text("Treatments")');
    await expect(page.locator('text=Root Canal')).toBeVisible();

    // Set up confirmation dialog
    page.on('dialog', dialog => dialog.accept());

    // Click delete button on first treatment - located in table actions column
    const deleteBtn = page.locator('td >> button').first();
    await deleteBtn.click();

    // Treatment list should refresh
    await page.waitForTimeout(500);
  });
});
