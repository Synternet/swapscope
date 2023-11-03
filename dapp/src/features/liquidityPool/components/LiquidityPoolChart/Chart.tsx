import { PlotRelayoutEvent } from 'plotly.js';
import { useCallback, useEffect, useRef, useState } from 'react';
import Plot from 'react-plotly.js';
import { useSelector } from 'react-redux';
import { liquidityPoolState } from '../../slice';
import { LiquidityPoolItem, TokenPair } from '../../types';
import { generateTraces, getChartWidth } from './LiquidityPoolChart.utils';

interface ChartProps {
  dateRange: [string, string];
  priceRange: [number, number];
  data: LiquidityPoolItem[];
  filteredData: LiquidityPoolItem[];
  tokenPair: TokenPair;
}

export function Chart(props: ChartProps) {
  const { data, filteredData, dateRange: initialDateRange, priceRange, tokenPair } = props;
  const { revision } = useSelector(liquidityPoolState);
  const [dateRange, setDateRange] = useState([...initialDateRange] as [string, string]);
  const [chartWidth, setChartWidth] = useState(0);
  const traces = generateTraces({ data, filteredData, dateRange, priceRange, tokenPair, chartWidth });
  const chartRef = useRef<Plot & { el: HTMLDivElement }>(null);

  const checkChartWidth = useCallback(() => {
    const newWidth = getChartWidth(chartRef.current?.el);
    if (newWidth && chartWidth !== newWidth) {
      setChartWidth(newWidth);
    }
  }, [chartWidth]);

  const handleRelayout = useCallback(
    (event: Readonly<PlotRelayoutEvent>) => {
      checkChartWidth();
      const xStart = event['xaxis.range[0]'] ?? event['xaxis.range']?.[0];
      const xEnd = event['xaxis.range[1]'] ?? event['xaxis.range']?.[1];
      if (xStart && xEnd) {
        setDateRange([new Date(xStart).toISOString(), new Date(xEnd).toISOString()]);
      }

      if (event['xaxis.autorange']) {
        setDateRange(initialDateRange);
      }
    },
    [initialDateRange, checkChartWidth],
  );

  const handleInitialized = useCallback(() => {
    checkChartWidth();
  }, [checkChartWidth]);

  useEffect(() => {
    setDateRange(initialDateRange);
  }, [initialDateRange]);

  return (
    <Plot
      ref={chartRef}
      data={traces}
      style={{ width: '100%', height: '100%' }}
      config={{ responsive: true, displayModeBar: true, displaylogo: false }}
      layout={{
        uirevision: revision,
        legend: {
          x: 1,
          y: 1.1,
          xanchor: 'right',
        },
        xaxis: {
          type: 'date',
        },
        yaxis: {},
        margin: {
          l: 40,
          r: 0,
          b: 40,
          t: 50,
          pad: 0,
        },
      }}
      onRelayout={handleRelayout}
      onInitialized={handleInitialized}
    />
  );
}
