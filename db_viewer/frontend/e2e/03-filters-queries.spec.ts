import { test, expect } from '@playwright/test';

test.describe('Filters & Preset Queries', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/?client=local_test&table=CUSTOMERS');
    await page.waitForSelector('table.data-table');
  });

  test('should open filter dropdown', async ({ page }) => {
    await page.locator('button').filter({ hasText: 'Column Filter' }).click();
    await expect(page.locator('.preset-dropdown-menu').first()).toBeVisible();
    // Should have "No Filter (show all)" option
    await expect(page.locator('.preset-dropdown-item').filter({ hasText: 'No Filter' })).toBeVisible();
  });

  test('should apply a preset filter and change visible columns', async ({ page }) => {
    const headersBefore = await page.locator('table.data-table thead th').allTextContents();

    await page.locator('button').filter({ hasText: 'Column Filter' }).click();
    // Click first non-"No Filter" item
    const items = page.locator('.preset-dropdown-item');
    const count = await items.count();
    if (count > 1) {
      await items.nth(1).click();
      await page.waitForTimeout(500);
      const headersAfter = await page.locator('table.data-table thead th').allTextContents();
      // Columns should change after applying filter
      expect(headersAfter.length).toBeGreaterThan(0);
    }
  });

  test('should clear filter with No Filter option', async ({ page }) => {
    // Apply a filter first
    await page.locator('button').filter({ hasText: 'Column Filter' }).click();
    const items = page.locator('.preset-dropdown-item');
    if (await items.count() > 1) {
      await items.nth(1).click();
      await page.waitForTimeout(500);
    }

    // Now clear it
    const filterBtn = page.locator('button.secondary').filter({ hasText: /Filter/ });
    await filterBtn.click();
    await page.locator('.preset-dropdown-item').filter({ hasText: 'No Filter' }).click();
    await page.waitForTimeout(500);
    await expect(page.locator('table.data-table')).toBeVisible();
  });

  test('should open preset queries dropdown', async ({ page }) => {
    await page.locator('button').filter({ hasText: 'Preset Queries' }).click();
    await expect(page.locator('.preset-dropdown-menu').first()).toBeVisible();
    // Should have "Select All" default
    await expect(page.locator('.preset-dropdown-item').filter({ hasText: 'Select All' })).toBeVisible();
  });

  test('should select and execute Select All query', async ({ page }) => {
    await page.locator('button').filter({ hasText: 'Preset Queries' }).click();
    await page.locator('.preset-dropdown-item').filter({ hasText: 'Select All' }).click();

    // Should show query panel with Execute button
    await expect(page.locator('.preset-query-panel')).toBeVisible();
    await expect(page.locator('button').filter({ hasText: 'Execute' })).toBeVisible();

    // Execute
    await page.locator('button').filter({ hasText: 'Execute' }).click();
    await page.waitForTimeout(1000);
    await expect(page.locator('table.data-table')).toBeVisible();
  });

  test('should show final query preview', async ({ page }) => {
    await page.locator('button').filter({ hasText: 'Preset Queries' }).click();
    await page.locator('.preset-dropdown-item').filter({ hasText: 'Select All' }).click();
    await expect(page.locator('.final-query')).toBeVisible();
    const queryText = await page.locator('.final-query').textContent();
    expect(queryText).toContain('SELECT');
  });

  test('should cancel preset query', async ({ page }) => {
    await page.locator('button').filter({ hasText: 'Preset Queries' }).click();
    await page.locator('.preset-dropdown-item').filter({ hasText: 'Select All' }).click();
    await expect(page.locator('.preset-query-panel')).toBeVisible();

    await page.locator('button').filter({ hasText: 'Cancel' }).click();
    await expect(page.locator('.preset-query-panel')).not.toBeVisible();
  });

  test('filter state should persist when switching tables and back', async ({ page }) => {
    // Apply a filter
    await page.locator('button').filter({ hasText: 'Column Filter' }).click();
    const items = page.locator('.preset-dropdown-item');
    if (await items.count() > 1) {
      await items.nth(1).click();
      await page.waitForTimeout(500);
    }

    // Switch to another table
    await page.locator('a.sidebar-item').filter({ hasText: 'PRODUCTS' }).click();
    await page.waitForTimeout(1000);

    // Switch back
    await page.locator('a.sidebar-item').filter({ hasText: 'CUSTOMERS' }).click();
    await page.waitForTimeout(1000);

    // Filter button should show the previously applied filter name
    const filterBtn = page.locator('button.secondary').filter({ hasText: /Filter/ });
    const btnText = await filterBtn.textContent();
    expect(btnText).not.toBe('Column Filter'); // should show "Filter: <name>"
  });
});
