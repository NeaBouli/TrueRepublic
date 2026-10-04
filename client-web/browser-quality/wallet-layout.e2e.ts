import { readFileSync } from 'node:fs';
import { test, expect, type Locator, type Page, type TestInfo } from '@playwright/test';

// Every engine covers maintained desktop/tablet/mobile sizes, including 320px.
const widths = [320, 375, 768, 1280];
const origin = 'http://127.0.0.1:4174';

async function contained(locator: Locator) {
  const metrics = await locator.evaluate((element) => {
    const rect = element.getBoundingClientRect();
    const parent = element.parentElement!.getBoundingClientRect();
    return {
      overflow: element.scrollWidth - element.clientWidth,
      clipped: (() => {
        for (let ancestor = element.parentElement; ancestor; ancestor = ancestor.parentElement) {
          const style = getComputedStyle(ancestor);
          const bounds = ancestor.getBoundingClientRect();
          if (['hidden', 'clip'].includes(style.overflowX) && (rect.left < bounds.left - 1 || rect.right > bounds.right + 1)) return true;
          if (['hidden', 'clip'].includes(style.overflowY) && (rect.top < bounds.top - 1 || rect.bottom > bounds.bottom + 1)) return true;
        }
        return false;
      })(),
      left: rect.left - Math.max(0, parent.left),
      right: Math.min(innerWidth, parent.right) - rect.right,
    };
  });
  expect(metrics.overflow).toBeLessThanOrEqual(1);
  expect(metrics.clipped).toBe(false);
  expect(metrics.left).toBeGreaterThanOrEqual(-1);
  expect(metrics.right).toBeGreaterThanOrEqual(-1);
}

async function unobscured(locator: Locator) {
  await locator.scrollIntoViewIfNeeded();
  expect(await locator.evaluate(element => {
    const r = element.getBoundingClientRect();
    const hit = document.elementFromPoint((r.left + r.right) / 2, (r.top + r.bottom) / 2);
    return hit === element || (hit !== null && element.contains(hit));
  })).toBe(true);
}

async function touchTarget(locator: Locator) {
  const box = await locator.boundingBox();
  expect(box!.width).toBeGreaterThanOrEqual(44);
  expect(box!.height).toBeGreaterThanOrEqual(44);
}

async function documentFits(page: Page) {
  expect(await page.evaluate(() => document.documentElement.scrollWidth - innerWidth)).toBeLessThanOrEqual(1);
}

async function record(page: Page, testInfo: TestInfo, name: string) {
  const measurements = await page.evaluate(() => ({
    viewport: { width: innerWidth, height: innerHeight },
    document: { width: document.documentElement.clientWidth, scrollWidth: document.documentElement.scrollWidth },
    elements: [...document.querySelectorAll('header img, header button, code, .font-mono.font-medium, button[aria-label="Copy wallet address"]')].map(element => {
      const rect = element.getBoundingClientRect();
      return {
        tag: element.tagName,
        left: rect.left, right: rect.right, top: rect.top, bottom: rect.bottom,
        width: rect.width, height: rect.height,
        clientWidth: element.clientWidth, scrollWidth: element.scrollWidth,
      };
    }),
  }));
  await testInfo.attach(`${name}-measurements`, { body: JSON.stringify(measurements, null, 2), contentType: 'application/json' });
  await testInfo.attach(name, { body: await page.screenshot({ fullPage: true }), contentType: 'image/png' });
}

for (const width of widths) {
  test(`GH-356 recovery containment, overlap and focus at ${width}px`, async ({ page }, testInfo) => {
    await page.setViewportSize({ width, height: 900 });
    const external: string[] = [];
    await page.route('**/*', async route => {
      if (new URL(route.request().url()).origin !== origin) {
        external.push(route.request().url());
        await route.abort();
      } else await route.continue();
    });
    await page.goto(`${origin}/browser-quality/harness/index.html`);
    await page.getByLabel('Wallet Name').fill('Layout fixture');
    await page.getByLabel('Password', { exact: true }).fill('test-only-password');
    await page.getByLabel('Confirm Password').fill('test-only-password');
    await page.getByRole('button', { name: 'Create Wallet', exact: true }).click();
    await expect(page.getByRole('heading', { name: 'Save Your Recovery Phrase' })).toBeVisible();
    const words = page.locator('.font-mono.font-medium');
    await expect(words).toHaveCount(24);
    for (const word of await words.all()) {
      await contained(word);
      await unobscured(word);
    }
    const boxes = await words.evaluateAll(elements => elements.map(e => {
      const r = e.getBoundingClientRect();
      return { left: r.left, right: r.right, top: r.top, bottom: r.bottom };
    }));
    for (let i = 0; i < boxes.length; i++) {
      for (let j = i + 1; j < boxes.length; j++) {
        const a = boxes[i], b = boxes[j];
        expect(a.right <= b.left || b.right <= a.left || a.bottom <= b.top || b.bottom <= a.top).toBe(true);
      }
    }
    // Onboarding must not be covered by navigation, even midway through scroll.
    await expect(page.getByRole('button', { name: 'Navigation menu' })).toHaveCount(0);
    const confirm = page.getByRole('button', { name: 'I Have Saved My Recovery Phrase' });
    await touchTarget(confirm);
    await unobscured(confirm);
    await confirm.scrollIntoViewIfNeeded();
    await page.evaluate(() => (document.activeElement as HTMLElement)?.blur());
    await page.keyboard.press('Tab');
    await expect(confirm).toBeFocused();
    await documentFits(page);
    await record(page, testInfo, `recovery-${width}`);
    expect(external).toEqual([]);
  });

  test(`GH-356 wallet address, logo and controls at ${width}px`, async ({ page }, testInfo) => {
    await page.setViewportSize({ width, height: 900 });
    const external: string[] = [];
    await page.route('**/*', async route => {
      if (new URL(route.request().url()).origin !== origin) {
        external.push(route.request().url());
        await route.abort();
      } else await route.continue();
    });
    await page.goto(`${origin}/browser-quality/harness/index.html?mode=wallet`);
    await expect(page.getByRole('heading', { name: 'TrueRepublic', exact: true })).toBeVisible();
    const address = page.locator('code').first();
    await contained(address);
    const copy = address.locator('..').getByRole('button');
    await contained(copy);
    const addressBox = await address.boundingBox();
    const copyBox = await copy.boundingBox();
    expect(addressBox!.x + addressBox!.width).toBeLessThanOrEqual(copyBox!.x);
    await touchTarget(copy);
    await unobscured(copy);
    const logo = page.getByRole('img', { name: 'TrueRepublic', exact: true });
    await expect.poll(() => logo.evaluate((e: HTMLImageElement) => e.complete && e.naturalWidth > 0)).toBe(true);
    await contained(logo);
    await unobscured(logo);
    const headerButtons = page.locator('header button');
    for (const button of await headerButtons.all()) {
      await contained(button);
      await touchTarget(button);
      await unobscured(button);
    }
    // FAB precedes the header in DOM; below lg it is the first tab stop.
    const fab = page.getByRole('button', { name: 'Navigation menu' });
    if (width < 1024) {
      await page.keyboard.press('Tab');
      await expect(fab).toBeFocused();
      await touchTarget(fab);
    }
    for (const button of [...await headerButtons.all(), page.getByTitle('Lock wallet'), copy]) {
      await page.keyboard.press('Tab');
      await expect(button).toBeFocused();
    }
    await documentFits(page);
    await record(page, testInfo, `wallet-${width}`);
    expect(external).toEqual([]);
  });
}

test('GH-356 header logo loads locally without a third-party request', async ({ page, request, baseURL }) => {
  const manifest = JSON.parse(readFileSync(new URL('../dist/.vite/manifest.json', import.meta.url), 'utf8'));
  const emittedLogo = manifest['src/assets/logo.png'].file as string;
  expect(manifest['src/components/wallet/WalletDashboard.tsx'].assets).toContain(emittedLogo);
  const asset = await request.get(`${baseURL}/${emittedLogo}`);
  expect(asset.ok()).toBe(true);
  expect(asset.headers()['content-type']).toContain('image/png');
  expect(await asset.body()).toEqual(readFileSync(new URL('../src/assets/logo.png', import.meta.url)));
  const external: string[] = [];
  await page.route('**/*', async route => {
    if (new URL(route.request().url()).origin !== origin) {
      external.push(route.request().url());
      await route.abort();
    } else await route.continue();
  });
  await page.goto(`${origin}/browser-quality/harness/index.html?mode=wallet`);
  const logo = page.getByRole('img', { name: 'TrueRepublic', exact: true });
  await expect.poll(() => logo.evaluate((e: HTMLImageElement) => e.complete && e.naturalWidth > 0)).toBe(true);
  expect(external).toEqual([]);
});
