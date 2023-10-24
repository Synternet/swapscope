import { convertToIso } from '@src/utils';
import { PlotRelayoutEvent } from 'plotly.js';
import { useCallback, useEffect, useState } from 'react';
import Plot from 'react-plotly.js';
import { useSelector } from 'react-redux';
import { liquidityPoolState } from '../../slice';
import { LiquidityPoolItem } from '../../types';
import { generateTraces } from './LiquidityPoolChart.utils';

interface ChartProps {
  dateRange: [string, string];
  priceRange: [number, number];
  data: LiquidityPoolItem[];
  filteredData: LiquidityPoolItem[];
}

export function Chart({ data, filteredData, dateRange: initialDateRange, priceRange }: ChartProps) {
  const { revision } = useSelector(liquidityPoolState);
  const [dateRange, setDateRange] = useState([...initialDateRange] as [string, string]);
  const traces = generateTraces({ data, filteredData, dateRange, priceRange });

  const handleRelayout = useCallback(
    (event: Readonly<PlotRelayoutEvent>) => {
      const xStart = event['xaxis.range[0]'] ?? event['xaxis.range']?.[0];
      const xEnd = event['xaxis.range[1]'] ?? event['xaxis.range']?.[1];
      if (xStart && xEnd) {
        setDateRange([convertToIso(xStart as string), convertToIso(xEnd as string)]);
      }

      if (event['xaxis.autorange']) {
        setDateRange(initialDateRange);
      }
    },
    [initialDateRange],
  );

  useEffect(() => {
    setDateRange(initialDateRange);
  }, [initialDateRange]);

  return (
    <Plot
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
          range: [...dateRange],
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
    />
  );
}
