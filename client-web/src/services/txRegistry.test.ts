import { describe, expect, it } from 'vitest';
import { BinaryWriter } from 'cosmjs-types/binary';
import { MsgTransfer } from 'cosmjs-types/ibc/applications/transfer/v1/tx';
import {
  assertPositiveInt64Decimal,
  createTxRegistry,
  customRegistryTypes,
  ibcRegistryTypes,
  MsgAddLiquidity,
  MsgAddMember,
  MsgApproveOnboarding,
  MsgCreateDomain,
  MsgOnboardToDomain,
  MsgPlaceStoneOnIssue,
  MsgPlaceStoneOnSuggestion,
  MsgRegisterIdentity,
  MsgRemoveLiquidity,
  MsgSubmitProposal,
  MsgSwapExact,
} from './txRegistry';

function toHex(bytes: Uint8Array): string {
  return Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('');
}

function encode<T>(
  codec: {
    encode: (message: T, writer?: BinaryWriter) => BinaryWriter;
  },
  value: T
): string {
  return toHex(codec.encode(value, BinaryWriter.create()).finish());
}

/**
 * Deterministic wire-byte vectors hand-derived from the Go gogoproto tags in
 * x/truedemocracy/msgs.go and x/dex/msgs.go (field number + wire type per
 * field, proto3 default omission, sdk.AccAddress as raw bytes, sdk.Coins as
 * repeated cosmos.base.v1beta1.Coin, int64 as varint).
 */
describe('custom message wire encoding', () => {
  it('encodes MsgCreateDomain exactly', () => {
    expect(
      encode(MsgCreateDomain, {
        name: 'dom',
        admin: Uint8Array.from([1, 2, 3, 4]),
        initialCoins: [{ denom: 'upnyx', amount: '42' }],
      })
    ).toBe('0a03646f6d1204010203041a0b0a0575706e797812023432');
  });

  it('omits proto3 defaults (empty sdk.Coins)', () => {
    expect(
      encode(MsgCreateDomain, {
        name: 'dom',
        admin: Uint8Array.from([1, 2, 3, 4]),
        initialCoins: [],
      })
    ).toBe('0a03646f6d120401020304');
  });

  it('encodes MsgSubmitProposal exactly', () => {
    expect(
      encode(MsgSubmitProposal, {
        sender: Uint8Array.from([9, 8]),
        domainName: 'd',
        issueName: 'i',
        suggestionName: 's',
        creator: 'c',
        fee: [{ denom: 'upnyx', amount: '7' }],
        externalLink: 'L',
      })
    ).toBe(
      '0a0209081201641a01692201732a0163320a0a0575706e79781201373a014c'
    );
  });

  it('encodes MsgPlaceStoneOnSuggestion exactly', () => {
    expect(
      encode(MsgPlaceStoneOnSuggestion, {
        sender: Uint8Array.from([1]),
        domainName: 'd',
        issueName: 'i',
        suggestionName: 's',
        memberAddr: 'm',
      })
    ).toBe('0a01011201641a01692201732a016d');
  });

  it('encodes MsgPlaceStoneOnIssue exactly', () => {
    expect(
      encode(MsgPlaceStoneOnIssue, {
        sender: Uint8Array.from([1]),
        domainName: 'd',
        issueName: 'i',
        memberAddr: 'm',
      })
    ).toBe('0a01011201641a016922016d');
  });

  it('encodes MsgApproveOnboarding exactly', () => {
    expect(
      encode(MsgApproveOnboarding, {
        sender: Uint8Array.from([1]),
        domainName: 'd',
        requesterAddr: 'r',
      })
    ).toBe('0a01011201641a0172');
  });

  it('encodes MsgAddMember exactly', () => {
    expect(
      encode(MsgAddMember, {
        sender: Uint8Array.from([1]),
        domainName: 'd',
        newMember: 'n',
      })
    ).toBe('0a01011201641a016e');
  });

  it('encodes MsgOnboardToDomain exactly', () => {
    expect(
      encode(MsgOnboardToDomain, {
        sender: Uint8Array.from([1]),
        domainName: 'd',
        domainPubKeyHex: 'dp',
        globalPubKeyHex: 'gp',
        signatureHex: 'sg',
      })
    ).toBe('0a01011201641a026470220267702a027367');
  });

  it('encodes MsgRegisterIdentity exactly', () => {
    expect(
      encode(MsgRegisterIdentity, {
        sender: Uint8Array.from([1]),
        domainName: 'd',
        commitment: 'ci',
      })
    ).toBe('0a01011201641a026369');
  });

  it('encodes MsgAddLiquidity exactly (multi-byte varint)', () => {
    expect(
      encode(MsgAddLiquidity, {
        sender: Uint8Array.from([1]),
        assetDenom: 'atom',
        pnyxAmt: '300',
        assetAmt: '1',
      })
    ).toBe('0a0101120461746f6d18ac022001');
  });

  it('encodes MsgRemoveLiquidity exactly', () => {
    expect(
      encode(MsgRemoveLiquidity, {
        sender: Uint8Array.from([1]),
        assetDenom: 'atom',
        shares: '5',
      })
    ).toBe('0a0101120461746f6d1805');
  });

  it('encodes MsgSwapExact exactly', () => {
    expect(
      encode(MsgSwapExact, {
        sender: Uint8Array.from([1]),
        inputDenom: 'upnyx',
        inputAmt: '2',
        outputDenom: 'atom',
        minOutput: '1',
      })
    ).toBe('0a0101120575706e79781802220461746f6d2801');
  });

  it('encodes the int64 upper bound without truncation', () => {
    // 9223372036854775807 = 0x7FFFFFFFFFFFFFFF → 9-byte varint ff*8 7f
    expect(
      encode(MsgRemoveLiquidity, {
        sender: Uint8Array.from([1]),
        assetDenom: 'atom',
        shares: '9223372036854775807',
      })
    ).toBe('0a0101120461746f6d18ffffffffffffffff7f');
  });

  it('round-trips a representative message', () => {
    const value = {
      sender: Uint8Array.from([9, 8]),
      domainName: 'd',
      issueName: 'i',
      suggestionName: 's',
      creator: 'c',
      fee: [{ denom: 'upnyx', amount: '7' }],
      externalLink: 'L',
    };
    const bytes = MsgSubmitProposal.encode(
      value,
      BinaryWriter.create()
    ).finish();
    expect(MsgSubmitProposal.decode(bytes)).toEqual(value);
  });
});

describe('canonical IBC transfer registration (GH-190)', () => {
  const transferValue = {
    sourcePort: 'transfer',
    sourceChannel: 'channel-0',
    token: { denom: 'upnyx', amount: '1000000' },
    sender: 'truerepublic1sender',
    receiver: 'cosmos1receiver',
    timeoutHeight: { revisionNumber: 1n, revisionHeight: 1000n },
    timeoutTimestamp: 1700000000000000000n,
    memo: '',
    encoding: '',
  };

  it('pins the upstream MsgTransfer codec at its canonical type URL', () => {
    expect(ibcRegistryTypes.map(([typeUrl]) => typeUrl)).toEqual([
      '/ibc.applications.transfer.v1.MsgTransfer',
    ]);
    expect(ibcRegistryTypes[0][1]).toBe(MsgTransfer);
  });

  it('encodes MsgTransfer through the registry with the exact golden bytes', () => {
    // Field order and wire types mirror ibc-go applications/transfer/v1/tx.proto:
    // 1 source_port, 2 source_channel, 3 token (Coin), 4 sender, 5 receiver,
    // 6 timeout_height (Height), 7 timeout_timestamp (uint64 varint).
    // Proto3 default fields (memo, encoding) are omitted.
    const registry = createTxRegistry();
    const bytes = registry.encode({
      typeUrl: '/ibc.applications.transfer.v1.MsgTransfer',
      value: transferValue,
    });
    expect(toHex(bytes)).toBe(
      '0a087472616e7366657212096368616e6e656c2d30' +
        '1a100a0575706e7978120731303030303030' +
        '22137472756572657075626c69633173656e646572' +
        '2a0f636f736d6f73317265636569766572' +
        '3205080110e807' +
        '388080a8b1e39fe7cb17'
    );
    const decoded = MsgTransfer.decode(bytes);
    expect(decoded.sourceChannel).toBe('channel-0');
    expect(decoded.timeoutHeight.revisionHeight).toBe(1000n);
    expect(decoded.timeoutTimestamp).toBe(1700000000000000000n);
  });

  it('keeps rejecting unregistered IBC type URLs', () => {
    const registry = createTxRegistry();
    for (const typeUrl of [
      '/ibc.applications.transfer.v1.MsgTransferV2',
      '/ibc.applications.transfer.v2.MsgSendPacket',
      '/ibc.lightclients.tendermint.v1.MsgMisbehaviour',
    ]) {
      expect(() => registry.encode({ typeUrl, value: {} })).toThrow(
        /Unregistered type url/
      );
    }
  });
});

describe('assertPositiveInt64Decimal', () => {
  it('accepts canonical positive decimals and the int64 upper bound', () => {
    expect(assertPositiveInt64Decimal('1', 'amt')).toBe('1');
    expect(assertPositiveInt64Decimal('9223372036854775807', 'amt')).toBe(
      '9223372036854775807'
    );
  });

  it('normalizes leading zeros to the canonical form', () => {
    expect(assertPositiveInt64Decimal('007', 'amt')).toBe('7');
  });

  it.each([
    '0',
    '-1',
    '9223372036854775808',
    '99999999999999999999',
    '1.5',
    'abc',
    '',
    ' 1',
    '1e3',
    '+7',
  ])('rejects %j', (value) => {
    expect(() => assertPositiveInt64Decimal(value, 'amt')).toThrow();
  });
});

describe('canonical transaction registry', () => {
  it('registers every custom type identity', () => {
    const registry = createTxRegistry();
    expect(customRegistryTypes.map(([typeUrl]) => typeUrl)).toEqual([
      '/truedemocracy.MsgCreateDomain',
      '/truedemocracy.MsgSubmitProposal',
      '/truedemocracy.MsgPlaceStoneOnSuggestion',
      '/truedemocracy.MsgPlaceStoneOnIssue',
      '/truedemocracy.MsgApproveOnboarding',
      '/truedemocracy.MsgAddMember',
      '/truedemocracy.MsgOnboardToDomain',
      '/truedemocracy.MsgRegisterIdentity',
      '/dex.MsgAddLiquidity',
      '/dex.MsgRemoveLiquidity',
      '/dex.MsgSwapExact',
    ]);
    for (const [typeUrl] of customRegistryTypes) {
      expect(registry.lookupType(typeUrl)).toBeDefined();
    }
  });

  it('retains the standard bank MsgSend through default types', () => {
    const registry = createTxRegistry();
    const bytes = registry.encode({
      typeUrl: '/cosmos.bank.v1beta1.MsgSend',
      value: {
        fromAddress: 'a',
        toAddress: 'b',
        amount: [{ denom: 'upnyx', amount: '3' }],
      },
    });
    expect(toHex(bytes)).toBe('0a01611201621a0a0a0575706e7978120133');
  });

  it('encodes a registered custom type through the registry', () => {
    const registry = createTxRegistry();
    const bytes = registry.encode({
      typeUrl: '/dex.MsgSwapExact',
      value: {
        sender: Uint8Array.from([1]),
        inputDenom: 'upnyx',
        inputAmt: '2',
        outputDenom: 'atom',
        minOutput: '1',
      },
    });
    expect(toHex(bytes)).toBe('0a0101120575706e79781802220461746f6d2801');
  });

  it.each(['0', '-1', '1.5', '9223372036854775808'])(
    'rejects unsafe DEX amount %j at the registry boundary',
    (inputAmt) => {
      const registry = createTxRegistry();
      expect(() =>
        registry.encode({
          typeUrl: '/dex.MsgSwapExact',
          value: {
            sender: Uint8Array.from([1]),
            inputDenom: 'upnyx',
            inputAmt,
            outputDenom: 'atom',
            minOutput: '1',
          },
        })
      ).toThrow();
    }
  );

  it.each([
    '/truedemocracy.MsgVoteToExclude', // real chain message, not client-supported
    '/truedemocracy.MsgCreateDomains', // near-miss typo
    '/dex.MsgSwap', // real chain message, not client-supported
    '/truerepublic.truedemocracy.MsgSubmitProposal', // retired wrong identity
    '/foo.Bar',
  ])('fails closed for unknown type URL %j', (typeUrl) => {
    const registry = createTxRegistry();
    expect(() => registry.encode({ typeUrl, value: {} })).toThrow(
      /Unregistered type url/
    );
  });
});
