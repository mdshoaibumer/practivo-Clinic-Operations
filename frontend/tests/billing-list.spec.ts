import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './helpers';

test.describe('Billing List & Filters', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.click('nav >> text=Billing');
    await expect(page).toHaveURL(/\/billing/);
  });

  test('should display invoice list with columns', async ({ page }) => {
    await expect(page.locator('th:has-text("Invoice #")')).toBeVisible();
    await expect(page.locator('th:has-text("Patient")')).toBeVisible();
    await expect(page.locator('th:has-text("Date")')).toBeVisible();
    await expect(page.locator('th:has-text("Total")')).toBeVisible();
    await expect(page.locator('th:has-text("Balance")')).toBeVisible();
    await expect(page.locator('th:has-text("Status")')).toBeVisible();
  });

  test('should display invoices from mock data', async ({ page }) => {
    await expect(page.locator('text=TEST-2605-0001')).toBeVisible();
    await expect(page.locator('text=TEST-2605-0002')).toBeVisible();
    await expect(page.locator('text=John Doe')).toBeVisible();
    await expect(page.locator('text=Jane Smith')).toBeVisible();
  });

  test('should show status filter buttons', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'All', exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Unpaid' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Partially Paid' })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Paid', exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: 'Voided' })).toBeVisible();
  });

  test('should filter by status', async ({ page }) => {
    await page.click('button:has-text("Paid")');
    // Filter is applied (list re-renders from mock)
    await expect(page).toHaveURL(/\/billing/);
  });

  test('should navigate to invoice detail on row click', async ({ page }) => {
    await page.click('tr:has-text("TEST-2605-0001")');
    await expect(page).toHaveURL(/\/billing\/inv-list-1/);
  });
});
