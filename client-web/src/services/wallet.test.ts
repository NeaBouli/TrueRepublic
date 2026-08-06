import { toBech32 } from '@cosmjs/encoding';
import { beforeEach, describe, expect, it } from 'vitest';
import { WalletService } from './wallet';

const STORAGE_KEY = 'truerepublic_wallets';

describe('wallet address migration', () => {
  beforeEach(() => localStorage.clear());

  it('rewrites the retired true prefix without changing encrypted data', () => {
    const data = Uint8Array.from({ length: 20 }, (_, index) => index + 1);
    const stored = {
      address: toBech32('true', data),
      mnemonic: 'encrypted-wallet-payload',
      name: 'Legacy wallet',
      createdAt: 1,
    };
    localStorage.setItem(STORAGE_KEY, JSON.stringify([stored]));

    const [wallet] = WalletService.loadWallets();

    expect(wallet).toEqual({
      ...stored,
      address: toBech32('truerepublic', data),
    });
    expect(JSON.parse(localStorage.getItem(STORAGE_KEY) ?? '[]')).toEqual([
      wallet,
    ]);
  });

  it('preserves canonical and malformed historical records', () => {
    const canonical = {
      address: toBech32('truerepublic', new Uint8Array(20).fill(7)),
      mnemonic: 'encrypted-one',
      name: 'Canonical',
      createdAt: 2,
    };
    const malformed = {
      address: 'not-an-address',
      mnemonic: 'encrypted-two',
      name: 'Malformed',
      createdAt: 3,
    };
    localStorage.setItem(
      STORAGE_KEY,
      JSON.stringify([canonical, malformed])
    );

    expect(WalletService.loadWallets()).toEqual([canonical, malformed]);
  });
});
