/**
 * Canonical custom transaction registry for the maintained client (GH-115).
 *
 * This module is the single source of truth for the chain's custom protobuf
 * message identities and their exact wire encoding. Field numbers, wire
 * types, and names mirror the Go gogoproto definitions byte for byte:
 *
 * - x/truedemocracy/msgs.go  (package truedemocracy, type URLs /truedemocracy.Msg*)
 * - x/dex/msgs.go            (package dex, type URLs /dex.Msg*)
 *
 * Encoding rules required by those Go structs:
 * - `sdk.AccAddress` fields are protobuf `bytes` and carry the raw address
 *   bytes (bech32-decoded), never the bech32 string.
 * - Plain Go `string` fields (including bech32-valued strings such as
 *   `creator`, `member_addr`, `requester_addr`, `new_member`) are protobuf
 *   length-delimited string fields.
 * - Go `int64` fields are protobuf varints. TS values are canonical decimal
 *   strings validated against the signed 64-bit range so no unsafe JS-number
 *   truncation can occur.
 * - `sdk.Coins` is `repeated cosmos.base.v1beta1.Coin`.
 *
 * Unknown custom type URLs fail closed: the registry contains exactly the
 * types below plus the CosmJS default types (which include
 * /cosmos.bank.v1beta1.MsgSend).
 *
 * ICS-20 transfers (GH-190) additionally pin the upstream cosmjs-types
 * MsgTransfer codec at its canonical type URL so the signing path does not
 * depend on the contents of the CosmJS default type list.
 */
import { BinaryReader, BinaryWriter } from 'cosmjs-types/binary';
import { Coin } from 'cosmjs-types/cosmos/base/v1beta1/coin';
import { MsgTransfer } from 'cosmjs-types/ibc/applications/transfer/v1/tx';
import { Registry, type GeneratedType } from '@cosmjs/proto-signing';
import { defaultRegistryTypes } from '@cosmjs/stargate';

/** Largest value a Go int64 can hold, as a decimal string bound. */
const MAX_INT64 = 9223372036854775807n;

/**
 * Validate and normalize a chain amount as a positive signed-int64 decimal
 * string. Rejects non-decimal input, zero, negative values, and anything
 * above the int64 upper bound so chain amounts can never silently truncate
 * or wrap. Returns the canonical decimal form (no leading zeros).
 */
export function assertPositiveInt64Decimal(
  value: string,
  field: string
): string {
  if (typeof value !== 'string' || !/^[0-9]+$/.test(value)) {
    throw new Error(`${field} must be a positive integer decimal string`);
  }
  const parsed = BigInt(value);
  if (parsed <= 0n) {
    throw new Error(`${field} must be positive`);
  }
  if (parsed > MAX_INT64) {
    throw new Error(`${field} exceeds the signed 64-bit maximum`);
  }
  return parsed.toString();
}

/* ------------------------------------------------------------------------ */
/* truedemocracy messages                                                    */
/* ------------------------------------------------------------------------ */

/** Go: MsgCreateDomain { name(1), admin AccAddress(2), initial_coins Coins(3) } */
export interface MsgCreateDomain {
  name: string;
  admin: Uint8Array;
  initialCoins: Coin[];
}

function createBaseMsgCreateDomain(): MsgCreateDomain {
  return { name: '', admin: new Uint8Array(0), initialCoins: [] };
}

export const MsgCreateDomain = {
  typeUrl: '/truedemocracy.MsgCreateDomain',
  encode(
    message: MsgCreateDomain,
    writer: BinaryWriter = BinaryWriter.create()
  ): BinaryWriter {
    if (message.name !== '') {
      writer.uint32(10).string(message.name);
    }
    if (message.admin.length !== 0) {
      writer.uint32(18).bytes(message.admin);
    }
    for (const coin of message.initialCoins) {
      Coin.encode(coin, writer.uint32(26).fork()).ldelim();
    }
    return writer;
  },
  decode(input: Uint8Array, length?: number): MsgCreateDomain {
    const reader =
      input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgCreateDomain();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.name = reader.string();
          break;
        case 2:
          message.admin = reader.bytes();
          break;
        case 3:
          message.initialCoins.push(Coin.decode(reader, reader.uint32()));
          break;
        default:
          reader.skip(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object: Partial<MsgCreateDomain>): MsgCreateDomain {
    return {
      name: object.name ?? '',
      admin: object.admin ?? new Uint8Array(0),
      initialCoins: Array.isArray(object.initialCoins)
        ? object.initialCoins.map((coin) => Coin.fromPartial(coin))
        : [],
    };
  },
};

/**
 * Go: MsgSubmitProposal { sender(1), domain_name(2), issue_name(3),
 * suggestion_name(4), creator string(5), fee Coins(6), external_link(7) }
 */
export interface MsgSubmitProposal {
  sender: Uint8Array;
  domainName: string;
  issueName: string;
  suggestionName: string;
  creator: string;
  fee: Coin[];
  externalLink: string;
}

function createBaseMsgSubmitProposal(): MsgSubmitProposal {
  return {
    sender: new Uint8Array(0),
    domainName: '',
    issueName: '',
    suggestionName: '',
    creator: '',
    fee: [],
    externalLink: '',
  };
}

export const MsgSubmitProposal = {
  typeUrl: '/truedemocracy.MsgSubmitProposal',
  encode(
    message: MsgSubmitProposal,
    writer: BinaryWriter = BinaryWriter.create()
  ): BinaryWriter {
    if (message.sender.length !== 0) {
      writer.uint32(10).bytes(message.sender);
    }
    if (message.domainName !== '') {
      writer.uint32(18).string(message.domainName);
    }
    if (message.issueName !== '') {
      writer.uint32(26).string(message.issueName);
    }
    if (message.suggestionName !== '') {
      writer.uint32(34).string(message.suggestionName);
    }
    if (message.creator !== '') {
      writer.uint32(42).string(message.creator);
    }
    for (const coin of message.fee) {
      Coin.encode(coin, writer.uint32(50).fork()).ldelim();
    }
    if (message.externalLink !== '') {
      writer.uint32(58).string(message.externalLink);
    }
    return writer;
  },
  decode(input: Uint8Array, length?: number): MsgSubmitProposal {
    const reader =
      input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgSubmitProposal();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sender = reader.bytes();
          break;
        case 2:
          message.domainName = reader.string();
          break;
        case 3:
          message.issueName = reader.string();
          break;
        case 4:
          message.suggestionName = reader.string();
          break;
        case 5:
          message.creator = reader.string();
          break;
        case 6:
          message.fee.push(Coin.decode(reader, reader.uint32()));
          break;
        case 7:
          message.externalLink = reader.string();
          break;
        default:
          reader.skip(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object: Partial<MsgSubmitProposal>): MsgSubmitProposal {
    return {
      sender: object.sender ?? new Uint8Array(0),
      domainName: object.domainName ?? '',
      issueName: object.issueName ?? '',
      suggestionName: object.suggestionName ?? '',
      creator: object.creator ?? '',
      fee: Array.isArray(object.fee)
        ? object.fee.map((coin) => Coin.fromPartial(coin))
        : [],
      externalLink: object.externalLink ?? '',
    };
  },
};

/**
 * Go: MsgPlaceStoneOnSuggestion { sender(1), domain_name(2), issue_name(3),
 * suggestion_name(4), member_addr string(5) }
 */
export interface MsgPlaceStoneOnSuggestion {
  sender: Uint8Array;
  domainName: string;
  issueName: string;
  suggestionName: string;
  memberAddr: string;
}

function createBaseMsgPlaceStoneOnSuggestion(): MsgPlaceStoneOnSuggestion {
  return {
    sender: new Uint8Array(0),
    domainName: '',
    issueName: '',
    suggestionName: '',
    memberAddr: '',
  };
}

export const MsgPlaceStoneOnSuggestion = {
  typeUrl: '/truedemocracy.MsgPlaceStoneOnSuggestion',
  encode(
    message: MsgPlaceStoneOnSuggestion,
    writer: BinaryWriter = BinaryWriter.create()
  ): BinaryWriter {
    if (message.sender.length !== 0) {
      writer.uint32(10).bytes(message.sender);
    }
    if (message.domainName !== '') {
      writer.uint32(18).string(message.domainName);
    }
    if (message.issueName !== '') {
      writer.uint32(26).string(message.issueName);
    }
    if (message.suggestionName !== '') {
      writer.uint32(34).string(message.suggestionName);
    }
    if (message.memberAddr !== '') {
      writer.uint32(42).string(message.memberAddr);
    }
    return writer;
  },
  decode(input: Uint8Array, length?: number): MsgPlaceStoneOnSuggestion {
    const reader =
      input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgPlaceStoneOnSuggestion();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sender = reader.bytes();
          break;
        case 2:
          message.domainName = reader.string();
          break;
        case 3:
          message.issueName = reader.string();
          break;
        case 4:
          message.suggestionName = reader.string();
          break;
        case 5:
          message.memberAddr = reader.string();
          break;
        default:
          reader.skip(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(
    object: Partial<MsgPlaceStoneOnSuggestion>
  ): MsgPlaceStoneOnSuggestion {
    return {
      sender: object.sender ?? new Uint8Array(0),
      domainName: object.domainName ?? '',
      issueName: object.issueName ?? '',
      suggestionName: object.suggestionName ?? '',
      memberAddr: object.memberAddr ?? '',
    };
  },
};

/**
 * Go: MsgPlaceStoneOnIssue { sender(1), domain_name(2), issue_name(3),
 * member_addr string(4) }
 */
export interface MsgPlaceStoneOnIssue {
  sender: Uint8Array;
  domainName: string;
  issueName: string;
  memberAddr: string;
}

function createBaseMsgPlaceStoneOnIssue(): MsgPlaceStoneOnIssue {
  return {
    sender: new Uint8Array(0),
    domainName: '',
    issueName: '',
    memberAddr: '',
  };
}

export const MsgPlaceStoneOnIssue = {
  typeUrl: '/truedemocracy.MsgPlaceStoneOnIssue',
  encode(
    message: MsgPlaceStoneOnIssue,
    writer: BinaryWriter = BinaryWriter.create()
  ): BinaryWriter {
    if (message.sender.length !== 0) {
      writer.uint32(10).bytes(message.sender);
    }
    if (message.domainName !== '') {
      writer.uint32(18).string(message.domainName);
    }
    if (message.issueName !== '') {
      writer.uint32(26).string(message.issueName);
    }
    if (message.memberAddr !== '') {
      writer.uint32(34).string(message.memberAddr);
    }
    return writer;
  },
  decode(input: Uint8Array, length?: number): MsgPlaceStoneOnIssue {
    const reader =
      input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgPlaceStoneOnIssue();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sender = reader.bytes();
          break;
        case 2:
          message.domainName = reader.string();
          break;
        case 3:
          message.issueName = reader.string();
          break;
        case 4:
          message.memberAddr = reader.string();
          break;
        default:
          reader.skip(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object: Partial<MsgPlaceStoneOnIssue>): MsgPlaceStoneOnIssue {
    return {
      sender: object.sender ?? new Uint8Array(0),
      domainName: object.domainName ?? '',
      issueName: object.issueName ?? '',
      memberAddr: object.memberAddr ?? '',
    };
  },
};

/**
 * Go: MsgApproveOnboarding { sender(1), domain_name(2),
 * requester_addr string(3) }
 */
export interface MsgApproveOnboarding {
  sender: Uint8Array;
  domainName: string;
  requesterAddr: string;
}

function createBaseMsgApproveOnboarding(): MsgApproveOnboarding {
  return { sender: new Uint8Array(0), domainName: '', requesterAddr: '' };
}

export const MsgApproveOnboarding = {
  typeUrl: '/truedemocracy.MsgApproveOnboarding',
  encode(
    message: MsgApproveOnboarding,
    writer: BinaryWriter = BinaryWriter.create()
  ): BinaryWriter {
    if (message.sender.length !== 0) {
      writer.uint32(10).bytes(message.sender);
    }
    if (message.domainName !== '') {
      writer.uint32(18).string(message.domainName);
    }
    if (message.requesterAddr !== '') {
      writer.uint32(26).string(message.requesterAddr);
    }
    return writer;
  },
  decode(input: Uint8Array, length?: number): MsgApproveOnboarding {
    const reader =
      input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgApproveOnboarding();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sender = reader.bytes();
          break;
        case 2:
          message.domainName = reader.string();
          break;
        case 3:
          message.requesterAddr = reader.string();
          break;
        default:
          reader.skip(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object: Partial<MsgApproveOnboarding>): MsgApproveOnboarding {
    return {
      sender: object.sender ?? new Uint8Array(0),
      domainName: object.domainName ?? '',
      requesterAddr: object.requesterAddr ?? '',
    };
  },
};

/** Go: MsgAddMember { sender(1), domain_name(2), new_member string(3) } */
export interface MsgAddMember {
  sender: Uint8Array;
  domainName: string;
  newMember: string;
}

function createBaseMsgAddMember(): MsgAddMember {
  return { sender: new Uint8Array(0), domainName: '', newMember: '' };
}

export const MsgAddMember = {
  typeUrl: '/truedemocracy.MsgAddMember',
  encode(
    message: MsgAddMember,
    writer: BinaryWriter = BinaryWriter.create()
  ): BinaryWriter {
    if (message.sender.length !== 0) {
      writer.uint32(10).bytes(message.sender);
    }
    if (message.domainName !== '') {
      writer.uint32(18).string(message.domainName);
    }
    if (message.newMember !== '') {
      writer.uint32(26).string(message.newMember);
    }
    return writer;
  },
  decode(input: Uint8Array, length?: number): MsgAddMember {
    const reader =
      input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgAddMember();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sender = reader.bytes();
          break;
        case 2:
          message.domainName = reader.string();
          break;
        case 3:
          message.newMember = reader.string();
          break;
        default:
          reader.skip(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object: Partial<MsgAddMember>): MsgAddMember {
    return {
      sender: object.sender ?? new Uint8Array(0),
      domainName: object.domainName ?? '',
      newMember: object.newMember ?? '',
    };
  },
};

/**
 * Go: MsgOnboardToDomain { sender(1), domain_name(2), domain_pub_key_hex(3),
 * global_pub_key_hex(4), signature_hex(5) }
 */
export interface MsgOnboardToDomain {
  sender: Uint8Array;
  domainName: string;
  domainPubKeyHex: string;
  globalPubKeyHex: string;
  signatureHex: string;
}

function createBaseMsgOnboardToDomain(): MsgOnboardToDomain {
  return {
    sender: new Uint8Array(0),
    domainName: '',
    domainPubKeyHex: '',
    globalPubKeyHex: '',
    signatureHex: '',
  };
}

export const MsgOnboardToDomain = {
  typeUrl: '/truedemocracy.MsgOnboardToDomain',
  encode(
    message: MsgOnboardToDomain,
    writer: BinaryWriter = BinaryWriter.create()
  ): BinaryWriter {
    if (message.sender.length !== 0) {
      writer.uint32(10).bytes(message.sender);
    }
    if (message.domainName !== '') {
      writer.uint32(18).string(message.domainName);
    }
    if (message.domainPubKeyHex !== '') {
      writer.uint32(26).string(message.domainPubKeyHex);
    }
    if (message.globalPubKeyHex !== '') {
      writer.uint32(34).string(message.globalPubKeyHex);
    }
    if (message.signatureHex !== '') {
      writer.uint32(42).string(message.signatureHex);
    }
    return writer;
  },
  decode(input: Uint8Array, length?: number): MsgOnboardToDomain {
    const reader =
      input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgOnboardToDomain();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sender = reader.bytes();
          break;
        case 2:
          message.domainName = reader.string();
          break;
        case 3:
          message.domainPubKeyHex = reader.string();
          break;
        case 4:
          message.globalPubKeyHex = reader.string();
          break;
        case 5:
          message.signatureHex = reader.string();
          break;
        default:
          reader.skip(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object: Partial<MsgOnboardToDomain>): MsgOnboardToDomain {
    return {
      sender: object.sender ?? new Uint8Array(0),
      domainName: object.domainName ?? '',
      domainPubKeyHex: object.domainPubKeyHex ?? '',
      globalPubKeyHex: object.globalPubKeyHex ?? '',
      signatureHex: object.signatureHex ?? '',
    };
  },
};

/** Go: MsgRegisterIdentity { sender(1), domain_name(2), commitment(3) } */
export interface MsgRegisterIdentity {
  sender: Uint8Array;
  domainName: string;
  commitment: string;
}

function createBaseMsgRegisterIdentity(): MsgRegisterIdentity {
  return { sender: new Uint8Array(0), domainName: '', commitment: '' };
}

export const MsgRegisterIdentity = {
  typeUrl: '/truedemocracy.MsgRegisterIdentity',
  encode(
    message: MsgRegisterIdentity,
    writer: BinaryWriter = BinaryWriter.create()
  ): BinaryWriter {
    if (message.sender.length !== 0) {
      writer.uint32(10).bytes(message.sender);
    }
    if (message.domainName !== '') {
      writer.uint32(18).string(message.domainName);
    }
    if (message.commitment !== '') {
      writer.uint32(26).string(message.commitment);
    }
    return writer;
  },
  decode(input: Uint8Array, length?: number): MsgRegisterIdentity {
    const reader =
      input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgRegisterIdentity();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sender = reader.bytes();
          break;
        case 2:
          message.domainName = reader.string();
          break;
        case 3:
          message.commitment = reader.string();
          break;
        default:
          reader.skip(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object: Partial<MsgRegisterIdentity>): MsgRegisterIdentity {
    return {
      sender: object.sender ?? new Uint8Array(0),
      domainName: object.domainName ?? '',
      commitment: object.commitment ?? '',
    };
  },
};

/* ------------------------------------------------------------------------ */
/* dex messages                                                              */
/* ------------------------------------------------------------------------ */

/**
 * Go: MsgAddLiquidity { sender(1), asset_denom(2), pnyx_amt int64(3),
 * asset_amt int64(4) }
 */
export interface MsgAddLiquidity {
  sender: Uint8Array;
  assetDenom: string;
  /** Canonical positive signed-int64 decimal string. */
  pnyxAmt: string;
  /** Canonical positive signed-int64 decimal string. */
  assetAmt: string;
}

function createBaseMsgAddLiquidity(): MsgAddLiquidity {
  return {
    sender: new Uint8Array(0),
    assetDenom: '',
    pnyxAmt: '0',
    assetAmt: '0',
  };
}

export const MsgAddLiquidity = {
  typeUrl: '/dex.MsgAddLiquidity',
  encode(
    message: MsgAddLiquidity,
    writer: BinaryWriter = BinaryWriter.create()
  ): BinaryWriter {
    if (message.sender.length !== 0) {
      writer.uint32(10).bytes(message.sender);
    }
    if (message.assetDenom !== '') {
      writer.uint32(18).string(message.assetDenom);
    }
    writer
      .uint32(24)
      .int64(assertPositiveInt64Decimal(message.pnyxAmt, 'pnyx_amt'));
    writer
      .uint32(32)
      .int64(assertPositiveInt64Decimal(message.assetAmt, 'asset_amt'));
    return writer;
  },
  decode(input: Uint8Array, length?: number): MsgAddLiquidity {
    const reader =
      input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgAddLiquidity();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sender = reader.bytes();
          break;
        case 2:
          message.assetDenom = reader.string();
          break;
        case 3:
          message.pnyxAmt = reader.int64().toString();
          break;
        case 4:
          message.assetAmt = reader.int64().toString();
          break;
        default:
          reader.skip(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object: Partial<MsgAddLiquidity>): MsgAddLiquidity {
    return {
      sender: object.sender ?? new Uint8Array(0),
      assetDenom: object.assetDenom ?? '',
      pnyxAmt: object.pnyxAmt ?? '0',
      assetAmt: object.assetAmt ?? '0',
    };
  },
};

/**
 * Go: MsgRemoveLiquidity { sender(1), asset_denom(2), shares int64(3) }
 */
export interface MsgRemoveLiquidity {
  sender: Uint8Array;
  assetDenom: string;
  /** Canonical positive signed-int64 decimal string. */
  shares: string;
}

function createBaseMsgRemoveLiquidity(): MsgRemoveLiquidity {
  return { sender: new Uint8Array(0), assetDenom: '', shares: '0' };
}

export const MsgRemoveLiquidity = {
  typeUrl: '/dex.MsgRemoveLiquidity',
  encode(
    message: MsgRemoveLiquidity,
    writer: BinaryWriter = BinaryWriter.create()
  ): BinaryWriter {
    if (message.sender.length !== 0) {
      writer.uint32(10).bytes(message.sender);
    }
    if (message.assetDenom !== '') {
      writer.uint32(18).string(message.assetDenom);
    }
    writer
      .uint32(24)
      .int64(assertPositiveInt64Decimal(message.shares, 'shares'));
    return writer;
  },
  decode(input: Uint8Array, length?: number): MsgRemoveLiquidity {
    const reader =
      input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgRemoveLiquidity();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sender = reader.bytes();
          break;
        case 2:
          message.assetDenom = reader.string();
          break;
        case 3:
          message.shares = reader.int64().toString();
          break;
        default:
          reader.skip(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object: Partial<MsgRemoveLiquidity>): MsgRemoveLiquidity {
    return {
      sender: object.sender ?? new Uint8Array(0),
      assetDenom: object.assetDenom ?? '',
      shares: object.shares ?? '0',
    };
  },
};

/**
 * Go: MsgSwapExact { sender(1), input_denom(2), input_amt int64(3),
 * output_denom(4), min_output int64(5) }
 */
export interface MsgSwapExact {
  sender: Uint8Array;
  inputDenom: string;
  /** Canonical positive signed-int64 decimal string. */
  inputAmt: string;
  outputDenom: string;
  /** Canonical positive signed-int64 decimal string (slippage bound). */
  minOutput: string;
}

function createBaseMsgSwapExact(): MsgSwapExact {
  return {
    sender: new Uint8Array(0),
    inputDenom: '',
    inputAmt: '0',
    outputDenom: '',
    minOutput: '0',
  };
}

export const MsgSwapExact = {
  typeUrl: '/dex.MsgSwapExact',
  encode(
    message: MsgSwapExact,
    writer: BinaryWriter = BinaryWriter.create()
  ): BinaryWriter {
    if (message.sender.length !== 0) {
      writer.uint32(10).bytes(message.sender);
    }
    if (message.inputDenom !== '') {
      writer.uint32(18).string(message.inputDenom);
    }
    writer
      .uint32(24)
      .int64(assertPositiveInt64Decimal(message.inputAmt, 'input_amt'));
    if (message.outputDenom !== '') {
      writer.uint32(34).string(message.outputDenom);
    }
    writer
      .uint32(40)
      .int64(assertPositiveInt64Decimal(message.minOutput, 'min_output'));
    return writer;
  },
  decode(input: Uint8Array, length?: number): MsgSwapExact {
    const reader =
      input instanceof BinaryReader ? input : new BinaryReader(input);
    const end = length === undefined ? reader.len : reader.pos + length;
    const message = createBaseMsgSwapExact();
    while (reader.pos < end) {
      const tag = reader.uint32();
      switch (tag >>> 3) {
        case 1:
          message.sender = reader.bytes();
          break;
        case 2:
          message.inputDenom = reader.string();
          break;
        case 3:
          message.inputAmt = reader.int64().toString();
          break;
        case 4:
          message.outputDenom = reader.string();
          break;
        case 5:
          message.minOutput = reader.int64().toString();
          break;
        default:
          reader.skip(tag & 7);
          break;
      }
    }
    return message;
  },
  fromPartial(object: Partial<MsgSwapExact>): MsgSwapExact {
    return {
      sender: object.sender ?? new Uint8Array(0),
      inputDenom: object.inputDenom ?? '',
      inputAmt: object.inputAmt ?? '0',
      outputDenom: object.outputDenom ?? '',
      minOutput: object.minOutput ?? '0',
    };
  },
};

/* ------------------------------------------------------------------------ */
/* registry                                                                  */
/* ------------------------------------------------------------------------ */

/**
 * The canonical ICS-20 transfer codec identity (GH-190). Registered
 * explicitly so the IBC signing path is covered by the fail-closed registry
 * even if the CosmJS default type list changes.
 */
export const ibcRegistryTypes: ReadonlyArray<[string, GeneratedType]> = [
  [MsgTransfer.typeUrl, MsgTransfer],
];

/**
 * The exact custom type identities this client may sign. Every other custom
 * type URL fails closed in the registry.
 */
export const customRegistryTypes: ReadonlyArray<[string, GeneratedType]> = [
  [MsgCreateDomain.typeUrl, MsgCreateDomain],
  [MsgSubmitProposal.typeUrl, MsgSubmitProposal],
  [MsgPlaceStoneOnSuggestion.typeUrl, MsgPlaceStoneOnSuggestion],
  [MsgPlaceStoneOnIssue.typeUrl, MsgPlaceStoneOnIssue],
  [MsgApproveOnboarding.typeUrl, MsgApproveOnboarding],
  [MsgAddMember.typeUrl, MsgAddMember],
  [MsgOnboardToDomain.typeUrl, MsgOnboardToDomain],
  [MsgRegisterIdentity.typeUrl, MsgRegisterIdentity],
  [MsgAddLiquidity.typeUrl, MsgAddLiquidity],
  [MsgRemoveLiquidity.typeUrl, MsgRemoveLiquidity],
  [MsgSwapExact.typeUrl, MsgSwapExact],
];

/**
 * Build the canonical signing registry: CosmJS default types (including
 * /cosmos.bank.v1beta1.MsgSend) plus the exact TrueRepublic custom types.
 */
export function createTxRegistry(): Registry {
  return new Registry([
    ...defaultRegistryTypes,
    ...ibcRegistryTypes,
    ...customRegistryTypes,
  ]);
}
