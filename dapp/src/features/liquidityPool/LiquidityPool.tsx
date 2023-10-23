import { Box, Container, styled } from '@mui/material';
import { useEffect, useMemo, useRef } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { filterItems } from './LiquidityPool.utils';
import { LiquidityPoolChart, LiquidityPoolFilter, LiquidityPoolTable } from './components';
import { liquidityPoolState, loadData } from './slice';
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
  const {
    items,
    liquiditySizeFilter,
    dateFilter,
    priceRange,
  } = useSelector(liquidityPoolState);
  const dispatch = useDispatch();
  const initialized = useRef(false);

  // @TODO move initialization somewhere else
  useEffect(() => {
    if (initialized.current) {
      return;
    }
    initialized.current = true;
    dispatch(loadData());
  }, [dispatch]);

  const filteredList: LiquidityPoolItem[] = useMemo(
    () => filterItems(items, { liquiditySize: liquiditySizeFilter.value, dateRange: dateFilter.range }),
    [items, liquiditySizeFilter, dateFilter],
  );

  return (
    <>
      <Placeholder>SwapScope</Placeholder>
      <Container maxWidth="xl">
        <LiquidityPoolChart data={items} filteredData={filteredList} dateRange={dateFilter.range} priceRange={priceRange} />
        <Box sx={{ my: '20px' }}>
          <LiquidityPoolFilter />
        </Box>
        <Box sx={{ my: '20px' }}>
          <LiquidityPoolTable list={filteredList} />
        </Box>
      </Container>
    </>
  );
}
