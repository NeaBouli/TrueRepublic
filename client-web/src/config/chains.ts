import type { ChainConfig } from '@/types/chain';

export function resolveChainEndpoint(configured: string | undefined, fallback: string): string {
  const endpoint = configured || fallback;
  if (endpoint.startsWith('/') && typeof window !== 'undefined') {
    return new URL(endpoint, window.location.origin).toString().replace(/\/$/, '');
  }
  return endpoint.replace(/\/$/, '');
}

export const TRUEREPUBLIC_MAINNET: ChainConfig = {
  chainId: 'truerepublic-1',
  chainName: 'TrueRepublic',
  rpc: resolveChainEndpoint(import.meta.env.VITE_TRUEREPUBLIC_RPC, 'http://localhost:26657'),
  rest: resolveChainEndpoint(import.meta.env.VITE_TRUEREPUBLIC_REST, 'http://localhost:1317'),
  bech32Prefix: 'true',
  coinDenom: 'PNYX',
  coinMinimalDenom: 'upnyx',
  coinDecimals: 6,
  gasPrice: '25000upnyx',
};

export const TRUEREPUBLIC_TESTNET: ChainConfig = {
  chainId: 'truerepublic-testnet',
  chainName: 'TrueRepublic Testnet',
  rpc: resolveChainEndpoint(import.meta.env.VITE_TRUEREPUBLIC_RPC, 'http://localhost:26657'),
  rest: resolveChainEndpoint(import.meta.env.VITE_TRUEREPUBLIC_REST, 'http://localhost:1317'),
  bech32Prefix: 'true',
  coinDenom: 'PNYX',
  coinMinimalDenom: 'upnyx',
  coinDecimals: 6,
  gasPrice: '25000upnyx',
};

export const DEFAULT_CHAIN = TRUEREPUBLIC_MAINNET;
