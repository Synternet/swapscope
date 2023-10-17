import { Box, Container, styled } from '@mui/material';
import { debounce } from 'lodash';
import { useEffect, useMemo, useRef } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { getPoolItemTotalUsd } from './LiquidityPool.utils';
import { LiquidityPoolChart, LiquidityPoolFilter, LiquidityPoolTable } from './components';
import { liquidityPoolState, loadData, setLiquiditySize } from './slice';
import { LiquidityPoolItem } from './types';

const Placeholder = styled(Box)(({ theme }) => ({
  padding: '16px',
  textAlign: 'center',
  background: theme.palette.grey[200],
  display: 'flex',
  alignItems: 'center',
  justifyContent: 'center',
}));

export function LiquidityPool() {
  const { items, liquiditySizeFilter, dateRange, priceRange } = useSelector(liquidityPoolState);
  const dispatch = useDispatch();
  const initialized = useRef(false);

  const handlePoolSizeUpdate = debounce((value: [number, number]) => {
    dispatch(setLiquiditySize({ value }));
  }, 300);

  // @TODO move initialization somewhere else
  useEffect(() => {
    if (initialized.current) {
      return;
    }
    initialized.current = true;
    dispatch(loadData());
  }, [dispatch]);

  const filteredList: LiquidityPoolItem[] = useMemo(() => {
    return items.filter((item) => {
      const amount = getPoolItemTotalUsd(item);
      if (amount > liquiditySizeFilter.value[1] || amount < liquiditySizeFilter.value[0]) {
        return false;
      }

      return true;
    });
  }, [items, liquiditySizeFilter]);

  return (
    <>
      <Placeholder>SwapScope</Placeholder>
      <Container maxWidth="xl">
        <LiquidityPoolChart data={items} filteredData={filteredList} dateRange={dateRange} priceRange={priceRange} />
        <Box sx={{ my: '20px', height: '50px' }}>
          <LiquidityPoolFilter liquiditySize={liquiditySizeFilter} updatePoolSize={handlePoolSizeUpdate} />
        </Box>
        <Box sx={{ my: '20px' }}>
          <LiquidityPoolTable list={filteredList} />
        </Box>
      </Container>
    </>
  );
}
