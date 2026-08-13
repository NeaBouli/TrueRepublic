import { SigningStargateClient } from '@cosmjs/stargate';
import { toBech32 } from '@cosmjs/encoding';
import type { OfflineSigner } from '@cosmjs/proto-signing';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { ChainConfig } from '@/types/chain';
import { connectSigningClient, deliverMessages } from './signingClient';

function mockClient(overrides?: {
  simulate?: ReturnType<typeof vi.fn>;
  signAndBroadcast?: ReturnType<typeof vi.fn>;
}): SigningStargateClient {
  return {
    simulate: overrides?.simulate ?? vi.fn().mockResolvedValue(100),
    signAndBroadcast:
      overrides?.signAndBroadcast ??
      vi.fn().mockResolvedValue({
        code: 0,
        transactionHash: 'ABC123',
        height: 7,
        rawLog: '',
      }),
  } as unknown as SigningStargateClient;
}

describe('canonical transaction delivery', () => {
  it('derives the fee from simulated gas and the configured gas price', async () => {
    const client = mockClient();

    await expect(
      deliverMessages(
        client,
        'truerepublic1sender',
        [{ typeUrl: '/cosmos.bank.v1beta1.MsgSend', value: {} }],
        '0.025upnyx'
      )
    ).resolves.toEqual({ hash: 'ABC123', height: 7, success: true });

    expect(client.signAndBroadcast).toHaveBeenCalledWith(
      'truerepublic1sender',
      expect.any(Array),
      {
        amount: [{ denom: 'upnyx', amount: '5' }],
        gas: '200',
      },
      ''
    );
  });

  it('fails closed when simulation fails and never broadcasts', async () => {
    const signAndBroadcast = vi.fn();
    const client = mockClient({
      simulate: vi.fn().mockRejectedValue(new Error('simulation rejected')),
      signAndBroadcast,
    });

    await expect(
      deliverMessages(client, 'truerepublic1sender', [], '0.025upnyx')
    ).rejects.toThrow('simulation rejected');
    expect(signAndBroadcast).not.toHaveBeenCalled();
  });

  it('surfaces a non-zero delivery code', async () => {
    const client = mockClient({
      signAndBroadcast: vi.fn().mockResolvedValue({
        code: 5,
        transactionHash: 'FAILED',
        height: 8,
        rawLog: 'chain rejected message',
      }),
    });

    await expect(
      deliverMessages(client, 'truerepublic1sender', [], '0.025upnyx')
    ).rejects.toThrow('chain rejected message');
  });

  it('retains the delivery code when the chain returns no raw log', async () => {
    const client = mockClient({
      signAndBroadcast: vi.fn().mockResolvedValue({
        code: 9,
        transactionHash: 'FAILED',
        height: 8,
        rawLog: '',
      }),
    });

    await expect(
      deliverMessages(client, 'truerepublic1sender', [], '0.025upnyx')
    ).rejects.toThrow('Transaction failed with code 9');
  });
});

/* ------------------------------------------------------------------------ */
/* connectSigningClient scoping (wallet/account/chain mismatch)              */
/* ------------------------------------------------------------------------ */

const SCOPING_CONFIG: ChainConfig = {
  chainId: 'truerepublic-1',
  chainName: 'TrueRepublic',
  rpc: 'http://localhost:26657',
  rest: 'http://localhost:1317',
  bech32Prefix: 'truerepublic',
  coinDenom: 'PNYX',
  coinMinimalDenom: 'upnyx',
  coinDecimals: 6,
  gasPrice: '25000upnyx',
};

const LOCAL_ACCOUNT = toBech32(
  'truerepublic',
  Uint8Array.from({ length: 20 }, (_, index) => index + 1)
);
const FOREIGN_ACCOUNT = toBech32('cosmos', new Uint8Array(20).fill(9));

function signerWith(addresses: string[]): OfflineSigner {
  return {
    getAccounts: async () =>
      addresses.map((address) => ({
        address,
        algo: 'secp256k1' as const,
        pubkey: new Uint8Array(33),
      })),
  } as unknown as OfflineSigner;
}

function connectedClient(chainId: string): SigningStargateClient {
  return {
    getChainId: vi.fn().mockResolvedValue(chainId),
    disconnect: vi.fn(),
  } as unknown as SigningStargateClient;
}

describe('signing client scoping', () => {
  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('connects when account prefix and chain id match the config', async () => {
    const client = connectedClient('truerepublic-1');
    const connect = vi
      .spyOn(SigningStargateClient, 'connectWithSigner')
      .mockResolvedValue(client);

    await expect(
      connectSigningClient(SCOPING_CONFIG, signerWith([LOCAL_ACCOUNT]))
    ).resolves.toBe(client);
    expect(connect).toHaveBeenCalledOnce();
  });

  it('rejects a signer without any account before connecting', async () => {
    const connect = vi.spyOn(SigningStargateClient, 'connectWithSigner');

    await expect(
      connectSigningClient(SCOPING_CONFIG, signerWith([]))
    ).rejects.toThrow('no account');
    expect(connect).not.toHaveBeenCalled();
  });

  it('rejects a foreign-network account before connecting', async () => {
    const connect = vi.spyOn(SigningStargateClient, 'connectWithSigner');

    await expect(
      connectSigningClient(SCOPING_CONFIG, signerWith([FOREIGN_ACCOUNT]))
    ).rejects.toThrow('does not belong to the truerepublic network');
    expect(connect).not.toHaveBeenCalled();
  });

  it('rejects a malformed account address before connecting', async () => {
    const connect = vi.spyOn(SigningStargateClient, 'connectWithSigner');

    await expect(
      connectSigningClient(SCOPING_CONFIG, signerWith(['not-an-address']))
    ).rejects.toThrow('not valid bech32');
    expect(connect).not.toHaveBeenCalled();
  });

  it('rejects a chain id mismatch and disconnects the client', async () => {
    const client = connectedClient('other-chain-9');
    vi.spyOn(SigningStargateClient, 'connectWithSigner').mockResolvedValue(
      client
    );

    await expect(
      connectSigningClient(SCOPING_CONFIG, signerWith([LOCAL_ACCOUNT]))
    ).rejects.toThrow('expected truerepublic-1');
    expect(client.disconnect).toHaveBeenCalledOnce();
  });
});
