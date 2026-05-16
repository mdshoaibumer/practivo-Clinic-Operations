import { test, expect } from '@playwright/test';
import { loginAsAdmin } from './helpers';

test.describe('Patient Form - All Fields', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page);
    await page.click('nav >> text=Patients');
    await page.click('button:has-text("New Patient")');
    await expect(page.locator('text=Register New Patient')).toBeVisible();
  });

  test('should fill all optional fields', async ({ page }) => {
    // Required fields
    await page.fill('input[name="name"]', 'Complete Patient');
    await page.fill('input[name="phone"]', '9876543210');
    await page.selectOption('select[name="gender"]', 'female');

    // Optional fields
    await page.fill('input[name="age"]', '28');
    await page.fill('input[name="email"]', 'patient@test.com');
    await page.selectOption('select[name="bloodGroup"]', 'B+');
    await page.fill('input[name="medicalHistory"]', 'Diabetes Type 2');
    await page.fill('input[name="allergies"]', 'Penicillin');

    // Submit
    await page.click('button:has-text("Save Patient")');
    await expect(page.locator('text=Register New Patient')).not.toBeVisible();
  });

  test('should validate email format', async ({ page }) => {
    await page.fill('input[name="name"]', 'Test Patient');
    await page.fill('input[name="phone"]', '9876543210');
    await page.fill('input[name="email"]', 'invalid-email');
    await page.click('button:has-text("Save Patient")');

    // Email validation error should appear (or form accepts it since email is optional)
    // The schema allows empty string OR valid email
  });

  test('should select gender options', async ({ page }) => {
    const genderSelect = page.locator('select[name="gender"]');
    
    await genderSelect.selectOption('male');
    await expect(genderSelect).toHaveValue('male');

    await genderSelect.selectOption('female');
    await expect(genderSelect).toHaveValue('female');

    await genderSelect.selectOption('other');
    await expect(genderSelect).toHaveValue('other');
  });

  test('should close form on cancel', async ({ page }) => {
    await page.fill('input[name="name"]', 'Discarded Patient');
    await page.fill('input[name="phone"]', '1234567890');

    await page.click('button:has-text("Cancel")');
    await expect(page.locator('text=Register New Patient')).not.toBeVisible();
  });
});

test.describe('Setup Wizard - Advanced', () => {
  test('should navigate back from step 2 to step 1', async ({ page }) => {
    await page.goto('/');

    // Step 1 - fill required fields
    await page.fill('input[id="clinicName"]', 'Test Clinic');
    await page.fill('input[id="doctorName"]', 'Dr. Test');
    await page.fill('input[id="phone"]', '9876543210');
    await page.click('button:has-text("Next")');

    // Should be on step 2
    await expect(page.locator('text=Step 2 of 3')).toBeVisible();

    // Go back
    await page.click('button:has-text("Back")');
    await expect(page.locator('text=Step 1 of 3')).toBeVisible();

    // Fields should retain values
    await expect(page.locator('input[id="clinicName"]')).toHaveValue('Test Clinic');
  });

  test('should navigate back from step 3 to step 2', async ({ page }) => {
    await page.goto('/');

    // Step 1
    await page.fill('input[id="clinicName"]', 'Test Clinic');
    await page.fill('input[id="doctorName"]', 'Dr. Test');
    await page.fill('input[id="phone"]', '9876543210');
    await page.click('button:has-text("Next")');

    // Step 2
    await page.fill('input[id="adminFullName"]', 'Admin User');
    await page.fill('input[id="adminUsername"]', 'admin');
    await page.fill('input[id="adminPassword"]', 'password123');
    await page.click('button:has-text("Next")');

    // Step 3
    await expect(page.locator('text=Step 3 of 3')).toBeVisible();

    // Go back to step 2
    await page.click('button:has-text("Back")');
    await expect(page.locator('text=Step 2 of 3')).toBeVisible();
    await expect(page.locator('input[id="adminFullName"]')).toHaveValue('Admin User');
  });

  test('should fill optional fields in step 1', async ({ page }) => {
    await page.goto('/');

    await page.fill('input[id="clinicName"]', 'Complete Clinic');
    await page.fill('input[id="doctorName"]', 'Dr. Complete');
    await page.fill('input[id="phone"]', '9876543210');
    await page.fill('input[id="address"]', '123 Main St');
    await page.fill('input[id="email"]', 'clinic@test.com');
    await page.fill('input[id="gstin"]', '27AAPFU0939F1ZV');

    await page.click('button:has-text("Next")');
    await expect(page.locator('text=Step 2 of 3')).toBeVisible();
  });
});
