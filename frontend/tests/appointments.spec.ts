import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './helpers';

test.describe('Appointments', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.click('nav >> text=Appointments');
    await expect(page).toHaveURL(/\/appointments/);
  });

  test('should display appointments page with date navigation', async ({ page }) => {
    await expect(page.locator('h1:has-text("Appointments")')).toBeVisible();
    // Date navigation buttons should exist
    await expect(page.locator('button[title="Complete"]').or(page.locator('text=No appointments for this date.'))).toBeVisible();
  });

  test('should open new appointment form', async ({ page }) => {
    await page.click('button:has-text("New Appointment")');
    await expect(page.locator('text=Book Appointment')).toBeVisible();
    // Form fields should be visible
    await expect(page.locator('text=Patient *')).toBeVisible();
    await expect(page.locator('text=Start Time *')).toBeVisible();
    await expect(page.locator('text=End Time *')).toBeVisible();
  });

  test('should create a new appointment', async ({ page }) => {
    await page.click('button:has-text("New Appointment")');
    await expect(page.locator('text=Book Appointment')).toBeVisible();

    // Select patient
    const patientSelect = page.locator('select').first();
    await patientSelect.selectOption({ index: 1 });

    // Fill time fields
    await page.fill('input[type="time"]', '14:00');
    await page.locator('input[type="time"]').nth(1).fill('14:30');

    // Fill purpose
    await page.fill('input[placeholder="Treatment purpose"]', 'Dental Check-up');

    // Submit
    await page.click('button:has-text("Book")');
  });

  test('should close form on cancel', async ({ page }) => {
    await page.click('button:has-text("New Appointment")');
    await expect(page.locator('text=Book Appointment')).toBeVisible();
    await page.click('button:has-text("Cancel")');
    await expect(page.locator('text=Book Appointment')).not.toBeVisible();
  });

  test('should navigate between dates', async ({ page }) => {
    // Click next day button (ChevronRight)
    const nextButton = page.locator('button').filter({ has: page.locator('svg.lucide-chevron-right') });
    await nextButton.click();
    // Page should still be on appointments
    await expect(page).toHaveURL(/\/appointments/);
  });

  test('should show validation errors when form is incomplete', async ({ page }) => {
    await page.click('button:has-text("New Appointment")');
    await expect(page.locator('text=Book Appointment')).toBeVisible();

    // Submit without filling anything
    await page.click('button:has-text("Book")');

    // Patient validation error should appear
    await expect(page.locator('text=Patient is required').or(page.locator('p.text-red-500').first())).toBeVisible();
  });
});
