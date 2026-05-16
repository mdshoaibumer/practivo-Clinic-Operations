import { test, expect } from '@playwright/test'
import { loginAsAdmin } from './helpers'

test.describe('Clinic Logo Feature', () => {
  test.beforeEach(async ({ page }) => {
    await loginAsAdmin(page)
    await page.click('a[href="/settings"]')
    await page.waitForSelector('text=Clinic Information')
  })

  test('should display logo upload section in clinic settings', async ({ page }) => {
    await expect(page.locator('text=Clinic Logo')).toBeVisible()
    await expect(page.locator('text=Upload a logo to display on printed invoices')).toBeVisible()
    // Should show upload button when no logo
    await expect(page.getByRole('button', { name: /Upload/ }).or(page.locator('span:has-text("Upload")'))).toBeVisible()
  })

  test('should upload a logo and show preview', async ({ page }) => {
    // Set the file input with a test image
    const fileInput = page.locator('input[type="file"][accept*="image/png"]')
    
    // Create a small test PNG file buffer
    const buffer = Buffer.from(
      'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
      'base64'
    )

    await fileInput.setInputFiles({
      name: 'test-logo.png',
      mimeType: 'image/png',
      buffer,
    })

    // Should show success message
    await expect(page.locator('text=Logo uploaded successfully')).toBeVisible()

    // Should show the logo preview image
    const logoImg = page.locator('img[alt="Clinic Logo"]')
    await expect(logoImg).toBeVisible()

    // Should now show "Change" instead of "Upload"
    await expect(page.locator('span:has-text("Change")')).toBeVisible()
  })

  test('should remove an uploaded logo', async ({ page }) => {
    // First upload a logo
    const fileInput = page.locator('input[type="file"][accept*="image/png"]')
    const buffer = Buffer.from(
      'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
      'base64'
    )
    await fileInput.setInputFiles({
      name: 'test-logo.png',
      mimeType: 'image/png',
      buffer,
    })
    await expect(page.locator('img[alt="Clinic Logo"]')).toBeVisible()

    // Click remove button
    await page.click('button[title="Remove logo"]')

    // Should show removal message
    await expect(page.locator('text=Logo removed')).toBeVisible()

    // Logo preview should disappear
    await expect(page.locator('img[alt="Clinic Logo"]')).not.toBeVisible()

    // Should show "Upload" button again
    await expect(page.locator('span:has-text("Upload")')).toBeVisible()
  })

  test('should display logo in invoice print view when logo exists', async ({ page }) => {
    // Upload logo via the settings page first (beforeEach already navigated here)
    const fileInput = page.locator('input[type="file"][accept*="image/png"]')
    const buffer = Buffer.from(
      'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
      'base64'
    )
    await fileInput.setInputFiles({
      name: 'test-logo.png',
      mimeType: 'image/png',
      buffer,
    })
    await expect(page.locator('text=Logo uploaded successfully')).toBeVisible()

    // Navigate to invoice detail via UI
    await page.click('a[href="/billing"]')
    await page.waitForSelector('text=Billing')
    await page.click('text=TEST-2605-0001')
    await page.waitForSelector('h1:has-text("Invoice")')

    // The print view is only rendered when isPrinting is true (after Print button click)
    // Override window.print to prevent actual print dialog
    await page.evaluate(() => { window.print = () => {} })

    // Click the Print button to trigger print view rendering
    await page.click('button:has-text("Print")')

    // Now the print view should be in the DOM with the logo
    const printLogo = page.locator('img[alt="Clinic Logo"]')
    await expect(printLogo).toBeAttached({ timeout: 3000 })
  })
})
