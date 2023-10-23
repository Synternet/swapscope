import { formatUsdCompact, isMax, isMin } from '@src/utils';
import { isWithinInterval } from 'date-fns';
import { LiquidityPoolItem, LiquiditySizeFilterOptions } from './types';

const poolSizeValues = [0, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9];

export const poolSizeSteps = poolSizeValues.map((x) => ({
  value: x,
  label: formatUsdCompact(x),
}));

export function getPoolSizeFilterOptions(list: LiquidityPoolItem[]): LiquiditySizeFilterOptions {
  let max = 0;
  list.forEach((item) => {
    const totalAmount = getPoolItemTotalUsd(item);

    if (totalAmount > max) {
      max = Math.ceil(totalAmount);
    }
  });

  const maxMark = poolSizeSteps.find((x) => x.value > max)!.value;

  return { max: maxMark, value: [0, maxMark] };
}

export function getPoolItemTotalUsd(item: LiquidityPoolItem) {
  const { pair } = item;
  return pair[0].priceUSD * pair[0].amount + pair[1].priceUSD * pair[1].amount;
}

export function getPriceRange(list: LiquidityPoolItem[]): [number, number] {
  const min = list.reduce(
    (acc, { lowerTokenRatio }) => (!isMin(lowerTokenRatio) && lowerTokenRatio < acc ? lowerTokenRatio : acc),
    Number.MAX_SAFE_INTEGER,
  );

  const max = list.reduce(
    (acc, { upperTokenRatio }) => (!isMax(upperTokenRatio) && upperTokenRatio > acc ? upperTokenRatio : acc),
    Number.MIN_SAFE_INTEGER,
  );

  return [min, max];
}

export function filterItems(
  items: LiquidityPoolItem[],
  filter: {
    liquiditySize: [number, number];
    dateRange: [string, string];
  },
) {
  const { liquiditySize, dateRange } = filter;
  return items.filter((item) => {
    const amount = getPoolItemTotalUsd(item);
    if (amount > liquiditySize[1] || amount < liquiditySize[0]) {
      return false;
    }

    if (!itemIsInDateRange(item, dateRange)) {
      return false;
    }

    return true;
  });
}

export function itemIsInDateRange(item: LiquidityPoolItem, dateRange: [string, string]){
  return isWithinInterval(new Date(item.timestamp), { start: new Date(dateRange[0]), end: new Date(dateRange[1]) });
}