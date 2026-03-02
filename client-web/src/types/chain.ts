export interface ChainConfig {
  chainId: string;
  chainName: string;
  rpc: string;
  rest: string;
  bech32Prefix: string;
  coinDenom: string;
  coinMinimalDenom: string;
  coinDecimals: number;
  gasPrice: string;
}
