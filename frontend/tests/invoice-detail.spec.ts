import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './helpers';

test.describe('Invoice Detail', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.goto('/billing/inv-123');
  });

  test('should display invoice details', async ({ page }) => {
    await expect(page.locator('text=Invoice TEST-2605-0001')).toBeVisible();

    // Status badge
    await expect(page.locator('text=Partial')).toBeVisible();

    // Patient info
    await expect(page.locator('text=John Doe')).toBeVisible();
    await expect(page.locator('text=9876543210')).toBeVisible();
  });

  test('should display invoice line items', async ({ page }) => {
    await expect(page.locator('text=Items')).toBeVisible();
    await expect(page.locator('text=Root Canal')).toBeVisible();
    // Tooth number
    await expect(page.locator('text=Tooth: 14')).toBeVisible();
  });

  test('should display payment history', async ({ page }) => {
    await expect(page.locator('text=Payment History')).toBeVisible();
    // Mock has one payment of 300000 paise = ₹3,000
    await expect(page.locator('text=cash')).toBeVisible();
  });

  test('should show record payment form', async ({ page }) => {
    await page.click('button:has-text("Record Payment")');
    await expect(page.locator('text=Amount (₹) *')).toBeVisible();
    await expect(page.locator('text=Method')).toBeVisible();
  });

  test('should validate payment amount', async ({ page }) => {
    await page.click('button:has-text("Record Payment")');
    // Click save without entering amount
    await page.click('button:has-text("Save Payment")');
    await expect(page.locator('text=Enter a valid amount')).toBeVisible();
  });

  test('should record a payment', async ({ page }) => {
    await page.click('button:has-text("Record Payment")');
    await page.fill('input[placeholder="Amount in rupees"]', '1000');
    await page.click('button:has-text("Save Payment")');
    // Form should close after successful payment
    await expect(page.locator('text=Amount (₹) *')).not.toBeVisible();
  });

  test('should cancel payment form', async ({ page }) => {
    await page.click('button:has-text("Record Payment")');
    await expect(page.locator('text=Amount (₹) *')).toBeVisible();
    await page.locator('button:has-text("Cancel")').last().click();
    await expect(page.locator('text=Amount (₹) *')).not.toBeVisible();
  });

  test('should navigate back to billing list', async ({ page }) => {
    await page.click('button:has(svg.lucide-arrow-left)');
    await expect(page).toHaveURL(/\/billing$/);
  });
});
