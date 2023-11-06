import { Box } from '@mui/material';
import { debounce } from 'lodash';
import { useDispatch, useSelector } from 'react-redux';
import { changeDateFilter, liquidityPoolState, setLiquiditySize, setOperationType, setTokenPair } from '../../slice';
import { OperationType, TokenPair } from '../../types';
import { DateFilter } from './DateFilter';
import { LiquiditySizeFilter } from './LiquiditySizeFilter';
import { OperationTypeFilter } from './OperationTypeFilter';
import { TokenPairFilter } from './TokenPairFilter';

export function LiquidityPoolFilter() {
  const { liquiditySizeFilter, dateFilter, tokenPair, tokenPairs, operationType } = useSelector(liquidityPoolState);
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

  const handleOperationTypeChange = (operationType: OperationType) => {
    dispatch(setOperationType({ operationType }));
  };

  return (
    <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
      <TokenPairFilter value={tokenPair} items={tokenPairs} onChange={handleTokenPairChangeChange} />
      <Box sx={{ display: 'flex', alignItems: 'center' }}>
        <LiquiditySizeFilter
          max={liquiditySizeFilter.max}
          value={liquiditySizeFilter.value}
          onChange={handlePoolSizeChange}
        />
        <Box sx={{ paddingLeft: '24px' }}>
          <OperationTypeFilter value={operationType} onChange={handleOperationTypeChange} />
        </Box>
      </Box>
      <Box sx={{ paddingLeft: '24px' }}>
        <DateFilter value={dateFilter.hours} onChange={handleDateChange} />
      </Box>
    </Box>
  );
}
