export interface Transaction {
  hash: string;
  height: number;
  timestamp: string;
  type: string;
  from: string;
  to?: string;
  amount?: {
    denom: string;
    amount: string;
  };
  fee: {
    denom: string;
    amount: string;
  };
  status: 'pending' | 'success' | 'failed';
  memo?: string;
}

export interface SendParams {
  to: string;
  amount: string;
  denom: string;
  memo?: string;
}

export interface TransactionResult {
  hash: string;
  height: number;
  success: boolean;
  error?: string;
}
