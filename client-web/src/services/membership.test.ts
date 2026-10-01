import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { DirectSecp256k1HdWallet } from '@cosmjs/proto-signing';
import { DEFAULT_CHAIN } from '@/config/chains';

const signing = vi.hoisted(() => ({
  connectSigningClient: vi.fn(),
  deliverMessages: vi.fn(),
}));

vi.mock('./signingClient', () => signing);

import { MembershipService, PREVIEW_IDENTITY_REGISTRATION_DISABLED } from './membership';

describe('MembershipService.registerIdentity preview boundary (issue #309)', () => {
  beforeEach(() => {
    signing.connectSigningClient.mockReset();
    signing.deliverMessages.mockReset();
  });

  it('fails closed before touching the wallet, signing client or delivery', async () => {
    const getAccounts = vi.fn();
    const wallet = { getAccounts } as unknown as DirectSecp256k1HdWallet;

    const result = await new MembershipService(DEFAULT_CHAIN).registerIdentity(
      wallet,
      'GH309',
      `${'0'.repeat(63)}1`
    );

    expect(result).toEqual({
      hash: '',
      height: 0,
      success: false,
      error: PREVIEW_IDENTITY_REGISTRATION_DISABLED,
    });
    expect(getAccounts).not.toHaveBeenCalled();
    expect(signing.connectSigningClient).not.toHaveBeenCalled();
    expect(signing.deliverMessages).not.toHaveBeenCalled();
  });

  it('names the review gates in the stable preview-disabled error', () => {
    expect(PREVIEW_IDENTITY_REGISTRATION_DISABLED).toContain('issue #309');
    expect(PREVIEW_IDENTITY_REGISTRATION_DISABLED).toContain('custody');
    expect(PREVIEW_IDENTITY_REGISTRATION_DISABLED).toContain('prover');
  });
});
