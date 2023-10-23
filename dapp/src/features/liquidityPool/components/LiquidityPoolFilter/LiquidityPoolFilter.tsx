import { Box } from '@mui/material';
import { debounce } from 'lodash';
import { useDispatch, useSelector } from 'react-redux';
import { changeDateFilter, liquidityPoolState, setLiquiditySize } from '../../slice';
import { DateFilter } from './DateFilter';
import { PoolSizeFilter } from './PoolSizeFilter';

export function LiquidityPoolFilter() {
  const { liquiditySizeFilter, dateFilter } = useSelector(liquidityPoolState);
  const dispatch = useDispatch();

  const handlePoolSizeChange = debounce((value: [number, number]) => {
    dispatch(setLiquiditySize({ value }));
  }, 300);

  const handleDateChange = (hours: number) => {
    dispatch(changeDateFilter({ hours }));
  };

  return (
    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
      <PoolSizeFilter max={liquiditySizeFilter.max} value={liquiditySizeFilter.value} onChange={handlePoolSizeChange} />
      <Box>
        <DateFilter value={dateFilter.hours} onChange={handleDateChange} />
      </Box>
    </Box>
  );
}
