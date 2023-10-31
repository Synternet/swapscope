import { Box, Slider, Typography } from '@mui/material';
import { useEffect, useState } from 'react';
import { poolSizeSteps } from '../../LiquidityPool.utils';
import { getStepFromValue, getStepRange, getStepValueText, getValueRange } from './LiquidityPoolFilter.utils';

interface PoolSizeFilterProps {
  max: number;
  value: [number, number];
  onChange: (value: [number, number]) => void;
}

export function PoolSizeFilter(props: PoolSizeFilterProps) {
  const { max, value: initialValue, onChange } = props;
  const [value, setValue] = useState<[number, number]>(getStepRange(initialValue));
  const marks = poolSizeSteps.map((x, idx) => ({ value: idx, label: x.label }));

  const handleChange = (_event: Event, newValue: number | number[]) => {
    const stepRange = newValue as [number, number];
    const valueRange = getValueRange(stepRange);
    setValue(stepRange);
    onChange(valueRange);
  };

  useEffect(() => {
    setValue(getStepRange(initialValue));
  }, [initialValue]);

  return (
    <Box sx={{ width: 300, padding: '0 24px', boxSizing: 'content-box' }}>
      <Typography gutterBottom>Liquidity add size</Typography>
      <Slider
        data-testid="PoolSizeFilter"
        onChange={handleChange}
        valueLabelDisplay="auto"
        getAriaValueText={getStepValueText}
        valueLabelFormat={getStepValueText}
        value={value}
        marks={marks}
        min={0}
        max={getStepFromValue(max)}
        step={1}
      />
    </Box>
  );
}
