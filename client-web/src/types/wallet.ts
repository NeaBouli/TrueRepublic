export interface Wallet {
  address: string;
  mnemonic?: string;
  name: string;
  createdAt: number;
}

export interface Balance {
  denom: string;
  amount: string;
}

export interface WalletState {
  currentWallet: Wallet | null;
  wallets: Wallet[];
  isLocked: boolean;
  password: string | null;
}

export interface CreateWalletParams {
  name: string;
  password: string;
}

export interface ImportWalletParams {
  name: string;
  mnemonic: string;
  password: string;
}
