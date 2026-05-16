import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './helpers';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
  });

  test('should display dashboard statistics', async ({ page }) => {
    await expect(page.locator('h1:has-text("Dashboard")')).toBeVisible();

    // Verify stat cards are present
    await expect(page.getByRole('heading', { name: 'Today\'s Appointments' }).first()).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Total Patients' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Today\'s Revenue' })).toBeVisible();
    await expect(page.getByRole('heading', { name: 'Outstanding Dues' })).toBeVisible();

    // Verify stat values from mock (todayAppointments: 5, totalPatients: 100)
    await expect(page.locator('div.text-2xl.font-bold').first()).toContainText('5');
    await expect(page.locator('div.text-2xl.font-bold').nth(1)).toContainText('100');
  });

  test('should display monthly revenue', async ({ page }) => {
    await expect(page.locator('text=This Month\'s Revenue')).toBeVisible();
    // Mock returns monthRevenue: 15000000 (₹1,50,000.00)
    const revenueCard = page.locator('div.text-3xl.font-bold.text-green-600');
    await expect(revenueCard).toBeVisible();
  });

  test('should show today appointments section', async ({ page }) => {
    // The dashboard has a "Today's Appointments" card
    const apptSection = page.locator('text=No appointments scheduled for today.');
    const apptList = page.locator('div.space-y-2');
    // Either empty message or appointment list should be visible
    const visible = await apptSection.isVisible() || await apptList.isVisible();
    expect(visible).toBeTruthy();
  });
});
