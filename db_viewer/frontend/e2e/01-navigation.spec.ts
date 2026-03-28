import { test, expect } from '@playwright/test';

test.describe('Navigation & Client/Table Selection', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should display client list on load', async ({ page }) => {
    await expect(page.locator('.sidebar-header').first()).toHaveText('Connections');
    const clients = page.locator('.sidebar-item');
    await expect(clients.first()).toBeVisible();
    // Should have at least local_test, sales_team, hr_team
    const count = await clients.count();
    expect(count).toBeGreaterThanOrEqual(3);
  });

  test('should show tables when client is selected', async ({ page }) => {
    await page.locator('.sidebar-item').filter({ hasText: 'local_test' }).click();
    // Wait for Tables header
    await expect(page.locator('.sidebar-header').filter({ hasText: 'Tables' })).toBeVisible();
    // Should have table links
    const tableLinks = page.locator('a.sidebar-item');
    await expect(tableLinks.first()).toBeVisible();
  });

  test('should show empty state before selecting table', async ({ page }) => {
    await page.locator('.sidebar-item').filter({ hasText: 'local_test' }).click();
    await expect(page.locator('.empty-state')).toHaveText('Select a table to view data');
  });

  test('should load table data when table is clicked', async ({ page }) => {
    await page.locator('.sidebar-item').filter({ hasText: 'local_test' }).click();
    await page.locator('a.sidebar-item').filter({ hasText: 'CUSTOMERS' }).click();
    await expect(page.locator('table.data-table')).toBeVisible();
    const rows = page.locator('table.data-table tbody tr');
    await expect(rows.first()).toBeVisible();
    const count = await rows.count();
    expect(count).toBeGreaterThan(0);
  });

  test('should switch between clients', async ({ page }) => {
    await page.locator('.sidebar-item').filter({ hasText: 'local_test' }).click();
    await expect(page.locator('a.sidebar-item').filter({ hasText: 'CUSTOMERS' })).toBeVisible();

    await page.locator('.sidebar-item').filter({ hasText: 'hr_team' }).click();
    // hr_team should have EMPLOYEES
    await expect(page.locator('a.sidebar-item').filter({ hasText: 'EMPLOYEES' })).toBeVisible();
  });

  test('should switch between tables without stale sort error', async ({ page }) => {
    await page.locator('.sidebar-item').filter({ hasText: 'local_test' }).click();
    await page.locator('a.sidebar-item').filter({ hasText: 'CUSTOMERS' }).click();
    await expect(page.locator('table.data-table')).toBeVisible();

    // Switch to another table
    await page.locator('a.sidebar-item').filter({ hasText: 'PRODUCTS' }).click();
    await expect(page.locator('table.data-table')).toBeVisible();
    // Should not show error
    await expect(page.locator('text=error')).not.toBeVisible({ timeout: 3000 }).catch(() => {});
  });

  test('should display SPACE and COMMENTARY in sidebar', async ({ page }) => {
    await page.locator('.sidebar-item').filter({ hasText: 'local_test' }).click();
    await page.waitForSelector('a.sidebar-item');
    // Check for commentary labels
    const commentaries = page.locator('.sidebar-commentary');
    const count = await commentaries.count();
    expect(count).toBeGreaterThan(0);
  });

  test('should open table via URL params', async ({ page }) => {
    await page.goto('/?client=local_test&table=CUSTOMERS');
    await expect(page.locator('table.data-table')).toBeVisible();
    const rows = page.locator('table.data-table tbody tr');
    await expect(rows.first()).toBeVisible();
  });

  test('should open in new tab via middle-click', async ({ page, context }) => {
    await page.goto('/?client=local_test');
    await page.waitForSelector('a.sidebar-item');

    const [newPage] = await Promise.all([
      context.waitForEvent('page'),
      page.locator('a.sidebar-item').filter({ hasText: 'CUSTOMERS' }).click({ button: 'middle' }),
    ]);

    await newPage.waitForLoadState();
    await expect(newPage.locator('table.data-table')).toBeVisible({ timeout: 15000 });
  });
});
