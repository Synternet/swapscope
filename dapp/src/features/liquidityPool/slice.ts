import { PayloadAction, createAction, createSlice } from '@reduxjs/toolkit';
import { dateNow } from '@src/utils';
import { addHours } from 'date-fns';
import { getPoolSizeFilterOptions, getPriceRange, poolSizeSteps } from './LiquidityPool.utils';
import { LiquidityPoolItem, LiquiditySizeFilterOptions } from './types';

interface LiquidityPoolState {
  items: LiquidityPoolItem[];
  dateFilter: { hours: number; range: [string, string] };
  priceRange: [number, number];
  liquiditySizeFilter: LiquiditySizeFilterOptions;
  revision: string;
}

const initialState = (): LiquidityPoolState => ({
  items: [],
  dateFilter: { hours: 12, range: [addHours(dateNow(), -12).toISOString(), new Date(dateNow()).toISOString()] },
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
      state.dateFilter.range = [
        addHours(dateNow(), -state.dateFilter.hours).toISOString(),
        new Date(dateNow()).toISOString(),
      ];
      state.priceRange = getPriceRange(items);
      state.revision = Date.now().toString();
    },
    changeDateFilter: (state, { payload }: PayloadAction<{ hours: number }>) => {
      const { hours } = payload;
      state.dateFilter = {
        hours: hours,
        range: [addHours(dateNow(), -hours).toISOString(), new Date(dateNow()).toISOString()],
      };
    },
    resetLiquidityPoolState: () => initialState(),
  },
});

export const liquidityPoolReducer = slice.reducer;
export const liquidityPoolState = (state: RootState) => state.liquidityPool;

export const { setLiquiditySize, setLiquidityPoolItems, changeDateFilter } = slice.actions;

const prefix = slice.name;
export const loadData = createAction(`${prefix}/loadData`);
