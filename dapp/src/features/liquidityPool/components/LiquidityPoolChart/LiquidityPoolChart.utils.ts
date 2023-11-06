import { formatPoolLimit, formatUsd, isMax, isMin, truncateNumber } from '@src/utils';
import { differenceInMilliseconds } from 'date-fns';
import { Data, PlotData } from 'plotly.js';
import { getPoolItemTotalUsd, itemIsInDateRange, matchTokenPair } from '../../LiquidityPool.utils';
import { LiquidityPoolItem, TokenPair } from '../../types';

const barWidthPx = 5;

const hoverTemplate = (token1: string, token2: string) => `
<span>\
 %{customdata[6]}<br />\
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
  tokenPair: TokenPair;
  chartWidth: number;
}

export function generateTraces({
  data,
  filteredData,
  dateRange,
  priceRange,
  tokenPair,
  chartWidth,
}: GenerateTracesOptions): Data[] {
  const diffMs = differenceInMilliseconds(new Date(dateRange[1]), new Date(dateRange[0]));
  const onePixel = diffMs / chartWidth;
  const customWidth = barWidthPx * onePixel;

  const bars = generateBars({ items: filteredData, priceRange, customWidth });

  const pricePoints = getPricePoints({ data, dateRange, tokenPair });
  const line: Data = {
    x: pricePoints.map((x) => getChartDate(x.timestamp)),
    y: pricePoints.map((x) => x.price),
    type: 'scatter',
    name: 'Actual price',
    hoverinfo: 'skip',
    mode: 'lines',
    marker: {
      color: '#4F3F85',
    },
  };

  const filteredMiddlePoints = filteredData.filter((x) => !isMax(x.upperTokenRatio) && !isMin(x.lowerTokenRatio));
  const middle: Data = {
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

interface GenerateBarsOptions {
  items: LiquidityPoolItem[];
  priceRange: [number, number];
  customWidth: number;
}

function generateBars({ items, priceRange, customWidth }: GenerateBarsOptions): Data & { base: any[] } {
  return {
    x: items.map((x) => getChartDate(x.timestamp)),
    y: items.map((x) =>
      isMin(x.lowerTokenRatio) || isMax(x.upperTokenRatio)
        ? priceRange[1] - priceRange[0]
        : x.upperTokenRatio - x.lowerTokenRatio,
    ),
    base: items.map((x) => (isMin(x.lowerTokenRatio) ? priceRange[0] : x.lowerTokenRatio)),
    name: 'Liquidity add/remove',
    type: 'bar',
    width: customWidth,
    hovertemplate: hoverTemplate(items[0]?.pair[0].symbol, items[0]?.pair[1].symbol),
    customdata: items.map((x) => [
      formatPoolLimit(x.lowerTokenRatio),
      formatPoolLimit(x.upperTokenRatio),
      truncateNumber(x.pair[0].amount),
      truncateNumber(x.pair[1].amount),
      formatUsd(getPoolItemTotalUsd(x)),
      truncateNumber(getTokenPrice(x.pair[1].priceUSD, x.pair[0].priceUSD)),
      x.operationType === 'add' ? "Liquidity add" : 'Liquidity remove',
    ]),
    marker: getBarMarker(items),
  };
}

export function getChartDate(isoDate: string) {
  return new Date(isoDate);
}

interface PricePoint {
  timestamp: string;
  price: string;
}

interface GetPricePointsOptions {
  data: LiquidityPoolItem[];
  dateRange: [string, string];
  tokenPair: TokenPair;
}

function getPricePoints(options: GetPricePointsOptions): PricePoint[] {
  const { data, dateRange, tokenPair } = options;
  const tokenPairItems = data.filter((x) => matchTokenPair(x, tokenPair));

  const uniqueFilteredPrices = tokenPairItems.filter((item, idx) => {
    if (!itemIsInDateRange(item, dateRange)) {
      return false;
    }

    const index = tokenPairItems.findIndex((x) => x.timestamp === item.timestamp);
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

const barColorMap: { [key in LiquidityPoolItemPosition]: { strong: string; weak: string } } = {
  [LiquidityPoolItemPosition.higher]: { strong: '#1abc9c', weak: 'rgba(26, 188, 156, 0.2)' },
  [LiquidityPoolItemPosition.lower]: { strong: '#e74c3c', weak: 'rgba(231, 76, 60, 0.2)' },
  [LiquidityPoolItemPosition.neutral]: { strong: '#f39c12', weak: 'rgba(243, 156, 18, 0.2)' },
};

function getBarMarker(items: LiquidityPoolItem[]): PlotData['marker'] {
  return {
    color: items.map((x) => {
      const position = getLiquidityPoolPosition(x);
      return x.operationType === 'add' ? barColorMap[position].strong : barColorMap[position].weak;
    }),
    line: {
      color: items.map((x) => {
        const position = getLiquidityPoolPosition(x);
        return barColorMap[position].strong;
      }),
      width: items.map((x) => {
        return x.operationType === 'add' ? 0 : 1;
      }),
    },
  };
}

const middlePointColorMap: { [key in LiquidityPoolItemPosition]: { strong: string; weak: string } } = {
  [LiquidityPoolItemPosition.higher]: { strong: '#01775e', weak: '#1abc9c' },
  [LiquidityPoolItemPosition.lower]: { strong: '#991208', weak: '#e74c3c' },
  [LiquidityPoolItemPosition.neutral]: { strong: '#b87107', weak: '#f39c12' },
};

function getMiddlePointColor(item: LiquidityPoolItem): string {
  const position = getLiquidityPoolPosition(item);
  return item.operationType === 'add' ? middlePointColorMap[position].strong : middlePointColorMap[position].weak;
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

export function getChartWidth(chartElement?: HTMLDivElement): number | undefined {
  if (!chartElement) {
    return;
  }

  const gridLayer = chartElement.querySelector('.gridlayer');
  if (!gridLayer) {
    return;
  }

  const { width } = gridLayer.getBoundingClientRect();
  return width;
}
