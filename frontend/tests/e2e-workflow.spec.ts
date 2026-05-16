import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './helpers';

test.describe('End-to-End Workflow', () => {
  test('full workflow: register patient → create invoice → record payment', async ({ page }) => {
    await loginAsAdmin(page);

    // --- Step 1: Register a new patient ---
    await page.click('nav >> text=Patients');
    await expect(page).toHaveURL(/\/patients/);

    await page.click('button:has-text("New Patient")');
    await expect(page.locator('text=Register New Patient')).toBeVisible();

    const patientName = `E2E Patient ${Date.now()}`;
    await page.fill('input[name="name"]', patientName);
    await page.fill('input[name="phone"]', '9012345678');
    await page.selectOption('select[name="gender"]', 'male');
    await page.fill('input[name="age"]', '45');
    await page.click('button:has-text("Save Patient")');

    // Form closes
    await expect(page.locator('text=Register New Patient')).not.toBeVisible();
    // Patient appears in the list
    await expect(page.locator(`text=${patientName}`)).toBeVisible();

    // --- Step 2: Create an invoice for the patient ---
    await page.click('nav >> text=Billing');
    await expect(page).toHaveURL(/\/billing/);

    await page.click('button:has-text("New Invoice")');
    await expect(page.getByRole('heading', { name: 'Create Invoice' })).toBeVisible();

    // Select the patient (pick first available)
    await page.locator('select').first().selectOption({ index: 1 });

    // Fill line item
    await page.fill('input[placeholder="Description"]', 'Root Canal Treatment');
    await page.fill('input[placeholder="₹ Price"]', '5000');

    // Create the invoice - shows confirmation dialog
    await page.click('button:has-text("Create Invoice")');
    // Wait for and click confirm button
    await page.waitForSelector('button:has-text("Confirm & Create Invoice")', { timeout: 5000 });
    await page.click('button:has-text("Confirm & Create Invoice")');

    // Should navigate to invoice detail
    await expect(page).toHaveURL(/\/billing\/inv-123/);

    // --- Step 3: Record a payment ---
    await page.click('button:has-text("Record Payment")');
    await expect(page.locator('text=Amount (₹) *')).toBeVisible();

    await page.fill('input[placeholder="Amount in rupees"]', '2500');
    await page.locator('select').last().selectOption('upi');
    await page.fill('input[placeholder="UPI ref / Card last 4"]', 'UPI12345');
    await page.click('button:has-text("Save Payment")');

    // Payment form closes
    await expect(page.locator('text=Amount (₹) *')).not.toBeVisible();
  });

  test('full workflow: book appointment from patient detail', async ({ page }) => {
    await loginAsAdmin(page);

    // Navigate to a patient
    await page.goto('/patients/p-fixed');
    await expect(page.locator('h1:has-text("John Doe")')).toBeVisible();

    // Click "Book Appointment" quick action
    await page.click('button:has-text("Book Appointment")');
    await expect(page).toHaveURL(/\/appointments/);
  });

  test('full workflow: dashboard → patient detail → create invoice', async ({ page }) => {
    await loginAsAdmin(page);

    // Start from dashboard
    await expect(page.locator('h1:has-text("Dashboard")')).toBeVisible();

    // Go to patients
    await page.click('nav >> text=Patients');
    await expect(page).toHaveURL(/\/patients/);

    // Click on a patient row
    await page.click('tr:has-text("Fixed Patient")');
    await expect(page).toHaveURL(/\/patients\/p-fixed/);

    // Use "Create Invoice" quick action
    await page.click('button:has-text("Create Invoice")');
    await expect(page).toHaveURL(/\/billing/);
  });

  test('should persist login across page navigations', async ({ page }) => {
    await loginAsAdmin(page);

    // Navigate around
    await page.click('nav >> text=Patients');
    await page.click('nav >> text=Billing');
    await page.click('nav >> text=Reports');
    await page.click('nav >> text=Dashboard');

    // Still logged in - user name visible
    await expect(page.locator('text=System Admin')).toBeVisible();
    await expect(page.locator('h1:has-text("Dashboard")')).toBeVisible();
  });
});
