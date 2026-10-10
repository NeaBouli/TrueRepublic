import { mkdirSync } from 'node:fs';
import { resolve } from 'node:path';
import process from 'node:process';
import type { Page } from '@playwright/test';
import { previewMockIdentityHash } from '../src/services/previewIdentityHash';
import { expect, test } from './fixtures';

// GH300B3a visual release contract: a ready preview identity opens the real
// VotingPanel from the suggestion list while submission is hard-disabled. The
// panel must never derive a canonical nullifier or query the chain for it, must
// not render the red "Nullifier status unavailable" error, and must keep the
// amber preview warning and the inert vote control. The synthetic secret 'c4'
// x32 lies outside the BN254 field, like most random preview secrets. Allowed
// RPC: the canonical bootstrap `status` (answered with a controlled error) and
// the typed Domain query for the fixture domain. Any Nullifier query, broadcast
// or other request fails the test.

const RPC_ORIGIN = 'http://localhost:26657';
const EVIDENCE_DIR = resolve(process.cwd(), 'node_modules/.cache/gh300b3a-voting-evidence');
const PASSWORD = 'gh300b3a-voting-contract';
const SECRET = 'c4'.repeat(32);
const IDENTITY = {
  secret: SECRET,
  commitment: previewMockIdentityHash(SECRET),
  nullifier: previewMockIdentityHash(`${SECRET}00`),
  createdAt: 1_700_000_000_000,
};
const LEGACY_RAW = `{"state":{"identity":${JSON.stringify(IDENTITY)}},"version":0}`;
const DOMAIN = 'GH300';
const ISSUE = 'GH300Issue';
const SUGGESTION = 'GH300Suggestion';
const DOMAIN_PATH = '/truedemocracy.Query/Domain';
const NULLIFIER_PATH = '/truedemocracy.Query/Nullifier';
const CONTROL_LABEL = 'Anonymous Voting Unavailable';
const PREVIEW_WARNING = 'ZKP preview — submission disabled';
const MIN_TEXT_CONTRAST = 4.5;

const viewports = [
  { name: 'desktop-1440x1000', width: 1440, height: 1000 },
  { name: 'tablet-landscape-1180x820', width: 1180, height: 820 },
  { name: 'tablet-portrait-820x1180', width: 820, height: 1180 },
  { name: 'mobile-390x844', width: 390, height: 844 },
] as const;

interface Ledger {
  status: number;
  domain: number;
  nullifier: number;
  violations: string[];
}

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

const DOMAIN_DATA = lengthDelimitedField1(new TextEncoder().encode(DOMAIN)).toString('hex');

function rpcEnvelope(body: string): { id: number; method: string; params: Record<string, unknown> } | null {
  try {
    const parsed = JSON.parse(body) as Record<string, unknown>;
    if (
      Object.keys(parsed).sort().join(',') !== 'id,jsonrpc,method,params' ||
      parsed.jsonrpc !== '2.0' ||
      !Number.isSafeInteger(parsed.id) ||
      typeof parsed.method !== 'string' ||
      typeof parsed.params !== 'object' ||
      parsed.params === null ||
      Array.isArray(parsed.params)
    ) {
      return null;
    }
    return { id: parsed.id as number, method: parsed.method, params: parsed.params as Record<string, unknown> };
  } catch {
    return null;
  }
}

async function guardRpc(page: Page, member: () => string, ledger: Ledger): Promise<void> {
  // Registered after the network guard, so it takes precedence for the RPC origin.
  await page.route(`${RPC_ORIGIN}/**`, async (route, request) => {
    const body = request.postData() ?? '';
    const rpc = request.method() === 'POST' && new URL(request.url()).pathname === '/' ? rpcEnvelope(body) : null;
    const forbid = async () => {
      ledger.violations.push(`${request.method()} ${request.url()} ${body.slice(0, 200)}`);
      await route.fulfill({ status: 500, body: 'forbidden RPC request in GH300B3a voting contract' });
    };
    if (!rpc || body.includes('broadcast_tx')) return forbid();
    if (rpc.method === 'abci_query' && rpc.params.path === NULLIFIER_PATH) {
      ledger.nullifier += 1;
      return forbid();
    }
    if (rpc.method === 'status' && Object.keys(rpc.params).length === 0) {
      ledger.status += 1;
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ jsonrpc: '2.0', id: rpc.id, error: { code: -32603, message: 'GH300B3a contract: no chain data' } }),
      });
      return;
    }
    const params = rpc.params;
    if (
      rpc.method === 'abci_query' &&
      Object.keys(params).sort().join(',') === 'data,path,prove' &&
      params.path === DOMAIN_PATH &&
      params.data === DOMAIN_DATA &&
      params.prove === false
    ) {
      ledger.domain += 1;
      const suggestion = {
        name: SUGGESTION,
        creator: member(),
        stones: 0,
        ratings: [],
        color: '',
        creation_date: 1_700_000_000,
        external_link: '',
      };
      const domain = {
        name: DOMAIN,
        admin: member(),
        members: [member()],
        treasury: [],
        issues: [{ name: ISSUE, stones: 0, suggestions: [suggestion], creation_date: 1_700_000_000, external_link: '' }],
        permission_reg: [],
        identity_commits: [],
        merkle_root: '',
      };
      const encoded = lengthDelimitedField1(new TextEncoder().encode(JSON.stringify(domain))).toString('base64');
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ jsonrpc: '2.0', id: rpc.id, result: { response: { code: 0, value: encoded } } }),
      });
      return;
    }
    return forbid();
  });
}

async function openVotingPanel(page: Page, ledger: Ledger): Promise<void> {
  let member = '';
  await guardRpc(page, () => member, ledger);
  // Runs after the fixture's storage reset on every navigation.
  await page.addInitScript((raw: string) => window.localStorage.setItem('identity-store', raw), LEGACY_RAW);

  // Ephemeral local test wallet; the only full navigation in this contract.
  await page.goto('/create');
  await page.getByLabel('Wallet Name').fill('GH300B3a voting contract');
  await page.getByLabel('Password', { exact: true }).fill(PASSWORD);
  await page.getByLabel('Confirm Password').fill(PASSWORD);
  await page.getByRole('button', { name: 'Create Wallet' }).click();
  await expect(page.getByRole('heading', { name: 'Save Your Recovery Phrase' })).toBeVisible();
  member = await page.evaluate(() => {
    const stored = JSON.parse(window.localStorage.getItem('wallet-store') ?? '{}');
    return String(stored?.state?.wallets?.[0]?.address ?? '');
  });
  expect(member).toMatch(/^truerepublic1/);

  // Client-side navigation keeps the unlocked in-memory wallet and identity session.
  const navigate = (path: string) =>
    page.evaluate((target) => {
      window.history.pushState({}, '', target);
      window.dispatchEvent(new PopStateEvent('popstate'));
    }, path);
  await navigate('/identity');
  await expect(page.getByRole('heading', { name: 'Unencrypted Preview Identity Found' })).toBeVisible({ timeout: 20_000 });
  await page.getByRole('button', { name: 'Encrypt Into This Wallet' }).click();
  await expect(page.getByTestId('identity-ready-marker')).toBeVisible({ timeout: 20_000 });

  await navigate(`/governance/domain/${DOMAIN}/issue/${ISSUE}`);
  await page.getByRole('button', { name: 'Vote', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'Vote Anonymously' })).toBeVisible();
  await expect(page.getByText(PREVIEW_WARNING)).toBeVisible();
}

for (const viewport of viewports) {
  test(`disabled preview voting shows no nullifier status error at ${viewport.name}`, async ({ page }) => {
    await page.setViewportSize({ width: viewport.width, height: viewport.height });
    const ledger: Ledger = { status: 0, domain: 0, nullifier: 0, violations: [] };
    await openVotingPanel(page, ledger);

    // Give a regressed status effect time to compute, query and render its error.
    await page.waitForTimeout(500);
    await expect(page.getByText(/Nullifier status unavailable/)).toHaveCount(0);
    await expect(page.getByRole('heading', { name: 'Already Voted' })).toHaveCount(0);

    const control = page.getByRole('button', { name: CONTROL_LABEL });
    await expect(control).toBeDisabled();
    const boxBefore = await control.boundingBox();

    // The disabled control is never in the tab sequence and never activates.
    const tabbableHasControl = await page.evaluate((label) => {
      const selector = 'a[href], button, input, select, textarea, [tabindex]';
      return [...document.querySelectorAll<HTMLElement>(selector)].some(
        (element) =>
          element.tabIndex >= 0 &&
          !(element as HTMLButtonElement).disabled &&
          element.textContent?.trim() === label
      );
    }, CONTROL_LABEL);
    expect(tabbableHasControl).toBe(false);
    await control.focus();
    await expect(control).not.toBeFocused();
    await control.click({ force: true });
    await page.keyboard.press('Enter');
    await page.keyboard.press('Space');
    await page.waitForTimeout(250);
    await expect(page.getByRole('heading', { name: 'Generating Anonymous Vote' })).toHaveCount(0);
    await expect(control).toBeDisabled();
    expect(await control.boundingBox()).toEqual(boxBefore);

    const layout = await page.evaluate(
      ({ label, warning }) => {
        const button = [...document.querySelectorAll('button')].find((element) => element.textContent?.trim() === label);
        const heading = [...document.querySelectorAll('strong')].find((element) => element.textContent?.trim() === warning);
        const notice = heading?.closest('.bg-amber-50');
        const card = button?.closest('.card');
        const modal = card?.closest('.overflow-y-auto');
        const text = notice?.querySelector('p');
        if (!button || !notice || !card || !modal || !text) throw new Error('voting panel elements missing');
        const box = (element: Element) => element.getBoundingClientRect();
        const inside = (inner: DOMRect, outer: DOMRect) =>
          inner.left >= outer.left - 0.5 && inner.right <= outer.right + 0.5 &&
          inner.top >= outer.top - 0.5 && inner.bottom <= outer.bottom + 0.5;
        const overlaps = (a: DOMRect, b: DOMRect) =>
          a.left < b.right && b.left < a.right && a.top < b.bottom && b.top < a.bottom;
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
        const b = box(button);
        const n = box(notice);
        const c = box(card);
        return {
          documentOverflow: document.documentElement.scrollWidth - document.documentElement.clientWidth,
          modalOverflowX: modal.scrollWidth - modal.clientWidth,
          button: { width: b.width, height: b.height },
          buttonInsideCard: inside(b, c),
          noticeInsideCard: inside(n, c),
          noticeAboveButton: n.bottom <= b.top,
          noticeButtonOverlap: overlaps(n, b),
          noticeTextOverflow: { x: text.scrollWidth - text.clientWidth, y: text.scrollHeight - text.clientHeight },
          buttonLabelOverflow: { x: button.scrollWidth - button.clientWidth, y: button.scrollHeight - button.clientHeight },
          warningTextContrast: contrast(window.getComputedStyle(text).color, window.getComputedStyle(notice).backgroundColor),
          redStatusBoxes: document.querySelectorAll('.bg-red-50').length,
        };
      },
      { label: CONTROL_LABEL, warning: PREVIEW_WARNING }
    );
    expect(layout.documentOverflow).toBeLessThanOrEqual(0);
    expect(layout.modalOverflowX).toBeLessThanOrEqual(0);
    expect(layout.button.height).toBeGreaterThanOrEqual(44);
    expect(layout.button.width).toBeGreaterThanOrEqual(44);
    expect(layout.buttonInsideCard).toBe(true);
    expect(layout.noticeInsideCard).toBe(true);
    expect(layout.noticeAboveButton).toBe(true);
    expect(layout.noticeButtonOverlap).toBe(false);
    expect(layout.noticeTextOverflow).toEqual({ x: 0, y: 0 });
    expect(layout.buttonLabelOverflow).toEqual({ x: 0, y: 0 });
    expect(layout.warningTextContrast).toBeGreaterThanOrEqual(MIN_TEXT_CONTRAST);
    expect(layout.redStatusBoxes).toBe(0);

    mkdirSync(EVIDENCE_DIR, { recursive: true });
    await page.screenshot({ path: resolve(EVIDENCE_DIR, `${test.info().project.name}-${viewport.name}.png`), fullPage: true });
    console.log(`GH300B3a voting ${viewport.name} layout=${JSON.stringify(layout)} rpc=${JSON.stringify({ ...ledger, violations: ledger.violations.length })}`);

    expect(ledger.nullifier).toBe(0);
    expect(ledger.violations).toEqual([]);
    expect(ledger.domain).toBeGreaterThanOrEqual(1);
  });
}
