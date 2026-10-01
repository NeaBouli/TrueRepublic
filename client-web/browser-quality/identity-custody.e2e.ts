import { mkdirSync, readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import process from 'node:process';
import type { Page } from '@playwright/test';
import { expect, test } from './fixtures';
import { previewMockIdentityHash } from '../src/services/previewIdentityHash';

// GH-309 identity custody visual contract (GH309B2D1). Each custody state of
// /identity is reached through the real UI with an ephemeral local test
// wallet and synthetic, pre-seeded storage only. The page must stay inside
// its layout, keep every control >= 44x44 CSS px with text contrast >= 4.5,
// keep disabled controls inert and out of the tab order, expose accessible
// names/descriptions/status, never show or log the identity secret, never
// use the clipboard and never send anything but the canonical Comet `status`
// bootstrap call (answered as a controlled error without data).

const RPC_ORIGIN = 'http://localhost:26657';
const EVIDENCE_DIR = resolve(process.cwd(), 'node_modules/.cache/gh309b2d1-custody-evidence');
const PASSWORD = 'gh309-custody-contract';
const SECRET = 'c4'.repeat(32);
const IDENTITY = {
  secret: SECRET,
  commitment: previewMockIdentityHash(SECRET),
  nullifier: previewMockIdentityHash(`${SECRET}00`),
  createdAt: 1_700_000_000_000,
};
const LEGACY_RAW = `{"state":{"identity":${JSON.stringify(IDENTITY)}},"version":0}`;
const QUARANTINED_RAW = LEGACY_RAW.replace(IDENTITY.nullifier, 'f'.repeat(64));

const viewports = [
  { name: 'desktop-1440x1000', width: 1440, height: 1000 },
  { name: 'tablet-landscape-1180x820', width: 1180, height: 820 },
  { name: 'tablet-portrait-820x1180', width: 820, height: 1180 },
  { name: 'mobile-390x844', width: 390, height: 844 },
] as const;

type CustodyState = 'locked' | 'absent' | 'legacy-pending' | 'quarantined' | 'error' | 'ready';

const STATES: Record<CustodyState, { seed: Record<string, string>; heading: string; wallet: boolean }> = {
  locked: { seed: {}, heading: 'Preview Identity Locked', wallet: false },
  absent: { seed: {}, heading: 'No Preview Identity', wallet: true },
  'legacy-pending': { seed: { 'identity-store': LEGACY_RAW }, heading: 'Unencrypted Preview Identity Found', wallet: true },
  quarantined: { seed: { 'identity-store': QUARANTINED_RAW }, heading: 'Preview Identity Not Used', wallet: true },
  error: { seed: { truerepublic_identity_vault_v1: 'not json' }, heading: 'Preview Identity Unavailable', wallet: true },
  ready: { seed: { 'identity-store': LEGACY_RAW }, heading: 'Preview Identity', wallet: true },
};

function isCanonicalStatus(body: string): boolean {
  try {
    const parsed = JSON.parse(body) as Record<string, unknown>;
    return (
      Object.keys(parsed).sort().join(',') === 'id,jsonrpc,method,params' &&
      parsed.jsonrpc === '2.0' &&
      Number.isSafeInteger(parsed.id) &&
      parsed.method === 'status' &&
      typeof parsed.params === 'object' &&
      parsed.params !== null &&
      !Array.isArray(parsed.params) &&
      Object.keys(parsed.params).length === 0
    );
  } catch {
    return false;
  }
}

async function guardRpc(page: Page, ledger: { status: number; violations: string[] }): Promise<void> {
  // Registered after the network guard, so it takes precedence for the RPC origin.
  await page.route(`${RPC_ORIGIN}/**`, async (route, request) => {
    const body = request.postData() ?? '';
    if (request.method() === 'POST' && new URL(request.url()).pathname === '/' && !body.includes('broadcast_tx') && isCanonicalStatus(body)) {
      ledger.status += 1;
      const id = (JSON.parse(body) as { id: number }).id;
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ jsonrpc: '2.0', id, error: { code: -32603, message: 'GH-309 custody contract: no chain data' } }),
      });
      return;
    }
    ledger.violations.push(`${request.method()} ${request.url()} ${body.slice(0, 200)}`);
    await route.fulfill({ status: 500, body: 'forbidden RPC request in GH-309 custody contract' });
  });
}

async function reachState(page: Page, state: CustodyState): Promise<void> {
  const { seed, heading, wallet } = STATES[state];
  // Runs after the fixture's storage reset on every navigation.
  await page.addInitScript((entries: Record<string, string>) => {
    for (const [key, value] of Object.entries(entries)) window.localStorage.setItem(key, value);
    const calls: string[] = [];
    (window as unknown as { gh309ClipboardCalls: string[] }).gh309ClipboardCalls = calls;
    Object.defineProperty(navigator, 'clipboard', {
      configurable: true,
      value: { writeText: async (text: string) => void calls.push(text), readText: async () => '' },
    });
  }, seed);

  if (!wallet) {
    await page.goto('/identity');
  } else {
    await page.goto('/create');
    await page.getByLabel('Wallet Name').fill('GH309 custody contract');
    await page.getByLabel('Password', { exact: true }).fill(PASSWORD);
    await page.getByLabel('Confirm Password').fill(PASSWORD);
    await page.getByRole('button', { name: 'Create Wallet' }).click();
    await expect(page.getByRole('heading', { name: 'Save Your Recovery Phrase' })).toBeVisible();
    // Client-side navigation keeps the unlocked in-memory wallet and identity session.
    await page.evaluate(() => {
      window.history.pushState({}, '', '/identity');
      window.dispatchEvent(new PopStateEvent('popstate'));
    });
  }
  if (state === 'ready') {
    await expect(page.getByRole('heading', { name: 'Unencrypted Preview Identity Found' })).toBeVisible();
    await page.getByRole('button', { name: 'Encrypt Into This Wallet' }).click();
  }
  await expect(page.getByRole('heading', { name: heading, exact: true })).toBeVisible({ timeout: 20_000 });
}

for (const viewport of viewports) {
  for (const state of Object.keys(STATES) as CustodyState[]) {
    test(`identity custody ${state} at ${viewport.name}`, async ({ page }) => {
      await page.setViewportSize({ width: viewport.width, height: viewport.height });
      const ledger = { status: 0, violations: [] as string[] };
      const consoleText: string[] = [];
      page.on('console', (message) => consoleText.push(message.text()));
      await guardRpc(page, ledger);
      await reachState(page, state);
      const heading = STATES[state].heading;

      // Disabled controls are inert for pointer, Enter and Space.
      const disabled = page.locator('main button:disabled');
      const disabledCount = await disabled.count();
      for (let index = 0; index < disabledCount; index += 1) {
        const control = disabled.nth(index);
        await control.click({ force: true });
        await control.focus();
        await expect(control).not.toBeFocused();
        await page.keyboard.press('Enter');
        await page.keyboard.press('Space');
      }
      await expect(page.getByRole('heading', { name: heading, exact: true })).toBeVisible();

      // Keyboard order: from an explicit anchor, Tab visits every enabled control in DOM order.
      const tabbableCount = await page.evaluate(() => {
        const tabbable = [...document.querySelectorAll<HTMLElement>('a[href], button, input, select, textarea, [tabindex]')].filter(
          (element) =>
            element.tabIndex >= 0 &&
            !(element as HTMLButtonElement).disabled &&
            element.getClientRects().length > 0 &&
            window.getComputedStyle(element).visibility !== 'hidden'
        );
        if (tabbable.some((element) => element.tabIndex > 0)) throw new Error('positive tabindex breaks DOM order');
        (window as unknown as { gh309Tabbable: HTMLElement[] }).gh309Tabbable = tabbable;
        tabbable[0]?.focus();
        return tabbable.length;
      });
      const focusedIndex = () =>
        page.evaluate(() =>
          (window as unknown as { gh309Tabbable: Element[] }).gh309Tabbable.indexOf(document.activeElement as Element)
        );
      expect(tabbableCount).toBeGreaterThan(0);
      const forward: number[] = [await focusedIndex()];
      for (let step = 1; step < tabbableCount; step += 1) {
        await page.keyboard.press('Tab');
        forward.push(await focusedIndex());
      }
      expect(forward).toEqual([...Array(tabbableCount).keys()]);
      const disabledFocusable = await page.evaluate(() =>
        [...document.querySelectorAll<HTMLButtonElement>('main button:disabled')].some(
          (button) => (window as unknown as { gh309Tabbable: Element[] }).gh309Tabbable.includes(button)
        )
      );
      expect(disabledFocusable).toBe(false);

      const audit = await page.evaluate((secret) => {
        const canvas = document.createElement('canvas');
        canvas.width = 1;
        canvas.height = 1;
        const context = canvas.getContext('2d', { willReadFrequently: true });
        if (!context) throw new Error('canvas unavailable');
        const rgba = (color: string): number[] => {
          context.clearRect(0, 0, 1, 1);
          context.fillStyle = color;
          context.fillRect(0, 0, 1, 1);
          return [...context.getImageData(0, 0, 1, 1).data];
        };
        const luminance = (color: string) => {
          const [r, g, b] = rgba(color).slice(0, 3).map((channel) => {
            const value = channel / 255;
            return value <= 0.03928 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
          });
          return 0.2126 * r + 0.7152 * g + 0.0722 * b;
        };
        const background = (element: Element | null): string => {
          for (let node = element; node; node = node.parentElement) {
            const color = window.getComputedStyle(node).backgroundColor;
            if (rgba(color)[3] > 0) return color;
          }
          return 'rgb(255, 255, 255)';
        };
        const contrast = (element: Element) => {
          const [high, low] = [luminance(window.getComputedStyle(element).color), luminance(background(element))].sort(
            (x, y) => y - x
          );
          return Math.round(((high + 0.05) / (low + 0.05)) * 100) / 100;
        };
        const card = document.querySelector('main .card');
        if (!card) throw new Error('custody card missing');
        const cardBox = card.getBoundingClientRect();
        const inside = (box: DOMRect) =>
          box.left >= cardBox.left - 0.5 && box.right <= cardBox.right + 0.5 && box.top >= cardBox.top - 0.5 && box.bottom <= cardBox.bottom + 0.5;
        const buttons = [...document.querySelectorAll<HTMLButtonElement>('button')].filter((button) => button.getClientRects().length > 0);
        const texts = [...card.querySelectorAll<HTMLElement>('h2, p, span, div.font-mono')].filter(
          (element) => element.getClientRects().length > 0 && !element.classList.contains('sr-only')
        );
        const describedBy = buttons.flatMap((button) => (button.getAttribute('aria-describedby') ?? '').split(/\s+/).filter(Boolean));
        return {
          documentOverflow: document.documentElement.scrollWidth - document.documentElement.clientWidth,
          buttons: buttons.map((button) => {
            const box = button.getBoundingClientRect();
            return {
              name: button.textContent?.trim() ?? '',
              width: Math.round(box.width),
              height: Math.round(box.height),
              disabled: button.disabled,
              contrast: contrast(button),
              insideCard: button.closest('main') ? inside(box) : true,
              labelOverflow: button.scrollWidth - button.clientWidth,
            };
          }),
          minTextContrast: Math.min(...texts.map(contrast)),
          textOverflow: texts.filter((element) => element.scrollWidth - element.clientWidth > 1).map((element) => element.tagName),
          textsInsideCard: texts.every((element) => inside(element.getBoundingClientRect())),
          describedByResolved: describedBy.every((id) => (document.getElementById(id)?.textContent ?? '').trim().length > 0),
          statusCount: document.querySelectorAll('[role="status"]').length,
          secretInDom: document.documentElement.outerHTML.includes(secret),
          clipboardCalls: (window as unknown as { gh309ClipboardCalls?: string[] }).gh309ClipboardCalls?.length ?? 0,
        };
      }, SECRET);

      expect(audit.documentOverflow).toBeLessThanOrEqual(0);
      for (const button of audit.buttons) {
        expect(button.name.length, 'accessible name').toBeGreaterThan(0);
        expect(button.width, button.name).toBeGreaterThanOrEqual(44);
        expect(button.height, button.name).toBeGreaterThanOrEqual(44);
        expect(button.contrast, button.name).toBeGreaterThanOrEqual(4.5);
        expect(button.insideCard, button.name).toBe(true);
        expect(button.labelOverflow, button.name).toBeLessThanOrEqual(0);
      }
      expect(audit.minTextContrast).toBeGreaterThanOrEqual(4.5);
      expect(audit.textOverflow).toEqual([]);
      expect(audit.textsInsideCard).toBe(true);
      expect(audit.describedByResolved).toBe(true);
      if (state !== 'ready') expect(audit.statusCount).toBeGreaterThan(0);
      expect(audit.secretInDom).toBe(false);

      mkdirSync(EVIDENCE_DIR, { recursive: true });
      await page.screenshot({ path: resolve(EVIDENCE_DIR, `${state}-${viewport.name}.png`), fullPage: true });

      if (state === 'ready') {
        const [download] = await Promise.all([
          page.waitForEvent('download'),
          page.getByRole('button', { name: 'Download Identity Backup File' }).click(),
        ]);
        expect(download.suggestedFilename()).toBe('truerepublic-preview-identity.json');
        const exported = JSON.parse(readFileSync((await download.path())!, 'utf8')) as Record<string, unknown>;
        expect(exported).toMatchObject({ format: 'truerepublic-preview-identity', version: 1, canonical: false });
        expect(String(exported.warning)).toContain('not canonical and not production-valid');
      }
      const clipboardCalls = await page.evaluate(
        () => (window as unknown as { gh309ClipboardCalls?: string[] }).gh309ClipboardCalls?.length ?? 0
      );
      expect(clipboardCalls).toBe(0);
      expect(consoleText.some((line) => line.includes(SECRET))).toBe(false);
      expect(ledger.violations).toEqual([]);
      console.log(
        `GH309B2D1 custody ${state} ${viewport.name} audit=${JSON.stringify({ ...audit, buttons: audit.buttons.length })} rpcStatus=${ledger.status}`
      );
    });
  }
}
