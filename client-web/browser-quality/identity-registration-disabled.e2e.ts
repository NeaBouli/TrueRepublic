import { mkdirSync } from 'node:fs';
import { resolve } from 'node:path';
import process from 'node:process';
import type { Page, Request, Route } from '@playwright/test';
import { expect, test } from './fixtures';

// GH-309 visual release contract: a member with a preview identity reaches the
// former identity-registration step, which must be unavailable, keyboard-safe,
// at least 44x44 CSS px and never reach the registration service. The only
// chain call answered with data is the exact typed Domain abci_query. Creating
// the wallet bootstraps a balance refresh whose canonical Comet `status` call
// (empty params object, nothing else) is answered with a controlled JSON-RPC
// error without data; the refresh then stops, so no bank query is allowed.
// The policy is evaluated in order: broadcast_tx* always fails, then canonical
// status, then the exact Domain query; any other method, path, shape or
// payload fails the test, and the observed RPC sequence must match exactly.

const DOMAIN = 'GH309';
const RPC_ORIGIN = 'http://localhost:26657';
const DOMAIN_QUERY_PATH = '/truedemocracy.Query/Domain';
// Playwright runs from client-web/; evidence stays in the ignored local cache, never in Git.
const EVIDENCE_DIR = resolve(process.cwd(), 'node_modules/.cache/gh309b1c1-a-final-evidence');
const DISABLED_LABEL = 'Registration Disabled in Preview';
const MIN_TEXT_CONTRAST = 4.5;

type RpcCategory = 'status' | 'abci_query Domain';
// Observed and required sequence: one bootstrap status, then the Domain query on
// entering /onboard and again after the identity is created.
const EXPECTED_RPC_SEQUENCE: RpcCategory[] = ['status', 'abci_query Domain', 'abci_query Domain'];

interface RpcLedger {
  allowed: RpcCategory[];
  broadcasts: string[];
  violations: string[];
}

const viewports = [
  { name: 'desktop-1440x1000', width: 1440, height: 1000 },
  { name: 'tablet-landscape-1180x820', width: 1180, height: 820 },
  { name: 'tablet-portrait-820x1180', width: 820, height: 1180 },
  { name: 'mobile-390x844', width: 390, height: 844 },
] as const;

// Minimal protobuf helpers mirroring moduleQuery.ts: the request carries the
// domain name as field 1 (string); the response result is field 1 (bytes)
// holding the module's JSON. Both directions are checked by the production
// codecs on every run: the page's encodeRequest output must equal this body
// byte for byte (otherwise it is a violation), and only the page's
// decodeResult/expectChainDomain can turn the response into the target state.
function varint(value: number): number[] {
  const out: number[] = [];
  let rest = value;
  while (rest >= 0x80) {
    out.push((rest & 0x7f) | 0x80);
    rest >>>= 7;
  }
  out.push(rest);
  return out;
}

function lengthDelimitedField1(bytes: Uint8Array): Buffer {
  return Buffer.from([0x0a, ...varint(bytes.length), ...bytes]);
}

function expectedDomainQueryBody(): string {
  return JSON.stringify({
    jsonrpc: '2.0',
    id: 1,
    method: 'abci_query',
    params: {
      path: DOMAIN_QUERY_PATH,
      data: lengthDelimitedField1(new TextEncoder().encode(DOMAIN)).toString('hex'),
      prove: false,
    },
  });
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

// Canonical JSON-RPC 2.0 envelope: exactly jsonrpc, id, method and params.
function hasCanonicalEnvelope(parsed: Record<string, unknown>): boolean {
  const keys = Object.keys(parsed).sort();
  return (
    keys.join(',') === 'id,jsonrpc,method,params' &&
    parsed.jsonrpc === '2.0' &&
    Number.isSafeInteger(parsed.id) &&
    typeof parsed.method === 'string'
  );
}

// Ordered RPC policy. Returns the allowed category, or records a broadcast or
// violation and returns null.
function classifyRpc(
  httpMethod: string,
  rawUrl: string,
  body: string,
  ledger: RpcLedger
): { id: unknown; category: RpcCategory } | null {
  const url = new URL(rawUrl);
  let parsed: unknown = null;
  try {
    parsed = JSON.parse(body);
  } catch {
    parsed = null;
  }
  const method = isPlainObject(parsed) && typeof parsed.method === 'string' ? parsed.method : '';
  // Also catches batched or otherwise wrapped broadcasts.
  if (method.startsWith('broadcast_tx') || url.pathname.includes('broadcast_tx') || body.includes('broadcast_tx')) {
    ledger.broadcasts.push(`${httpMethod} ${url.href} ${method}`);
    return null;
  }
  const reject = () => {
    ledger.violations.push(`${httpMethod} ${url.href} ${body.slice(0, 200)}`);
    return null;
  };
  if (httpMethod !== 'POST' || url.pathname !== '/' || !isPlainObject(parsed) || !hasCanonicalEnvelope(parsed)) {
    return reject();
  }
  if (method === 'status' && isPlainObject(parsed.params) && Object.keys(parsed.params).length === 0) {
    return { id: parsed.id, category: 'status' };
  }
  if (body === expectedDomainQueryBody()) {
    return { id: parsed.id, category: 'abci_query Domain' };
  }
  return reject();
}

async function routeRpc(page: Page, member: () => string, ledger: RpcLedger): Promise<void> {
  // Registered after the network guard, so it takes precedence for the RPC origin.
  await page.route(`${RPC_ORIGIN}/**`, async (route: Route, request: Request) => {
    const body = request.postData() ?? '';
    const allowed = classifyRpc(request.method(), request.url(), body, ledger);
    if (!allowed) {
      await route.fulfill({ status: 500, body: 'forbidden RPC request in GH-309 visual contract' });
      return;
    }
    ledger.allowed.push(allowed.category);
    if (allowed.category !== 'abci_query Domain') {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          jsonrpc: '2.0',
          id: allowed.id,
          error: { code: -32603, message: 'GH-309 visual contract: no chain data' },
        }),
      });
      return;
    }
    const domain = {
      name: DOMAIN,
      admin: member(),
      members: [member()],
      treasury: [],
      issues: [],
      permission_reg: [],
      identity_commits: [],
      merkle_root: '',
    };
    const value = lengthDelimitedField1(new TextEncoder().encode(JSON.stringify(domain))).toString('base64');
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ jsonrpc: '2.0', id: allowed.id, result: { response: { code: 0, value } } }),
    });
  });
}

for (const viewport of viewports) {
  test(`identity registration is unavailable at ${viewport.name}`, async ({ page }) => {
    await page.setViewportSize({ width: viewport.width, height: viewport.height });
    const ledger: RpcLedger = { allowed: [], broadcasts: [], violations: [] };
    let memberAddress = '';
    await routeRpc(page, () => memberAddress, ledger);

    // Ephemeral local test wallet; the only full navigation in this contract.
    await page.goto('/create');
    await page.getByLabel('Wallet Name').fill('GH309 visual contract');
    await page.getByLabel('Password', { exact: true }).fill('gh309-visual-contract');
    await page.getByLabel('Confirm Password').fill('gh309-visual-contract');
    await page.getByRole('button', { name: 'Create Wallet' }).click();
    await expect(page.getByRole('heading', { name: 'Save Your Recovery Phrase' })).toBeVisible();

    memberAddress = await page.evaluate(() => {
      const stored = JSON.parse(window.localStorage.getItem('wallet-store') ?? '{}');
      return String(stored?.state?.wallets?.[0]?.address ?? '');
    });
    expect(memberAddress).toMatch(/^truerepublic1/);

    // Client-side navigation keeps the unlocked in-memory wallet.
    await page.evaluate((path) => {
      window.history.pushState({}, '', path);
      window.dispatchEvent(new PopStateEvent('popstate'));
    }, `/onboard/${DOMAIN}`);
    await page.getByRole('button', { name: 'Create Anonymous Identity' }).click();

    await expect(page.getByRole('heading', { name: 'Identity Registration Unavailable' })).toBeVisible();
    const control = page.getByRole('button', { name: DISABLED_LABEL });
    await expect(control).toBeDisabled();
    await expect(page.getByTestId('identity-registration-disabled-notice')).toContainText('No transaction is sent');
    const urlBefore = page.url();
    const boxBefore = await control.boundingBox();

    // Tab order: every tabbable element is visited in DOM order and the disabled control never is.
    // The traversal starts from an explicit focus anchor on the first tabbable element: browsers
    // differ in where sequential navigation starts after the clicked "Create Anonymous Identity"
    // button is removed (Firefox keeps that position), so a blur() start is not browser-neutral.
    const tabbableCount = await page.evaluate((label) => {
      const selector = 'a[href], button, input, select, textarea, [tabindex]';
      const tabbable = [...document.querySelectorAll<HTMLElement>(selector)].filter(
        (element) =>
          element.tabIndex >= 0 &&
          !(element as HTMLButtonElement).disabled &&
          element.getClientRects().length > 0 &&
          window.getComputedStyle(element).visibility !== 'hidden'
      );
      if (tabbable.some((element) => element.textContent?.trim() === label)) throw new Error('disabled control is tabbable');
      if (tabbable.some((element) => element.tabIndex > 0)) throw new Error('positive tabindex breaks DOM order');
      (window as unknown as { gh309Tabbable: HTMLElement[] }).gh309Tabbable = tabbable;
      tabbable[0]?.focus();
      return tabbable.length;
    }, DISABLED_LABEL);
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
    const backward: number[] = [];
    for (let step = 1; step < tabbableCount; step += 1) {
      await page.keyboard.press('Shift+Tab');
      backward.push(await focusedIndex());
    }
    expect(backward).toEqual([...Array(tabbableCount - 1).keys()].reverse());

    // No activation by focus, pointer, Enter or Space.
    await control.focus();
    await expect(control).not.toBeFocused();
    await control.click({ force: true });
    await page.keyboard.press('Enter');
    await page.keyboard.press('Space');
    await page.waitForTimeout(250);
    await expect(page.getByRole('heading', { name: 'Identity Registration Unavailable' })).toBeVisible();
    await expect(control).toBeDisabled();
    expect(page.url()).toBe(urlBefore);
    expect(await control.boundingBox()).toEqual(boxBefore);

    const layout = await page.evaluate((label) => {
      const button = [...document.querySelectorAll('button')].find(
        (element) => element.textContent?.trim() === label
      );
      const notice = document.querySelector('[data-testid="identity-registration-disabled-notice"]');
      const card = notice?.closest('.card');
      const text = notice?.querySelector('p');
      if (!button || !notice || !card || !text) throw new Error('registration-disabled elements missing');
      const box = (element: Element) => element.getBoundingClientRect();
      const b = box(button);
      const n = box(notice);
      const c = box(card);
      const t = box(text);
      const inside = (inner: DOMRect, outer: DOMRect) =>
        inner.left >= outer.left - 0.5 && inner.right <= outer.right + 0.5 &&
        inner.top >= outer.top - 0.5 && inner.bottom <= outer.bottom + 0.5;
      // Resolve any CSS color syntax to sRGB through a canvas, then apply WCAG 2.x contrast.
      const canvas = document.createElement('canvas');
      canvas.width = 1;
      canvas.height = 1;
      const context = canvas.getContext('2d', { willReadFrequently: true });
      if (!context) throw new Error('canvas unavailable');
      const rgb = (color: string): number[] => {
        context.clearRect(0, 0, 1, 1);
        context.fillStyle = color;
        context.fillRect(0, 0, 1, 1);
        return [...context.getImageData(0, 0, 1, 1).data.slice(0, 3)];
      };
      const luminance = (color: string) => {
        const [r, g, bl] = rgb(color).map((channel) => {
          const value = channel / 255;
          return value <= 0.03928 ? value / 12.92 : ((value + 0.055) / 1.055) ** 2.4;
        });
        return 0.2126 * r + 0.7152 * g + 0.0722 * bl;
      };
      const contrast = (foreground: string, background: string) => {
        const [high, low] = [luminance(foreground), luminance(background)].sort((x, y) => y - x);
        return Math.round(((high + 0.05) / (low + 0.05)) * 100) / 100;
      };
      const buttonStyle = window.getComputedStyle(button);
      const noticeStyle = window.getComputedStyle(notice);
      const textStyle = window.getComputedStyle(text);
      return {
        documentOverflow: document.documentElement.scrollWidth - document.documentElement.clientWidth,
        button: { width: b.width, height: b.height },
        buttonInsideCard: inside(b, c),
        noticeInsideCard: inside(n, c),
        noticeTextInsideNotice: inside(t, n),
        noticeTextOverflow: { x: text.scrollWidth - text.clientWidth, y: text.scrollHeight - text.clientHeight },
        buttonLabelOverflow: { x: button.scrollWidth - button.clientWidth, y: button.scrollHeight - button.clientHeight },
        buttonColors: { color: buttonStyle.color, background: buttonStyle.backgroundColor },
        buttonCursor: buttonStyle.cursor,
        disabledLabelContrast: contrast(buttonStyle.color, buttonStyle.backgroundColor),
        warningTextContrast: contrast(textStyle.color, noticeStyle.backgroundColor),
      };
    }, DISABLED_LABEL);
    expect(layout.documentOverflow).toBeLessThanOrEqual(0);
    expect(layout.button.width).toBeGreaterThanOrEqual(44);
    expect(layout.button.height).toBeGreaterThanOrEqual(44);
    expect(layout.buttonInsideCard).toBe(true);
    expect(layout.noticeInsideCard).toBe(true);
    expect(layout.noticeTextInsideNotice).toBe(true);
    expect(layout.noticeTextOverflow).toEqual({ x: 0, y: 0 });
    expect(layout.buttonLabelOverflow).toEqual({ x: 0, y: 0 });
    expect(layout.buttonCursor).toBe('not-allowed');
    expect(layout.disabledLabelContrast).toBeGreaterThanOrEqual(MIN_TEXT_CONTRAST);

    mkdirSync(EVIDENCE_DIR, { recursive: true });
    await page.screenshot({ path: resolve(EVIDENCE_DIR, `${viewport.name}.png`), fullPage: true });
    const rpc = {
      sequence: ledger.allowed,
      status: ledger.allowed.filter((category) => category === 'status').length,
      domain: ledger.allowed.filter((category) => category === 'abci_query Domain').length,
      broadcasts: ledger.broadcasts.length,
      violations: ledger.violations.length,
    };
    console.log(`GH309B1C1a ${viewport.name} layout=${JSON.stringify(layout)} rpc=${JSON.stringify(rpc)}`);

    expect(ledger.broadcasts).toEqual([]);
    expect(ledger.violations).toEqual([]);
    expect(ledger.allowed).toEqual(EXPECTED_RPC_SEQUENCE);
    // GH-335: the typed Domain query leaves the page and its decoded state drives the target step.
    expect(rpc.domain).toBeGreaterThanOrEqual(1);
  });
}

// Pure policy contract (no page): every widened or non-canonical bootstrap
// variant and every bank query must be rejected, broadcasts must be recorded.
test.describe('GH-309 RPC policy', () => {
  const rpcUrl = `${RPC_ORIGIN}/`;
  const envelope = (method: string, params: unknown, extra: Record<string, unknown> = {}) =>
    JSON.stringify({ jsonrpc: '2.0', id: 7, method, params, ...extra });
  const classify = (body: string, httpMethod = 'POST', url = rpcUrl) => {
    const ledger: RpcLedger = { allowed: [], broadcasts: [], violations: [] };
    const result = classifyRpc(httpMethod, url, body, ledger);
    return { category: result?.category ?? null, ledger };
  };

  test('allows only the canonical status and the exact Domain query', () => {
    expect(classify(envelope('status', {})).category).toBe('status');
    expect(classify(expectedDomainQueryBody()).category).toBe('abci_query Domain');
  });

  const rejected: Array<[string, string]> = [
    ['status with null params', envelope('status', null)],
    ['status with array params', envelope('status', [])],
    ['status with extra params', envelope('status', { height: '1' })],
    ['status without params', JSON.stringify({ jsonrpc: '2.0', id: 7, method: 'status' })],
    ['status with extra top-level field', envelope('status', {}, { extra: true })],
    ['status with string id', JSON.stringify({ jsonrpc: '2.0', id: '7', method: 'status', params: {} })],
    ['status with wrong jsonrpc', JSON.stringify({ jsonrpc: '1.0', id: 7, method: 'status', params: {} })],
    ['batched status', JSON.stringify([{ jsonrpc: '2.0', id: 7, method: 'status', params: {} }])],
    ['bank AllBalances', envelope('abci_query', { path: '/cosmos.bank.v1beta1.Query/AllBalances', data: '', prove: false })],
    ['bank Balance', envelope('abci_query', { path: '/cosmos.bank.v1beta1.Query/Balance', data: '', prove: false })],
    ['Domain with other id', expectedDomainQueryBody().replace('"id":1', '"id":2')],
    ['Domain with prove true', expectedDomainQueryBody().replace('"prove":false', '"prove":true')],
    ['other method', envelope('abci_info', {})],
  ];
  for (const [name, body] of rejected) {
    test(`rejects ${name}`, () => {
      const { category, ledger } = classify(body);
      expect(category).toBeNull();
      expect(ledger.violations).toHaveLength(1);
      expect(ledger.broadcasts).toEqual([]);
    });
  }

  test('rejects canonical status over GET or another path', () => {
    expect(classify(envelope('status', {}), 'GET').category).toBeNull();
    expect(classify(envelope('status', {}), 'POST', `${RPC_ORIGIN}/status`).category).toBeNull();
  });

  for (const method of ['broadcast_tx_sync', 'broadcast_tx_async', 'broadcast_tx_commit']) {
    test(`records ${method} as a broadcast`, () => {
      const { category, ledger } = classify(envelope(method, { tx: 'AA==' }));
      expect(category).toBeNull();
      expect(ledger.broadcasts).toHaveLength(1);
    });
  }
  test('records a batched broadcast', () => {
    const { ledger } = classify(JSON.stringify([{ jsonrpc: '2.0', id: 7, method: 'broadcast_tx_sync', params: {} }]));
    expect(ledger.broadcasts).toHaveLength(1);
  });
});
