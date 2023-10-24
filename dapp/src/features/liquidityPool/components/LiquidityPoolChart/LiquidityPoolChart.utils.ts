import { formatPoolLimit, formatUsd, isMax, isMin, truncateNumber } from '@src/utils';
import { Data } from 'plotly.js';
import { getPoolItemTotalUsd, itemIsInDateRange } from '../../LiquidityPool.utils';
import { LiquidityPoolItem } from '../../types';
import { differenceInMilliseconds } from 'date-fns';

const chartWidthPx = 1000;
const barWidthPx = 5;

const hoverTemplate = (token1: string, token2: string) => `
<span style="color: white">\
 %{x|%Y-%m-%d %H:%M:%S}<br />\
  Low: %{customdata[0]}<br />\
  High: %{customdata[1]}<br />\
  ${token1} amount: %{customdata[2]}<br />\
  ${token2} amount: %{customdata[3]}<br />\
  Total amount: %{customdata[4]}<br />\
  Actual Price: %{customdata[5]}\
</span><extra></extra>
`;

interface GenerateTracesOptions {
  data: LiquidityPoolItem[];
  filteredData: LiquidityPoolItem[];
  dateRange: [string, string];
  priceRange: [number, number];
}

export function generateTraces({ data, filteredData, dateRange, priceRange }: GenerateTracesOptions): Data[] {
  const diffMs = differenceInMilliseconds(new Date(dateRange[1]), new Date(dateRange[0]));
  const onePixel = diffMs / chartWidthPx;
  const customWidth = Math.round(barWidthPx * onePixel);

  const bars: Data & { base: any[] } = {
    x: filteredData.map((x) => getChartDate(x.timestamp)),
    y: filteredData.map((x) =>
      isMin(x.lowerTokenRatio) || isMax(x.upperTokenRatio)
        ? priceRange[1] - priceRange[0]
        : x.upperTokenRatio - x.lowerTokenRatio,
    ),
    base: filteredData.map((x) => (isMin(x.lowerTokenRatio) ? priceRange[0] : x.lowerTokenRatio)),
    name: 'Added Liquidity',
    type: 'bar',
    width: customWidth,
    hovertemplate: hoverTemplate(filteredData[0]?.pair[0].symbol, filteredData[0]?.pair[1].symbol),
    customdata: filteredData.map((x) => [
      formatPoolLimit(x.lowerTokenRatio),
      formatPoolLimit(x.upperTokenRatio),
      truncateNumber(x.pair[0].amount),
      truncateNumber(x.pair[1].amount),
      formatUsd(getPoolItemTotalUsd(x)),
      truncateNumber(getTokenPrice(x.pair[1].priceUSD, x.pair[0].priceUSD)),
    ]),
    marker: {
      color: filteredData.map(getBarColor),
    },
  };

  const pricePoints = getPricePoints(data, dateRange);
  const line: Data = {
    x: pricePoints.map((x) => getChartDate(x.timestamp)),
    y: pricePoints.map((x) => x.price),
    type: 'scatter',
    name: 'Actual Price',
    hoverinfo: 'skip',
    marker: {
      color: '#4F3F85',
    },
  };

  const filteredMiddlePoints = filteredData.filter((x) => !isMax(x.upperTokenRatio) && !isMin(x.lowerTokenRatio));
  var middle: Data = {
    x: filteredMiddlePoints.map((x) => getChartDate(x.timestamp)),
    y: filteredMiddlePoints.map((x) =>
      isMin(x.lowerTokenRatio) || isMax(x.upperTokenRatio)
        ? (priceRange[1] + priceRange[0]) / 2
        : (x.upperTokenRatio + x.lowerTokenRatio) / 2,
    ),
    mode: 'markers',
    type: 'scatter',
    name: 'Middle',
    hoverinfo: 'skip',
    marker: {
      color: filteredMiddlePoints.map(getMiddlePointColor),
    },
  };

  const traces = [bars, middle, line];

  return traces;
}

function getChartDate(isoDate: string){
  return new Date(isoDate);
}

interface PricePoint {
  timestamp: string;
  price: string;
}

function getPricePoints(data: LiquidityPoolItem[], dateRange: [string, string]): PricePoint[] {
  const uniqueFilteredPrices = data.filter((item, idx) => {
    if (!itemIsInDateRange(item, dateRange)) {
      return false;
    }

    const index = data.findIndex((x) => x.timestamp === item.timestamp);
    return index === idx;
  });

  if (uniqueFilteredPrices.length === 0) {
    return [];
  }

  const mapped: PricePoint[] = uniqueFilteredPrices.map((x) => ({
    timestamp: x.timestamp,
    price: truncateNumber(getTokenPrice(x.pair[1].priceUSD, x.pair[0].priceUSD)),
  }));

  const firstPoint: PricePoint = { timestamp: dateRange[0], price: mapped[0].price };
  const lastPoint: PricePoint = { timestamp: dateRange[1], price: mapped[mapped.length - 1].price };
  return [firstPoint, ...mapped, lastPoint];
}

enum LiquidityPoolItemPosition {
  higher,
  lower,
  neutral,
}

const barColorMap: { [key in LiquidityPoolItemPosition]: string } = {
  [LiquidityPoolItemPosition.higher]: '#1abc9c',
  [LiquidityPoolItemPosition.lower]: '#e74c3c',
  [LiquidityPoolItemPosition.neutral]: '#f39c12',
};

function getBarColor(item: LiquidityPoolItem): string {
  const position = getLiquidityPoolPosition(item);
  return barColorMap[position];
}

const middlePointColorMap: { [key in LiquidityPoolItemPosition]: string } = {
  [LiquidityPoolItemPosition.higher]: '#01775e',
  [LiquidityPoolItemPosition.lower]: '#991208',
  [LiquidityPoolItemPosition.neutral]: '#b87107',
};

function getMiddlePointColor(item: LiquidityPoolItem): string {
  const position = getLiquidityPoolPosition(item);
  return middlePointColorMap[position];
}

function getLiquidityPoolPosition(item: LiquidityPoolItem): LiquidityPoolItemPosition {
  const { lowerTokenRatio, upperTokenRatio, pair } = item;
  const price = getTokenPrice(pair[1].priceUSD, pair[0].priceUSD);

  if (isMin(lowerTokenRatio) && isMax(upperTokenRatio)) {
    return LiquidityPoolItemPosition.neutral;
  }

  const middle = (lowerTokenRatio + upperTokenRatio) / 2;

  if (middle > price) {
    return LiquidityPoolItemPosition.higher;
  }

  return LiquidityPoolItemPosition.lower;
}

function getTokenPrice(token1Usd: number, token2Usd: number) {
  return token1Usd / token2Usd;
}
