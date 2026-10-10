/**
 * Historical PREVIEW/MOCK identity hash (issue #309).
 *
 * 32-bit FNV-1a over UTF-16 code units, rendered as 8 hex chars repeated to
 * 64. This is the placeholder the preview client always used for identity
 * commitments and nullifiers. It is NOT MiMC, NOT BN254 and has no
 * cryptographic strength; it exists only to generate preview identities and
 * to check that legacy preview records are internally consistent. Never use
 * it for anything submitted on-chain.
 */
export function previewMockIdentityHash(input: string): string {
  let hash = 0x811c9dc5;
  for (let i = 0; i < input.length; i++) {
    hash ^= input.charCodeAt(i);
    hash = Math.imul(hash, 0x01000193);
  }
  const hex = (hash >>> 0).toString(16).padStart(8, '0');
  // Pad to 64 chars (32 bytes) for consistency with real hashes
  return (hex + hex + hex + hex + hex + hex + hex + hex).slice(0, 64);
}
