import { test, expect } from '@playwright/test';

test.describe('Table Info, Column Info & Field Editing', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/?client=local_test&table=CUSTOMERS');
    await page.waitForSelector('table.data-table');
  });

  test('should open table info without error', async ({ page }) => {
    await page.locator('button').filter({ hasText: 'Table Info' }).click();
    await page.waitForTimeout(2000);
    // Should open a floating window
    await expect(page.locator('.floating-window')).toBeVisible();
    // Should NOT contain error text
    const content = await page.locator('.floating-window').textContent();
    expect(content).not.toContain('ORA-00942');
    expect(content).not.toContain('error');
  });

  test('should open column info via info icon', async ({ page }) => {
    // Column headers should have info icons (ℹ)
    const infoIcon = page.locator('table.data-table thead th .col-info-icon').first();
    await expect(infoIcon).toBeVisible();
    await infoIcon.click();
    await expect(page.locator('.floating-window')).toBeVisible();
  });

  test('column info should be read-only (no edit box)', async ({ page }) => {
    const infoIcon = page.locator('table.data-table thead th .col-info-icon').first();
    await infoIcon.click();
    await expect(page.locator('.floating-window')).toBeVisible();
    // Should NOT have textarea or save button
    const textarea = page.locator('.floating-window textarea');
    await expect(textarea).not.toBeVisible();
  });

  test('should open field edit on cell click', async ({ page }) => {
    const cell = page.locator('table.data-table tbody tr:first-child td.clickable-cell').first();
    await cell.click();
    await expect(page.locator('.floating-window')).toBeVisible();
  });

  test('field edit should have textarea and save button', async ({ page }) => {
    const cell = page.locator('table.data-table tbody tr:first-child td.clickable-cell').first();
    await cell.click();
    await expect(page.locator('.floating-window')).toBeVisible();
    await expect(page.locator('.floating-window textarea')).toBeVisible();
    await expect(page.locator('.floating-window button').filter({ hasText: /Save|save/ })).toBeVisible();
  });

  test('should close floating window', async ({ page }) => {
    const cell = page.locator('table.data-table tbody tr:first-child td.clickable-cell').first();
    await cell.click();
    await expect(page.locator('.floating-window')).toBeVisible();

    // Close button
    const closeBtn = page.locator('.floating-window .floating-header button').first();
    await closeBtn.click();
    await expect(page.locator('.floating-window')).not.toBeVisible();
  });

  test('should refresh data on Refresh button', async ({ page }) => {
    await page.locator('button').filter({ hasText: 'Refresh' }).click();
    await page.waitForTimeout(1000);
    await expect(page.locator('table.data-table')).toBeVisible();
    const rows = page.locator('table.data-table tbody tr');
    const count = await rows.count();
    expect(count).toBeGreaterThan(0);
  });
});
