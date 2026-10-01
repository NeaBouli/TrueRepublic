/**
 * Shared client custody envelope (PBKDF2-SHA256 -> AES-GCM-256, WebCrypto only).
 *
 * Wire format, unchanged from the original wallet implementation:
 *   current: `v2:` + Base64(salt(16) | iv(12) | ciphertext+tag), 600,000 iterations
 *   legacy:  Base64(salt(16) | iv(12) | ciphertext+tag), 100,000 iterations, no prefix
 *
 * Without `aad` the bytes and decrypt behaviour are identical to the historical
 * wallet envelope. With `aad`, the same UTF-8 string must be supplied as
 * AES-GCM `additionalData` on both seal and open. Every failure surfaces as one
 * bounded CustodyEnvelopeError; raw WebCrypto errors are never propagated.
 */

export const CUSTODY_ENVELOPE_VERSION = 'v2';
export const CURRENT_PBKDF2_ITERATIONS = 600_000;
export const LEGACY_PBKDF2_ITERATIONS = 100_000;

const SALT_BYTES = 16;
const IV_BYTES = 12;
const TAG_BYTES = 16;
// salt (16) + iv (12) + minimum AES-GCM tag (16)
const MIN_PAYLOAD_BYTES = SALT_BYTES + IV_BYTES + TAG_BYTES;
const DEFAULT_MAX_ENVELOPE_CHARS = 8_192;
const MAX_AAD_CHARS = 512;
const BASE64_PATTERN = /^[A-Za-z0-9+/]+={0,2}$/;
const CURRENT_PREFIX = `${CUSTODY_ENVELOPE_VERSION}:`;

export const CUSTODY_ENVELOPE_FAILURE = 'Incorrect password or corrupted data';

export class CustodyEnvelopeError extends Error {
  constructor() {
    super(CUSTODY_ENVELOPE_FAILURE);
    this.name = 'CustodyEnvelopeError';
  }
}

export interface CustodyEnvelopeOptions {
  /** Associated data bound to the ciphertext; omit for the historical wallet format. */
  aad?: string;
  /** Narrower bound on the serialized envelope (<= 8192), checked before Base64 decode or KDF work. */
  maxEnvelopeChars?: number;
  /** Accept the unprefixed 100k payload (wallet storage only). */
  acceptLegacy?: boolean;
}

export interface OpenedEnvelope {
  plaintext: string;
  /** True when the payload used the legacy iteration count and should be re-sealed. */
  needsUpgrade: boolean;
}

// Password policy (length, strength) stays with each caller; the historical
// wallet path accepted any string here and must keep doing so.
function assertPassword(password: string): void {
  if (typeof password !== 'string') throw new CustodyEnvelopeError();
}

function aadBytes(options: CustodyEnvelopeOptions): Uint8Array | undefined {
  if (options.aad === undefined) return undefined;
  if (typeof options.aad !== 'string' || options.aad.length === 0 || options.aad.length > MAX_AAD_CHARS) {
    throw new CustodyEnvelopeError();
  }
  return new TextEncoder().encode(options.aad);
}

function maxEnvelopeChars(options: CustodyEnvelopeOptions): number {
  const max = options.maxEnvelopeChars ?? DEFAULT_MAX_ENVELOPE_CHARS;
  // Callers may only narrow the hard default, never widen it.
  if (!Number.isSafeInteger(max) || max <= CURRENT_PREFIX.length || max > DEFAULT_MAX_ENVELOPE_CHARS) {
    throw new CustodyEnvelopeError();
  }
  return max;
}

function gcmParams(iv: Uint8Array, aad: Uint8Array | undefined): AesGcmParams {
  const params: AesGcmParams = { name: 'AES-GCM', iv: iv as ArrayBufferView<ArrayBuffer> };
  if (aad) params.additionalData = aad as ArrayBufferView<ArrayBuffer>;
  return params;
}

async function deriveKey(password: string, salt: Uint8Array, iterations: number): Promise<CryptoKey> {
  const passwordBytes = new TextEncoder().encode(password);
  const keyMaterial = await crypto.subtle.importKey(
    'raw',
    passwordBytes as ArrayBufferView<ArrayBuffer>,
    'PBKDF2',
    false,
    ['deriveKey']
  );

  return crypto.subtle.deriveKey(
    {
      name: 'PBKDF2',
      salt: salt as ArrayBufferView<ArrayBuffer>,
      iterations,
      hash: 'SHA-256',
    },
    keyMaterial,
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt', 'decrypt']
  );
}

/**
 * Seal plaintext with a fresh random salt and IV at the current iteration count.
 */
export async function sealEnvelope(
  plaintext: string,
  password: string,
  options: CustodyEnvelopeOptions = {}
): Promise<string> {
  assertPassword(password);
  if (typeof plaintext !== 'string') throw new CustodyEnvelopeError();
  const aad = aadBytes(options);
  const max = maxEnvelopeChars(options);

  let encrypted: ArrayBuffer;
  const salt = crypto.getRandomValues(new Uint8Array(SALT_BYTES));
  const iv = crypto.getRandomValues(new Uint8Array(IV_BYTES));
  try {
    const key = await deriveKey(password, salt, CURRENT_PBKDF2_ITERATIONS);
    encrypted = await crypto.subtle.encrypt(
      gcmParams(iv, aad),
      key,
      new TextEncoder().encode(plaintext) as ArrayBufferView<ArrayBuffer>
    );
  } catch {
    throw new CustodyEnvelopeError();
  }

  // Combine salt + iv + ciphertext
  const combined = new Uint8Array(salt.length + iv.length + encrypted.byteLength);
  combined.set(salt, 0);
  combined.set(iv, salt.length);
  combined.set(new Uint8Array(encrypted), salt.length + iv.length);

  const envelope = `${CURRENT_PREFIX}${btoa(String.fromCharCode(...combined))}`;
  if (envelope.length > max) throw new CustodyEnvelopeError();
  return envelope;
}

/**
 * Validate structure and bounds before any Base64 decode or key derivation,
 * then authenticate and decrypt. Wrong password, wrong AAD, tampering and
 * malformed input are indistinguishable to the caller.
 */
export async function openEnvelope(
  envelope: string,
  password: string,
  options: CustodyEnvelopeOptions = {}
): Promise<OpenedEnvelope> {
  assertPassword(password);
  const aad = aadBytes(options);
  const max = maxEnvelopeChars(options);
  if (typeof envelope !== 'string' || envelope.length > max) throw new CustodyEnvelopeError();

  const isCurrent = envelope.startsWith(CURRENT_PREFIX);
  if (!isCurrent && !options.acceptLegacy) throw new CustodyEnvelopeError();
  const payload = isCurrent ? envelope.slice(CURRENT_PREFIX.length) : envelope;
  if (payload.length % 4 !== 0 || !BASE64_PATTERN.test(payload)) throw new CustodyEnvelopeError();

  let combined: Uint8Array;
  try {
    combined = Uint8Array.from(atob(payload), (c) => c.charCodeAt(0));
  } catch {
    throw new CustodyEnvelopeError();
  }
  if (combined.length < MIN_PAYLOAD_BYTES) throw new CustodyEnvelopeError();

  const salt = combined.slice(0, SALT_BYTES);
  const iv = combined.slice(SALT_BYTES, SALT_BYTES + IV_BYTES);
  const ciphertext = combined.slice(SALT_BYTES + IV_BYTES);

  try {
    const key = await deriveKey(
      password,
      salt,
      isCurrent ? CURRENT_PBKDF2_ITERATIONS : LEGACY_PBKDF2_ITERATIONS
    );
    const decrypted = await crypto.subtle.decrypt(
      gcmParams(iv, aad),
      key,
      ciphertext as ArrayBufferView<ArrayBuffer>
    );
    return {
      plaintext: new TextDecoder().decode(decrypted),
      needsUpgrade: !isCurrent,
    };
  } catch {
    throw new CustodyEnvelopeError();
  }
}
