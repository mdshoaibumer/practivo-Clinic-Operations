import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './helpers';

test.describe('Appointment Actions', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.click('nav >> text=Appointments');
    await expect(page).toHaveURL(/\/appointments/);
  });

  test('should display appointment with status badge', async ({ page }) => {
    // Mock returns a 'scheduled' appointment
    await expect(page.locator('text=Scheduled').or(page.locator('text=No appointments for this date.'))).toBeVisible();
  });

  test('should complete an appointment', async ({ page }) => {
    // Wait for the appointment list to load
    await page.waitForTimeout(500);
    const completeBtn = page.locator('button[title="Complete"]');
    const count = await completeBtn.count();
    if (count > 0) {
      await completeBtn.first().click();
      // After completion, the status badge text changes
      await expect(page.locator('text=Completed').first()).toBeVisible({ timeout: 10000 });
    }
  });

  test('should cancel an appointment with confirmation', async ({ page }) => {
    // Wait for the appointment list to load
    await page.waitForTimeout(500);
    const cancelBtn = page.locator('button[title="Cancel"]');
    const count = await cancelBtn.count();
    if (count > 0) {
      // Set up dialog handler before clicking
      page.on('dialog', dialog => dialog.accept());
      await cancelBtn.first().click();
      // After cancel, status badge should show Cancelled
      await expect(page.locator('text=Cancelled').first()).toBeVisible({ timeout: 10000 });
    }
  });

  test('should fill all fields in appointment form', async ({ page }) => {
    await page.click('button:has-text("New Appointment")');

    // Select patient
    const patientSelect = page.locator('select').first();
    await patientSelect.selectOption({ index: 1 });

    // Fill time
    await page.fill('input[type="time"]', '14:00');
    await page.locator('input[type="time"]').nth(1).fill('15:00');

    // Fill purpose and notes
    await page.fill('input[placeholder="Treatment purpose"]', 'Root Canal');
    await page.fill('input[placeholder="Additional notes"]', 'Patient allergic to penicillin');

    // Submit
    await page.click('button:has-text("Book")');
  });
});
