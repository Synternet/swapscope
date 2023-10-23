export interface LiquidityPoolItemToken {
  symbol: string;
  amount: number;
  priceUSD: number;
}

export interface LiquidityPoolItem {
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

export interface LiquiditySizeFilterOptions {
  max: number;
  value: [number, number];
}
