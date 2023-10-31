import { Box } from '@mui/material';
import { debounce } from 'lodash';
import { useDispatch, useSelector } from 'react-redux';
import { changeDateFilter, liquidityPoolState, setLiquiditySize, setTokenPair } from '../../slice';
import { TokenPair } from '../../types';
import { DateFilter } from './DateFilter';
import { PoolSizeFilter } from './PoolSizeFilter';
import { TokenPairFilter } from './TokenPairFilter';

export function LiquidityPoolFilter() {
  const { liquiditySizeFilter, dateFilter, tokenPair, tokenPairs } = useSelector(liquidityPoolState);
  const dispatch = useDispatch();

  const handlePoolSizeChange = debounce((value: [number, number]) => {
    dispatch(setLiquiditySize({ value }));
  }, 300);

  const handleDateChange = (hours: number) => {
    dispatch(changeDateFilter({ hours }));
  };

  const handleTokenPairChangeChange = (tokenPair: TokenPair) => {
    dispatch(setTokenPair({ tokenPair }));
  };

  return (
    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
      <TokenPairFilter value={tokenPair} items={tokenPairs} onChange={handleTokenPairChangeChange} />
      <Box sx={{ display: 'flex', alignItems: 'center' }}>
        <PoolSizeFilter
          max={liquiditySizeFilter.max}
          value={liquiditySizeFilter.value}
          onChange={handlePoolSizeChange}
        />
        <Box sx={{ paddingLeft: '24px' }}>
          <DateFilter value={dateFilter.hours} onChange={handleDateChange} />
        </Box>
      </Box>
    </Box>
  );
}
