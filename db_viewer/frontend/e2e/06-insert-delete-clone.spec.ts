import { test, expect } from '@playwright/test';

test.describe('Insert, Delete, Clone Row', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/?client=local_test&table=CUSTOMERS');
    await page.waitForSelector('table.data-table');
  });

  // ─── INSERT ────────────────────────────────────────────────────────────────

  test('should open Insert New Row floating window', async ({ page }) => {
    await page.locator('button').filter({ hasText: '➕ Insert' }).click();
    await expect(page.locator('.floating-window')).toBeVisible();
    await expect(page.locator('.floating-header .title')).toContainText('Insert New Row');
  });

  test('insert form should show all table columns', async ({ page }) => {
    await page.locator('button').filter({ hasText: '➕ Insert' }).click();
    await expect(page.locator('.insert-form')).toBeVisible();
    const fields = page.locator('.insert-field-card');
    const count = await fields.count();
    expect(count).toBeGreaterThan(0);
  });

  test('should be able to type in insert form fields', async ({ page }) => {
    await page.locator('button').filter({ hasText: '➕ Insert' }).click();
    await expect(page.locator('.insert-form')).toBeVisible();

    const firstInput = page.locator('.insert-input').first();
    await firstInput.click();
    await firstInput.fill('TEST_VALUE_123');
    await expect(firstInput).toHaveValue('TEST_VALUE_123');
  });

  test('should update INSERT query preview when typing', async ({ page }) => {
    await page.locator('button').filter({ hasText: '➕ Insert' }).click();
    await expect(page.locator('.insert-form')).toBeVisible();

    const firstInput = page.locator('.insert-input').first();
    await firstInput.fill('PREVIEW_TEST');
    // Wait for debounced preview (300ms)
    await page.waitForTimeout(600);
    const preview = page.locator('.insert-query-sql');
    const text = await preview.textContent();
    expect(text).not.toBe('-- fill in values above --');
    expect(text).toContain('INSERT');
  });

  test('should insert a new row and reload table', async ({ page }) => {
    // Count rows before
    const rowsBefore = await page.locator('table.data-table tbody tr').count();

    await page.locator('button').filter({ hasText: '➕ Insert' }).click();
    await expect(page.locator('.insert-form')).toBeVisible();

    // Fill required fields with valid values matching CUSTOMERS schema
    const ts = Date.now();
    const fieldValues: Record<string, string> = {
      CUSTOMER_ID: String(90000 + (ts % 9000)),
      FIRST_NAME: 'TestFirst',
      LAST_NAME: 'TestLast',
      EMAIL: `test_${ts}@example.com`,
    };
    for (const [col, val] of Object.entries(fieldValues)) {
      const card = page.locator('.insert-field-card').filter({ has: page.locator(`.insert-field-name:text-is("${col}")`) });
      if (await card.count() > 0) {
        await card.locator('.insert-input').fill(val);
      }
    }

    await page.locator('button').filter({ hasText: 'Insert Row' }).click();
    await expect(page.locator('.insert-message.success')).toBeVisible({ timeout: 5000 });

    // Window should close and table reload
    await expect(page.locator('.floating-window')).not.toBeVisible({ timeout: 5000 });
    await page.waitForSelector('table.data-table');
    const rowsAfter = await page.locator('table.data-table tbody tr').count();
    expect(rowsAfter).toBeGreaterThanOrEqual(rowsBefore);
  });

  test('should close insert window without inserting', async ({ page }) => {
    await page.locator('button').filter({ hasText: '➕ Insert' }).click();
    await expect(page.locator('.floating-window')).toBeVisible();
    await page.locator('.floating-header button').filter({ hasText: '✕' }).click();
    await expect(page.locator('.floating-window')).not.toBeVisible();
  });

  // ─── DELETE ────────────────────────────────────────────────────────────────

  test('should open Delete Row confirmation window', async ({ page }) => {
    // Open row menu on first row
    await page.locator('.row-menu-btn').first().click();
    await expect(page.locator('.row-menu-dropdown')).toBeVisible();
    await page.locator('.row-menu-item-danger').filter({ hasText: 'Delete Row' }).click();

    await expect(page.locator('.floating-window')).toBeVisible();
    await expect(page.locator('.floating-header .title')).toContainText('Delete Row');
  });

  test('delete confirm window should show row data and DELETE query', async ({ page }) => {
    await page.locator('.row-menu-btn').first().click();
    await page.locator('.row-menu-item-danger').filter({ hasText: 'Delete Row' }).click();

    await expect(page.locator('.delete-confirm')).toBeVisible();
    await expect(page.locator('.delete-warning')).toContainText('Are you sure');
    await expect(page.locator('.delete-row-summary')).toBeVisible();

    // Wait for DELETE query to load
    await page.waitForTimeout(1000);
    const queryText = await page.locator('.delete-query-sql').textContent();
    expect(queryText).toContain('DELETE');
  });

  test('should delete a row and reload table', async ({ page }) => {
    // First insert a row with no FK children so we can safely delete it
    const ts = Date.now();
    await page.locator('button').filter({ hasText: '➕ Insert' }).click();
    await expect(page.locator('.insert-form')).toBeVisible();
    const insertValues: Record<string, string> = {
      CUSTOMER_ID: String(70000 + (ts % 9000)),
      FIRST_NAME: 'DeleteMe',
      LAST_NAME: 'TestRow',
      EMAIL: `deleteme_${ts}@example.com`,
    };
    for (const [col, val] of Object.entries(insertValues)) {
      const card = page.locator('.insert-field-card').filter({ has: page.locator(`.insert-field-name:text-is("${col}")`) });
      if (await card.count() > 0) await card.locator('.insert-input').fill(val);
    }
    await page.locator('button').filter({ hasText: 'Insert Row' }).click();
    await expect(page.locator('.insert-message.success')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('.floating-window')).not.toBeVisible({ timeout: 5000 });
    await page.waitForSelector('table.data-table');

    // Sort by CUSTOMER_ID desc to find our newly inserted row at top
    // Instead, find the row with our email by scrolling — use last row (highest ID after sort)
    // Reload and find the inserted row via sort descending on CUSTOMER_ID
    const rowsBefore = await page.locator('table.data-table tbody tr').count();

    // Click CUSTOMER_ID header twice to sort descending so our new row is first
    const headers = page.locator('table.data-table thead th');
    const headerCount = await headers.count();
    for (let i = 0; i < headerCount; i++) {
      const text = await headers.nth(i).textContent();
      if (text && text.includes('CUSTOMER_ID')) {
        await headers.nth(i).click(); // asc
        await page.waitForTimeout(500);
        await headers.nth(i).click(); // desc
        await page.waitForTimeout(500);
        break;
      }
    }

    // Now first row should be our inserted row (highest ID)
    await page.locator('.row-menu-btn').first().click();
    await page.locator('.row-menu-item-danger').filter({ hasText: 'Delete Row' }).click();
    await expect(page.locator('.delete-confirm')).toBeVisible();

    await page.locator('.delete-btn-confirm').click();
    await expect(page.locator('.delete-message.success')).toBeVisible({ timeout: 5000 });

    // Window should close and table reload
    await expect(page.locator('.floating-window')).not.toBeVisible({ timeout: 5000 });
    await page.waitForSelector('table.data-table');
    const rowsAfter = await page.locator('table.data-table tbody tr').count();
    expect(rowsAfter).toBeLessThan(rowsBefore);
  });

  test('should close delete window without deleting', async ({ page }) => {
    await page.locator('.row-menu-btn').first().click();
    await page.locator('.row-menu-item-danger').filter({ hasText: 'Delete Row' }).click();
    await expect(page.locator('.floating-window')).toBeVisible();
    await page.locator('.floating-header button').filter({ hasText: '✕' }).click();
    await expect(page.locator('.floating-window')).not.toBeVisible();
  });

  // ─── CLONE ─────────────────────────────────────────────────────────────────

  test('should open Clone Row floating window', async ({ page }) => {
    await page.locator('.row-menu-btn').first().click();
    await page.locator('.row-menu-item').filter({ hasText: 'Clone Row' }).click();

    await expect(page.locator('.floating-window')).toBeVisible();
    await expect(page.locator('.floating-header .title')).toContainText('Clone Row');
  });

  test('clone form should be pre-filled with row values', async ({ page }) => {
    // Get first row's first visible cell value
    const firstCellValue = await page.locator('table.data-table tbody tr').first()
      .locator('td.clickable-cell').first().textContent();

    await page.locator('.row-menu-btn').first().click();
    await page.locator('.row-menu-item').filter({ hasText: 'Clone Row' }).click();
    await expect(page.locator('.insert-form')).toBeVisible();

    // At least one input should have a non-empty value (pre-filled)
    const inputs = page.locator('.insert-input');
    const inputCount = await inputs.count();
    let hasPrefilledValue = false;
    for (let i = 0; i < inputCount; i++) {
      const val = await inputs.nth(i).inputValue();
      if (val && val.trim() !== '') {
        hasPrefilledValue = true;
        break;
      }
    }
    expect(hasPrefilledValue).toBe(true);
  });

  test('clone form should show all table columns', async ({ page }) => {
    await page.locator('.row-menu-btn').first().click();
    await page.locator('.row-menu-item').filter({ hasText: 'Clone Row' }).click();
    await expect(page.locator('.insert-form')).toBeVisible();

    const fields = page.locator('.insert-field-card');
    const count = await fields.count();
    expect(count).toBeGreaterThan(0);
  });

  test('should be able to edit cloned row fields before inserting', async ({ page }) => {
    await page.locator('.row-menu-btn').first().click();
    await page.locator('.row-menu-item').filter({ hasText: 'Clone Row' }).click();
    await expect(page.locator('.insert-form')).toBeVisible();

    const firstInput = page.locator('.insert-input').first();
    await firstInput.triple_click?.() || await firstInput.click({ clickCount: 3 });
    await firstInput.fill('CLONED_EDIT_VALUE');
    await expect(firstInput).toHaveValue('CLONED_EDIT_VALUE');
  });

  test('should insert cloned row and reload table', async ({ page }) => {
    const rowsBefore = await page.locator('table.data-table tbody tr').count();

    await page.locator('.row-menu-btn').first().click();
    await page.locator('.row-menu-item').filter({ hasText: 'Clone Row' }).click();
    await expect(page.locator('.insert-form')).toBeVisible();
    // Wait for prefill to complete
    await page.waitForTimeout(300);

    // Override unique fields with valid typed values to avoid constraint violations
    const ts = Date.now();
    const overrides: Record<string, string> = {
      CUSTOMER_ID: String(60000 + (ts % 9000)),
      EMAIL: `clone_${ts}@example.com`,
    };
    for (const [col, val] of Object.entries(overrides)) {
      const card = page.locator('.insert-field-card').filter({ has: page.locator(`.insert-field-name:text-is("${col}")`) });
      if (await card.count() > 0) {
        await card.locator('.insert-input').fill(val);
      }
    }

    await page.locator('button').filter({ hasText: 'Insert Row' }).click();
    await expect(page.locator('.insert-message.success')).toBeVisible({ timeout: 5000 });

    await expect(page.locator('.floating-window')).not.toBeVisible({ timeout: 5000 });
    await page.waitForSelector('table.data-table');
    const rowsAfter = await page.locator('table.data-table tbody tr').count();
    expect(rowsAfter).toBeGreaterThanOrEqual(rowsBefore);
  });

  // ─── MULTIPLE WINDOWS ──────────────────────────────────────────────────────

  test('should support multiple floating windows open simultaneously', async ({ page }) => {
    // Open insert window
    await page.locator('button').filter({ hasText: '➕ Insert' }).click();
    await expect(page.locator('.floating-window')).toHaveCount(1);

    // Open delete window
    await page.locator('.row-menu-btn').first().click();
    await page.locator('.row-menu-item-danger').filter({ hasText: 'Delete Row' }).click();
    await expect(page.locator('.floating-window')).toHaveCount(2);

    // Close both
    const closeButtons = page.locator('.floating-header button').filter({ hasText: '✕' });
    await closeButtons.first().click();
    await expect(page.locator('.floating-window')).toHaveCount(1);
    await closeButtons.first().click();
    await expect(page.locator('.floating-window')).toHaveCount(0);
  });
});
