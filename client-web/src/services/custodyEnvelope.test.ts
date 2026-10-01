import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import process from 'node:process';
import { afterEach, describe, expect, it, vi } from 'vitest';
import {
  CUSTODY_ENVELOPE_FAILURE,
  CustodyEnvelopeError,
  openEnvelope,
  sealEnvelope,
} from './custodyEnvelope';

// Frozen fixtures. FIXTURE_V2 was produced by the pre-extraction
// WalletService.encrypt (main 1283a445) with getRandomValues stubbed to
// salt 0x01..0x10 and IV 0xa0..0xab; FIXTURE_LEGACY is an unprefixed
// 100,000-iteration payload (salt 0x30..0x3f, IV 0x50..0x5b). Synthetic only.
const PASSWORD = 'synthetic-fixture-password';
const PLAINTEXT_V2 = 'synthetic fixture plaintext GH309B2B';
const FIXTURE_V2 =
  'v2:AQIDBAUGBwgJCgsMDQ4PEKChoqOkpaanqKmqq5yJEyOoHtqMLb7JPZBTOKTnrHOQYhZxKJhfcftN2YB9feKUB28xy4kH6EqN5y3Wr3gYOdU=';
const PLAINTEXT_LEGACY = 'synthetic legacy fixture GH309B2B';
const FIXTURE_LEGACY =
  'MDEyMzQ1Njc4OTo7PD0+P1BRUlNUVVZXWFlaW/SiN6jy4N0EZ41Rh3sX8vAqD8Ear78jvuWOChQwwokTlBNlmvxexZNsWftrEyEf/E0=';
const AAD = 'truerepublic/identity/v1|synthetic';

function stubRandom(...chunks: Uint8Array[]) {
  const queue = [...chunks];
  return vi.spyOn(crypto, 'getRandomValues').mockImplementation(((array: Uint8Array) => {
    const next = queue.shift();
    if (!next) throw new Error('unexpected getRandomValues call');
    array.set(next);
    return array;
  }) as typeof crypto.getRandomValues);
}

function decodePayload(envelope: string): Uint8Array {
  return Uint8Array.from(atob(envelope.replace(/^v2:/, '')), (c) => c.charCodeAt(0));
}

function encodePayload(bytes: Uint8Array, prefix = 'v2:'): string {
  return `${prefix}${btoa(String.fromCharCode(...bytes))}`;
}

async function expectFailure(promise: Promise<unknown>): Promise<void> {
  const error = await promise.then(
    () => null,
    (caught: unknown) => caught
  );
  expect(error).toBeInstanceOf(CustodyEnvelopeError);
  expect((error as Error).message).toBe(CUSTODY_ENVELOPE_FAILURE);
}

describe('custody envelope', { timeout: 30_000 }, () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('seals byte-identically to the historical wallet envelope without AAD', async () => {
    stubRandom(
      Uint8Array.from({ length: 16 }, (_, i) => i + 1),
      Uint8Array.from({ length: 12 }, (_, i) => 0xa0 + i)
    );
    await expect(sealEnvelope(PLAINTEXT_V2, PASSWORD)).resolves.toBe(FIXTURE_V2);
  });

  it('opens the frozen v2 and, only when allowed, the legacy 100k payload', async () => {
    await expect(openEnvelope(FIXTURE_V2, PASSWORD)).resolves.toEqual({
      plaintext: PLAINTEXT_V2,
      needsUpgrade: false,
    });
    await expect(openEnvelope(FIXTURE_LEGACY, PASSWORD, { acceptLegacy: true })).resolves.toEqual({
      plaintext: PLAINTEXT_LEGACY,
      needsUpgrade: true,
    });
    await expectFailure(openEnvelope(FIXTURE_LEGACY, PASSWORD));
  });

  it('uses fresh salt, IV and ciphertext for every seal', async () => {
    const first = decodePayload(await sealEnvelope(PLAINTEXT_V2, PASSWORD, { aad: AAD }));
    const second = decodePayload(await sealEnvelope(PLAINTEXT_V2, PASSWORD, { aad: AAD }));
    expect(first.slice(0, 16)).not.toEqual(second.slice(0, 16));
    expect(first.slice(16, 28)).not.toEqual(second.slice(16, 28));
    expect(first.slice(28)).not.toEqual(second.slice(28));
  });

  it('binds AAD: identical AAD opens, any other, missing or extra AAD fails', async () => {
    const bound = await sealEnvelope(PLAINTEXT_V2, PASSWORD, { aad: AAD });
    await expect(openEnvelope(bound, PASSWORD, { aad: AAD })).resolves.toMatchObject({
      plaintext: PLAINTEXT_V2,
    });
    await expectFailure(openEnvelope(bound, PASSWORD, { aad: `${AAD}x` }));
    await expectFailure(openEnvelope(bound, PASSWORD));
    await expectFailure(openEnvelope(FIXTURE_V2, PASSWORD, { aad: AAD }));
  });

  it('fails closed on a wrong password and on modified ciphertext or tag', async () => {
    await expectFailure(openEnvelope(FIXTURE_V2, 'wrong-fixture-password'));
    const bytes = decodePayload(FIXTURE_V2);
    const flippedCiphertext = bytes.slice();
    flippedCiphertext[30] ^= 0x01;
    await expectFailure(openEnvelope(encodePayload(flippedCiphertext), PASSWORD));
    const flippedTag = bytes.slice();
    flippedTag[flippedTag.length - 1] ^= 0x80;
    await expectFailure(openEnvelope(encodePayload(flippedTag), PASSWORD));
    const flippedSalt = bytes.slice();
    flippedSalt[0] ^= 0x01;
    await expectFailure(openEnvelope(encodePayload(flippedSalt), PASSWORD));
  });

  it('rejects malformed or oversized input before any Base64 decode or key derivation', async () => {
    const importKey = vi.spyOn(crypto.subtle, 'importKey');
    const atobSpy = vi.spyOn(globalThis, 'atob');
    const cases: Array<[string, Parameters<typeof openEnvelope>[2]]> = [
      [`v2:${'A'.repeat(8_192)}`, {}],
      [`v2:${'A'.repeat(100)}`, { maxEnvelopeChars: 64 }],
      ['v3:AAAA', {}],
      ['v2:not base64!', {}],
      ['v2:AAA', {}],
      ['v2:', {}],
      [FIXTURE_LEGACY, {}],
    ];
    for (const [envelope, options] of cases) {
      await expectFailure(openEnvelope(envelope, PASSWORD, options));
    }
    await expectFailure(openEnvelope(FIXTURE_V2, PASSWORD, { aad: 'x'.repeat(513) }));
    await expectFailure(openEnvelope(FIXTURE_V2, PASSWORD, { aad: '' }));
    await expectFailure(openEnvelope(42 as unknown as string, PASSWORD));
    expect(atobSpy).not.toHaveBeenCalled();
    expect(importKey).not.toHaveBeenCalled();
  });

  it('rejects a well-formed payload shorter than salt, IV and tag before key derivation', async () => {
    const importKey = vi.spyOn(crypto.subtle, 'importKey');
    await expectFailure(openEnvelope(encodePayload(new Uint8Array(43)), PASSWORD));
    expect(importKey).not.toHaveBeenCalled();
  });

  it('never surfaces raw WebCrypto errors', async () => {
    vi.spyOn(crypto.subtle, 'deriveKey').mockRejectedValue(new DOMException('raw failure', 'OperationError'));
    await expectFailure(openEnvelope(FIXTURE_V2, PASSWORD));
    await expectFailure(sealEnvelope(PLAINTEXT_V2, PASSWORD));
  });

  it('leaves WalletService without a duplicate crypto implementation', () => {
    const source = readFileSync(resolve(process.cwd(), 'src/services/wallet.ts'), 'utf8');
    for (const primitive of ['crypto.subtle', 'getRandomValues', 'PBKDF2', 'AES-GCM', 'atob(', 'btoa(', '600_000', '100_000']) {
      expect(source).not.toContain(primitive);
    }
    expect(source).toContain("from './custodyEnvelope'");
  });
});
