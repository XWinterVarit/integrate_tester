import { test, expect } from '@playwright/test';

test.describe('Multi-Client Support', () => {
  test('sales_team client should show its tables', async ({ page }) => {
    await page.goto('/');
    await page.locator('.sidebar-item').filter({ hasText: 'sales_team' }).click();
    await expect(page.locator('.sidebar-header').filter({ hasText: 'Tables' })).toBeVisible();
    await expect(page.locator('a.sidebar-item').filter({ hasText: 'CUSTOMERS' })).toBeVisible();
    await expect(page.locator('a.sidebar-item').filter({ hasText: 'ORDERS' })).toBeVisible();
  });

  test('hr_team client should show its tables', async ({ page }) => {
    await page.goto('/');
    await page.locator('.sidebar-item').filter({ hasText: 'hr_team' }).click();
    await expect(page.locator('.sidebar-header').filter({ hasText: 'Tables' })).toBeVisible();
    await expect(page.locator('a.sidebar-item').filter({ hasText: 'EMPLOYEES' })).toBeVisible();
  });

  test('sales_team ORDERS table should load data', async ({ page }) => {
    await page.goto('/?client=sales_team&table=ORDERS');
    await expect(page.locator('table.data-table')).toBeVisible();
    const rows = page.locator('table.data-table tbody tr');
    await expect(rows.first()).toBeVisible();
  });

  test('hr_team EMPLOYEES table should load data', async ({ page }) => {
    await page.goto('/?client=hr_team&table=EMPLOYEES');
    await expect(page.locator('table.data-table')).toBeVisible();
    const rows = page.locator('table.data-table tbody tr');
    await expect(rows.first()).toBeVisible();
  });

  test('different clients should have different filter presets', async ({ page }) => {
    // Check local_test CUSTOMERS filters
    await page.goto('/?client=local_test&table=CUSTOMERS');
    await page.waitForSelector('table.data-table');
    await page.locator('button').filter({ hasText: 'Column Filter' }).click();
    const localItems = await page.locator('.preset-dropdown-item').allTextContents();
    await page.keyboard.press('Escape');

    // Check sales_team CUSTOMERS filters
    await page.goto('/?client=sales_team&table=CUSTOMERS');
    await page.waitForSelector('table.data-table');
    await page.locator('button').filter({ hasText: 'Column Filter' }).click();
    const salesItems = await page.locator('.preset-dropdown-item').allTextContents();

    // They should have different filter options
    expect(localItems.join()).not.toBe(salesItems.join());
  });
});
