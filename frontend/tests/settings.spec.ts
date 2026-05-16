import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './helpers';

test.describe('Settings', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.click('nav >> text=Settings');
    await expect(page).toHaveURL(/\/settings/);
  });

  test('should display settings page with clinic tab', async ({ page }) => {
    await expect(page.locator('h1:has-text("Settings")')).toBeVisible();
    await expect(page.locator('text=Clinic Information')).toBeVisible();

    // Should load clinic settings from mock
    const clinicNameInput = page.locator('input').first();
    await expect(clinicNameInput).toHaveValue('Clinmitra Test Clinic');
  });

  test('should save clinic settings', async ({ page }) => {
    const clinicNameInput = page.locator('input').first();
    await clinicNameInput.clear();
    await clinicNameInput.fill('Updated Clinic Name');

    await page.click('button:has-text("Save")');
    await expect(page.locator('text=Settings saved successfully')).toBeVisible();
  });

  test('should switch to treatments tab', async ({ page }) => {
    await page.click('button:has-text("Treatments")');

    // Should show treatment list from mock
    await expect(page.locator('text=Root Canal')).toBeVisible();
    await expect(page.locator('text=Cleaning')).toBeVisible();
  });

  test('should switch to password tab and show form', async ({ page }) => {
    await page.click('button:has-text("Password")');

    // Password form fields should be visible
    await expect(page.locator('input[type="password"]').first()).toBeVisible();
  });

  test('should reject wrong current password', async ({ page }) => {
    await page.click('button:has-text("Password")');

    await page.locator('input[type="password"]').nth(0).fill('wrongpassword');
    await page.locator('input[type="password"]').nth(1).fill('newpass123');
    await page.locator('input[type="password"]').nth(2).fill('newpass123');

    await page.click('button:has-text("Change Password")');
    await expect(page.locator('div.text-red-600, div.bg-red-50').filter({ hasText: /password/i })).toBeVisible();
  });

  test('should switch to backup tab', async ({ page }) => {
    await page.click('button:has-text("Backup")');

    // Should show backup-related UI
    await expect(page.getByRole('button', { name: 'Create Backup' })).toBeVisible();
  });

  test('should navigate between all tabs', async ({ page }) => {
    // Clinic -> Treatments -> Password -> Backup -> Clinic
    await page.click('button:has-text("Treatments")');
    await expect(page.locator('text=Root Canal')).toBeVisible();

    await page.click('button:has-text("Password")');
    await expect(page.locator('input[type="password"]').first()).toBeVisible();

    await page.click('button:has-text("Backup")');
    // Back to Clinic
    await page.click('button:has-text("Clinic")');
    await expect(page.locator('text=Clinic Information')).toBeVisible();
  });
});
