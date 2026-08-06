import type { SigningStargateClient } from '@cosmjs/stargate';
import { describe, expect, it, vi } from 'vitest';
import { deliverMessages } from './signingClient';

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
