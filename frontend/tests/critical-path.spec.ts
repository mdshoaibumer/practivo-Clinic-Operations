import { test, expect } from '@playwright/test';

test.describe('Critical Path: Clinic Workflow', () => {
  test.beforeEach(async ({ page }) => {
    // Assuming there is a mock login mechanism or login is not required in mock mode
    // We navigate to the root which redirects to dashboard if logged in, or directly login
    await page.goto('/');
    
    // Check if we are redirected to login
    if (page.url().includes('/login')) {
      await page.fill('input[type="text"]', 'demo_user');
      await page.fill('input[type="password"]', 'password123');
      await page.click('button[type="submit"]');
      await page.waitForURL('/dashboard');
    }
  });

  test('should complete end-to-end clinical workflow', async ({ page }) => {
    // 1. Create Patient
    await page.click('text=Patients');
    await page.waitForURL('**/patients');
    await page.click('button:has-text("New Patient")');
    
    // Fill Patient Form
    const uniquePhone = `9${Math.floor(Math.random() * 1000000000).toString().padStart(9, '0')}`;
    await page.fill('input[name="name"]', 'John Doe E2E');
    await page.fill('input[name="phone"]', uniquePhone);
    await page.selectOption('select[name="gender"]', 'male');
    await page.fill('input[name="age"]', '30');
    await page.click('button:has-text("Save Patient")');
    
    // Verify toast
    await expect(page.locator('text=Patient Created')).toBeVisible();

    // 2. Book Appointment
    await page.click('text=Appointments');
    await page.waitForURL('**/appointments');
    await page.click('button:has-text("New Appointment")');
    
    // Fill Appointment Form
    // Wait for patients to load in the select
    await page.waitForSelector(`select[name="patientId"] option:has-text("John Doe E2E")`, { state: 'attached' });
    const selectLocator = page.locator('select[name="patientId"]');
    await selectLocator.selectOption({ label: `John Doe E2E (${uniquePhone})` });
    
    // Set Time
    await page.fill('input[name="startTime"]', '10:00');
    await page.fill('input[name="endTime"]', '11:00');
    await page.fill('input[name="purpose"]', 'Routine Checkup');
    await page.click('button:has-text("Book")');
    
    // Verify toast
    await expect(page.locator('text=Appointment booked')).toBeVisible();

    // 3. Generate Invoice
    await page.click('text=Billing');
    await page.waitForURL('**/billing');
    await page.click('button:has-text("New Invoice")');
    
    // Select Patient
    const billingSelectLocator = page.locator('select').first(); // First select is patient select
    await billingSelectLocator.selectOption({ label: `John Doe E2E (${uniquePhone})` });
    
    // Add Item (already has one empty item)
    await page.fill('input[placeholder="Description"]', 'Consultation Fee');
    await page.fill('input[placeholder="₹ Price"]', '500');
    
    await page.click('button:has-text("Create Invoice")');
    
    // Confirm dialog
    await page.click('button:has-text("Confirm & Create")');
    
    // Wait to be redirected to invoice detail page
    await page.waitForURL(/\/billing\/[a-zA-Z0-9-]+/);
    
    // Verify on Invoice Detail Page
    await expect(page.locator('text=John Doe E2E')).toBeVisible();
    await expect(page.locator('text=₹500.00')).toBeVisible();
  });
});
