import { formatPoolLimit, formatUsd, isMax, isMin, truncateNumber } from '@src/utils';
import { Data } from 'plotly.js';
import { getPoolItemTotalUsd } from '../../LiquidityPool.utils';
import { LiquidityPoolItem } from '../../types';

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
  const diffMs = new Date(dateRange[1]).getTime() - new Date(dateRange[0]).getTime();
  const onePixel = diffMs / chartWidthPx;
  const customWidth = Math.round(barWidthPx * onePixel);

  const bars: Data & { base: any[] } = {
    x: filteredData.map((x) => x.timestamp),
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
      truncateNumber(getTokenPrice(x.pair[1].price, x.pair[0].price)),
    ]),
    marker: {
      color: filteredData.map(getBarColor),
    },
  };

  const uniquePrices = data.filter((item, idx) => {
    const index = data.findIndex((x) => x.timestamp === item.timestamp);
    return index === idx;
  });

  const line: Data = {
    x: uniquePrices.map((x) => x.timestamp),
    y: uniquePrices.map((x) => truncateNumber(getTokenPrice(x.pair[1].price, x.pair[0].price))),
    type: 'scatter',
    name: 'Actual Price',
    hoverinfo: 'skip',
    marker: {
      color: '#4F3F85',
    },
  };

  const filteredMiddlePoints = filteredData.filter((x) => !isMax(x.upperTokenRatio) && !isMin(x.lowerTokenRatio));
  var middle: Data = {
    x: filteredMiddlePoints.map((x) => x.timestamp),
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
  const price = getTokenPrice(pair[1].price, pair[0].price);

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
