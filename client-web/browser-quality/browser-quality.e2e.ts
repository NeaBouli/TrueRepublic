import AxeBuilder from '@axe-core/playwright';
import type { Locator, Page } from '@playwright/test';
import { expect, test } from './fixtures';

const safeRoutes = [
  { path: '/unlock', heading: 'No Wallet Found' },
  { path: '/create', heading: 'Create New Wallet' },
  { path: '/import', heading: 'Import Wallet' },
] as const;

const viewportWidths = [320, 375, 768, 1280] as const;

async function expectVisibleKeyboardFocus(locator: Locator): Promise<void> {
  await expect(locator).toBeFocused();
  const hasVisibleIndicator = await locator.evaluate((element) => {
    const style = window.getComputedStyle(element);
    const outlineVisible =
      style.outlineStyle !== 'none' && Number.parseFloat(style.outlineWidth) > 0;
    return outlineVisible || style.boxShadow !== 'none';
  });
  expect(hasVisibleIndicator).toBe(true);
}

async function pressTabTo(page: Page, locator: Locator): Promise<void> {
  await page.keyboard.press('Tab');
  await expectVisibleKeyboardFocus(locator);
}

for (const route of safeRoutes) {
  test(`${route.path} has no serious or critical accessibility violations`, async ({
    page,
  }) => {
    await page.goto(route.path);
    await expect(page.getByRole('heading', { name: route.heading })).toBeVisible();

    const results = await new AxeBuilder({ page })
      .withTags(['wcag2a', 'wcag2aa', 'wcag21a', 'wcag21aa'])
      .analyze();
    const blockingViolations = results.violations.filter(
      ({ impact }) => impact === 'serious' || impact === 'critical'
    );

    expect(blockingViolations).toEqual([]);
  });
}

test('all safe onboarding controls are keyboard reachable with visible focus', async (
  { page },
  testInfo
) => {
  test.skip(
    testInfo.project.name.includes('mobile'),
    'Physical keyboard coverage runs on desktop engines'
  );

  await page.goto('/unlock');
  await expect(page.getByRole('heading', { name: 'No Wallet Found' })).toBeVisible();
  await pressTabTo(page, page.getByRole('button', { name: 'Create New Wallet' }));
  await pressTabTo(page, page.getByRole('button', { name: 'Import Wallet' }));

  await page.goto('/create');
  await expect(page.getByRole('heading', { name: 'Create New Wallet' })).toBeVisible();
  await pressTabTo(page, page.getByRole('textbox', { name: 'Wallet Name' }));
  await pressTabTo(page, page.getByLabel('Password', { exact: true }));
  await pressTabTo(page, page.getByLabel('Confirm Password'));
  await pressTabTo(page, page.getByRole('button', { name: 'Create Wallet' }));
  await pressTabTo(page, page.getByRole('button', { name: 'Import Existing Wallet' }));

  await page.goto('/import');
  await expect(page.getByRole('heading', { name: 'Import Wallet' })).toBeVisible();
  await pressTabTo(page, page.getByRole('textbox', { name: 'Wallet Name' }));
  await pressTabTo(page, page.getByRole('textbox', { name: 'Recovery Phrase' }));
  await pressTabTo(page, page.getByLabel('Password', { exact: true }));
  await pressTabTo(page, page.getByLabel('Confirm Password'));
  await pressTabTo(page, page.getByRole('button', { name: 'Import Wallet' }));
  await pressTabTo(page, page.getByRole('button', { name: 'Create New Wallet' }));
});

for (const width of viewportWidths) {
  test(`safe pages do not overflow horizontally at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 900 });

    for (const route of safeRoutes) {
      await page.goto(route.path);
      await expect(page.getByRole('heading', { name: route.heading })).toBeVisible();
      const dimensions = await page.evaluate(() => ({
        clientWidth: document.documentElement.clientWidth,
        scrollWidth: document.documentElement.scrollWidth,
      }));
      expect(dimensions.scrollWidth).toBeLessThanOrEqual(dimensions.clientWidth);
    }
  });
}

test('a delayed lazy Create route exposes status then becomes usable', async ({ page }) => {
  let releaseChunk: (() => void) | undefined;
  const chunkReleased = new Promise<void>((resolve) => {
    releaseChunk = resolve;
  });

  await page.route(/\/assets\/CreateWallet-[^/]+\.js$/, async (route) => {
    await chunkReleased;
    await route.continue();
  });

  const navigation = page.goto('/create');
  await expect(page.getByRole('status')).toHaveText('Loading page…');
  releaseChunk?.();
  await navigation;

  await expect(
    page.getByRole('heading', { name: 'Create New Wallet' })
  ).toBeVisible();
  await expect(page.getByRole('textbox', { name: 'Wallet Name' })).toBeEditable();
});
