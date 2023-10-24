import { formatUsd, truncateNumber } from '@src/utils';
import { LiquidityPoolItemToken } from '../../types';

interface TokenInfoProps {
  token: LiquidityPoolItemToken;
}

export function TokenInfo({ token }: TokenInfoProps) {
  const { amount, symbol, priceUSD } = token;
  const usdAmount = Math.round(amount * priceUSD);

  return `${truncateNumber(amount, 2)} ${symbol} (${formatUsd(usdAmount)})`;
}
