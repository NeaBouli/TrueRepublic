import type { ChainConfig } from '@/types/chain';

export const TRUEREPUBLIC_MAINNET: ChainConfig = {
  chainId: 'truerepublic-1',
  chainName: 'TrueRepublic',
  rpc: 'http://localhost:26657',
  rest: 'http://localhost:1317',
  bech32Prefix: 'true',
  coinDenom: 'PNYX',
  coinMinimalDenom: 'pnyx',
  coinDecimals: 6,
  gasPrice: '0.025pnyx',
};

export const TRUEREPUBLIC_TESTNET: ChainConfig = {
  chainId: 'truerepublic-testnet',
  chainName: 'TrueRepublic Testnet',
  rpc: 'http://localhost:26657',
  rest: 'http://localhost:1317',
  bech32Prefix: 'true',
  coinDenom: 'PNYX',
  coinMinimalDenom: 'pnyx',
  coinDecimals: 6,
  gasPrice: '0.025pnyx',
};

export const DEFAULT_CHAIN = TRUEREPUBLIC_MAINNET;
