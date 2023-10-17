import { PayloadAction, createAction, createSlice } from '@reduxjs/toolkit';
import { addHours } from 'date-fns';
import { getPoolSizeFilterOptions, getPriceRange, poolSizeSteps } from './LiquidityPool.utils';
import { LiquidityPoolItem, LiquiditySizeFilterOptions } from './types';

interface LiquidityPoolState {
  items: LiquidityPoolItem[];
  dateRange: [string, string];
  priceRange: [number, number];
  liquiditySizeFilter: LiquiditySizeFilterOptions;
  revision: string;
}

const initialState = (): LiquidityPoolState => ({
  items: [],
  dateRange: [addHours(new Date(), -24).toISOString(), new Date().toISOString()],
  priceRange: [0, 100],
  liquiditySizeFilter: {
    max: poolSizeSteps[poolSizeSteps.length - 1].value,
    value: [0, poolSizeSteps[poolSizeSteps.length - 1].value],
  },
  revision: 'initial',
});

const slice = createSlice({
  name: 'liquidityPool',
  initialState: initialState(),
  reducers: {
    setLiquiditySize: (state, { payload }: PayloadAction<{ value: [number, number] }>) => {
      state.liquiditySizeFilter.value = payload.value;
    },
    setLiquidityPoolItems: (state, { payload }: PayloadAction<{ items: LiquidityPoolItem[] }>) => {
      const { items } = payload;
      state.items = items;
      state.liquiditySizeFilter = getPoolSizeFilterOptions(items);
      state.dateRange = [items[0].timestamp, items[items.length - 1].timestamp];
      state.priceRange = getPriceRange(items);
      state.revision = Date.now().toString();
    },
    resetLiquidityPoolState: () => initialState(),
  },
});

export const liquidityPoolReducer = slice.reducer;
export const liquidityPoolState = (state: RootState) => state.liquidityPool;

export const { setLiquiditySize, setLiquidityPoolItems } = slice.actions;

const prefix = slice.name;
export const loadData = createAction(`${prefix}/loadData`);
