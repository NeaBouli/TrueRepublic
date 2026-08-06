import { describe, expect, it } from 'vitest';
import { DEFAULT_CHAIN, resolveChainEndpoint } from './chains';

describe('PNYX chain metadata', () => {
  it('uses the canonical six-decimal base denomination', () => {
    expect(DEFAULT_CHAIN.coinDenom).toBe('PNYX');
    expect(DEFAULT_CHAIN.coinMinimalDenom).toBe('upnyx');
    expect(DEFAULT_CHAIN.coinDecimals).toBe(6);
    expect(DEFAULT_CHAIN.gasPrice.endsWith('upnyx')).toBe(true);
    expect(DEFAULT_CHAIN.gasPrice).toBe('25000upnyx');
  });

  it('resolves same-origin proxy paths for container builds', () => {
    expect(resolveChainEndpoint('/rpc', 'http://localhost:26657')).toBe(
      `${window.location.origin}/rpc`
    );
    expect(resolveChainEndpoint('/api', 'http://localhost:1317')).toBe(
      `${window.location.origin}/api`
    );
    expect(resolveChainEndpoint('https://rpc.example/', 'http://localhost:26657')).toBe(
      'https://rpc.example'
    );
  });
});
