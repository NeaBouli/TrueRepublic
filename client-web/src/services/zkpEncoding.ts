import { keccak256, sha256 } from '@cosmjs/crypto';

// Must match github.com/consensys/gnark-crypto/ecc/bn254/fr.
export const BN254_SCALAR_MODULUS =
  21888242871839275222246405745257275088548364400416034343698204186575808495617n;

const MIMC_ROUNDS = 110;
const VOTE_CONTEXT_DOMAIN = new TextEncoder().encode('TrueRepublic/vote/v1');
const textEncoder = new TextEncoder();

function bytesToBigInt(bytes: Uint8Array): bigint {
  let result = 0n;
  for (const byte of bytes) result = (result << 8n) | BigInt(byte);
  return result;
}

function bigintToBytes(value: bigint, size = 32): Uint8Array {
  if (value < 0n || value >= 1n << BigInt(size * 8)) {
    throw new Error(`value does not fit in ${size} bytes`);
  }
  const result = new Uint8Array(size);
  let remaining = value;
  for (let index = size - 1; index >= 0; index--) {
    result[index] = Number(remaining & 0xffn);
    remaining >>= 8n;
  }
  return result;
}

function fieldElement(bytes: Uint8Array, label: string): bigint {
  if (bytes.length > 32) {
    throw new Error(`${label} must be at most 32 bytes`);
  }
  const value = bytesToBigInt(bytes);
  if (value >= BN254_SCALAR_MODULUS) {
    throw new Error(`${label} is not a canonical BN254 field element`);
  }
  return value;
}

function pow5(value: bigint): bigint {
  const square = (value * value) % BN254_SCALAR_MODULUS;
  return (square * square * value) % BN254_SCALAR_MODULUS;
}

const mimcConstants = (() => {
  const constants: bigint[] = [];
  let digest = keccak256(textEncoder.encode('seed'));
  for (let round = 0; round < MIMC_ROUNDS; round++) {
    digest = keccak256(digest);
    constants.push(bytesToBigInt(digest) % BN254_SCALAR_MODULUS);
  }
  return constants;
})();

/** Matches gnark-crypto BN254 MiMC's Miyaguchi-Preneel construction. */
export function mimcBn254(elements: readonly Uint8Array[]): Uint8Array {
  let state = 0n;
  for (const [index, encoded] of elements.entries()) {
    const input = fieldElement(encoded, `MiMC input ${index}`);
    let encrypted = input;
    for (const constant of mimcConstants) {
      encrypted = pow5(
        (encrypted + state + constant) % BN254_SCALAR_MODULUS
      );
    }
    encrypted = (encrypted + state) % BN254_SCALAR_MODULUS;
    state = (encrypted + state + input) % BN254_SCALAR_MODULUS;
  }
  return bigintToBytes(state);
}

export function hashToBn254Field(data: Uint8Array): Uint8Array {
  return bigintToBytes(bytesToBigInt(sha256(data)) % BN254_SCALAR_MODULUS);
}

function uint32BigEndian(value: number): Uint8Array {
  const result = new Uint8Array(4);
  new DataView(result.buffer).setUint32(0, value, false);
  return result;
}

function int64BigEndian(value: number): Uint8Array {
  if (!Number.isSafeInteger(value)) throw new Error('rating must be an integer');
  const result = new Uint8Array(8);
  new DataView(result.buffer).setBigInt64(0, BigInt(value), false);
  return result;
}

function concatBytes(parts: readonly Uint8Array[]): Uint8Array {
  const length = parts.reduce((total, part) => total + part.length, 0);
  const result = new Uint8Array(length);
  let offset = 0;
  for (const part of parts) {
    result.set(part, offset);
    offset += part.length;
  }
  return result;
}

export function encodeVoteContext(
  chainId: string,
  domainName: string,
  issueName: string,
  suggestionName: string,
  rating?: number
): Uint8Array {
  const parts: Uint8Array[] = [VOTE_CONTEXT_DOMAIN];
  for (const value of [chainId, domainName, issueName, suggestionName]) {
    const encoded = textEncoder.encode(value);
    parts.push(uint32BigEndian(encoded.length), encoded);
  }
  if (rating !== undefined) parts.push(int64BigEndian(rating));
  return concatBytes(parts);
}

export function computeVoteNullifierScope(
  chainId: string,
  domainName: string,
  issueName: string,
  suggestionName: string
): Uint8Array {
  return hashToBn254Field(
    encodeVoteContext(chainId, domainName, issueName, suggestionName)
  );
}

export function computeVoteSignal(
  chainId: string,
  domainName: string,
  issueName: string,
  suggestionName: string,
  rating: number
): Uint8Array {
  return hashToBn254Field(
    encodeVoteContext(chainId, domainName, issueName, suggestionName, rating)
  );
}

export function hexToBytes(hex: string): Uint8Array {
  if (!/^(?:[0-9a-f]{2})+$/u.test(hex)) {
    throw new Error('value must be non-empty canonical lowercase hex');
  }
  return Uint8Array.from(hex.match(/.{2}/gu) ?? [], (byte) =>
    Number.parseInt(byte, 16)
  );
}

export function bytesToHex(bytes: Uint8Array): string {
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('');
}
