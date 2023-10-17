import { formatUsd, truncateNumber } from '@src/utils';
import { LiquidityPoolItemToken } from '../../types';

interface TokenInfoProps {
  token: LiquidityPoolItemToken;
}

export function TokenInfo({ token }: TokenInfoProps) {
  const { amount, symbol, price } = token;
  const usdAmount = Math.round(amount * price);

  return `${truncateNumber(amount, 2)} ${symbol} (${formatUsd(usdAmount)})`;
}
