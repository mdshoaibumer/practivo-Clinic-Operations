import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './helpers';

test.describe('Navigation & Auth Guards', () => {
  test('should redirect to setup wizard when not configured', async ({ page }) => {
    await page.goto('/');
    // Without setup complete, should redirect to /setup
    await expect(page).toHaveURL(/\/setup/);
  });

  test('should redirect to login when setup is complete but not authenticated', async ({ page }) => {
    await page.goto('/');
    await page.evaluate(() => {
      sessionStorage.setItem('_setupComplete', 'true');
    });
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/login/);
  });

  test('should redirect unknown routes to dashboard when authenticated', async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/nonexistent-page');
    await expect(page).toHaveURL(/\/dashboard/);
  });

  test('should navigate through all sidebar links', async ({ page }) => {
    await loginAsAdmin(page);

    // Dashboard
    await page.click('nav >> text=Dashboard');
    await expect(page).toHaveURL(/\/dashboard/);
    await expect(page.locator('h1:has-text("Dashboard")')).toBeVisible();

    // Patients
    await page.click('nav >> text=Patients');
    await expect(page).toHaveURL(/\/patients/);
    await expect(page.locator('h1:has-text("Patients")')).toBeVisible();

    // Billing
    await page.click('nav >> text=Billing');
    await expect(page).toHaveURL(/\/billing/);
    await expect(page.locator('h1:has-text("Billing")')).toBeVisible();

    // Appointments
    await page.click('nav >> text=Appointments');
    await expect(page).toHaveURL(/\/appointments/);
    await expect(page.locator('h1:has-text("Appointments")')).toBeVisible();

    // Reports
    await page.click('nav >> text=Reports');
    await expect(page).toHaveURL(/\/reports/);
    await expect(page.locator('h1:has-text("Reports")')).toBeVisible();

    // Settings
    await page.click('nav >> text=Settings');
    await expect(page).toHaveURL(/\/settings/);
    await expect(page.locator('h1:has-text("Settings")')).toBeVisible();
  });

  test('should logout and redirect to login', async ({ page }) => {
    await loginAsAdmin(page);
    await page.click('button:has-text("Logout")');
    await expect(page).toHaveURL(/\/login/);
  });

  test('should display user name in sidebar', async ({ page }) => {
    await loginAsAdmin(page);
    await expect(page.locator('text=System Admin')).toBeVisible();
  });

  test('should toggle sidebar collapse', async ({ page }) => {
    await loginAsAdmin(page);
    // The sidebar should show "Clinmitra Dental" title initially
    await expect(page.locator('text=Clinmitra Dental')).toBeVisible();

    // Click the collapse button (ChevronLeft icon button)
    const collapseBtn = page.locator('aside button').first();
    await collapseBtn.click();

    // After collapse, the title should be hidden
    await expect(page.locator('h1:has-text("Clinmitra Dental")')).not.toBeVisible();
  });
});
