import { describe, expect, it } from 'vitest';
import { DEFAULT_CHAIN } from './chains';

describe('PNYX chain metadata', () => {
  it('uses the canonical six-decimal base denomination', () => {
    expect(DEFAULT_CHAIN.coinDenom).toBe('PNYX');
    expect(DEFAULT_CHAIN.coinMinimalDenom).toBe('upnyx');
    expect(DEFAULT_CHAIN.coinDecimals).toBe(6);
    expect(DEFAULT_CHAIN.gasPrice.endsWith('upnyx')).toBe(true);
  });
});
