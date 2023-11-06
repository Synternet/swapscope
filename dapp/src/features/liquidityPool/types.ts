export interface LiquidityPoolItemToken {
  symbol: string;
  amount: number;
  priceUSD: number;
}

export interface LiquidityPoolItemMessage {
  id: string;
  timestamp: string;
  address: string;
  lowerTokenRatio: number;
  currentTokenRatio: number;
  upperTokenRatio: number;
  valueAddedUSD: number;
  pair: [LiquidityPoolItemToken, LiquidityPoolItemToken];
  txHash: string;
}

export interface LiquidityPoolItem extends LiquidityPoolItemMessage {
  operationType: Exclude<OperationType, 'all'>;
}

export interface LiquiditySizeFilterOptions {
  max: number;
  value: [number, number];
}

export interface TokenPair {
  symbol1: string;
  symbol2: string;
}

export type OperationType = 'all' | 'add' | 'remove';
