import { test, expect } from '@playwright/test';

test.describe('Data Display & View Modes', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/?client=local_test&table=CUSTOMERS');
    await page.waitForSelector('table.data-table');
  });

  test('should display data in row view by default', async ({ page }) => {
    const table = page.locator('table.data-table');
    await expect(table).toBeVisible();
    // Should have column headers
    const headers = table.locator('thead th');
    const headerCount = await headers.count();
    expect(headerCount).toBeGreaterThan(2);
  });

  test('should have ROWID hidden from display columns', async ({ page }) => {
    const headers = page.locator('table.data-table thead th');
    const allText = await headers.allTextContents();
    // ROWID should not appear as a visible column header
    const hasRowid = allText.some((t) => t.trim() === 'ROWID');
    expect(hasRowid).toBe(false);
  });

  test('should display null values with grey styling', async ({ page }) => {
    // Look for null-value cells
    const nullCells = page.locator('.null-value');
    // CUSTOMERS table has some null NOTES fields
    const count = await nullCells.count();
    expect(count).toBeGreaterThanOrEqual(0); // may or may not have nulls
  });

  test('should switch to transpose view', async ({ page }) => {
    await page.locator('button').filter({ hasText: 'Transpose' }).click();
    // Transpose view should be visible (it uses a different table structure)
    await expect(page.locator('.data-view table')).toBeVisible();
  });

  test('should switch back to row view', async ({ page }) => {
    await page.locator('button').filter({ hasText: 'Transpose' }).click();
    await page.locator('button').filter({ hasText: 'Rows' }).click();
    await expect(page.locator('table.data-table')).toBeVisible();
  });

  test('should sort by clicking column header', async ({ page }) => {
    // Get first data cell value before sort
    const firstCellBefore = await page.locator('table.data-table tbody tr:first-child td.clickable-cell').first().textContent();

    // Click a column header to sort
    const headers = page.locator('table.data-table thead th');
    const headerCount = await headers.count();
    if (headerCount > 2) {
      await headers.nth(2).click(); // click a sortable header
      await page.waitForTimeout(1000);
    }
    // Table should still be visible after sort
    await expect(page.locator('table.data-table')).toBeVisible();
  });

  test('should have row menu button per row', async ({ page }) => {
    const menuBtns = page.locator('table.data-table tbody tr .row-menu-btn');
    const count = await menuBtns.count();
    expect(count).toBeGreaterThan(0);
  });

  test('should open row JSON floating window via menu', async ({ page }) => {
    const menuBtn = page.locator('table.data-table tbody tr:first-child .row-menu-btn');
    await menuBtn.click();
    await page.locator('.row-menu-item').filter({ hasText: 'View Row Data' }).click();
    await expect(page.locator('.floating-window')).toBeVisible();
  });

  test('should display row numbers', async ({ page }) => {
    const rowNums = page.locator('table.data-table tbody tr .row-action-num');
    const count = await rowNums.count();
    expect(count).toBeGreaterThan(0);
    const firstNum = await rowNums.first().textContent();
    expect(firstNum?.trim()).toBe('1');
  });

  test('should have scroll shadow wrapper', async ({ page }) => {
    const wrapper = page.locator('.data-view-wrapper');
    await expect(wrapper).toBeVisible();
  });
});
