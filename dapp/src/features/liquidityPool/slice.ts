import { PayloadAction, createAction, createSlice } from '@reduxjs/toolkit';
import { dateNow } from '@src/utils';
import { addHours } from 'date-fns';
import { getPoolSizeFilterOptions, getPriceRange, poolSizeSteps } from './LiquidityPool.utils';
import { LiquidityPoolItem, LiquiditySizeFilterOptions, TokenPair } from './types';

interface LiquidityPoolState {
  items: LiquidityPoolItem[];
  dateFilter: { hours: number; range: [string, string] };
  priceRange: [number, number];
  liquiditySizeFilter: LiquiditySizeFilterOptions;
  tokenPair: TokenPair;
  tokenPairs: TokenPair[];
  revision: string;
}

export const defaultTokenPair: TokenPair = { symbol1: 'USDC', symbol2: 'WETH' };

const initialState = (): LiquidityPoolState => ({
  items: [],
  dateFilter: { hours: 12, range: [addHours(dateNow(), -12).toISOString(), new Date(dateNow()).toISOString()] },
  priceRange: [0, 100],
  liquiditySizeFilter: {
    max: poolSizeSteps[poolSizeSteps.length - 1].value,
    value: [0, poolSizeSteps[poolSizeSteps.length - 1].value],
  },
  tokenPair: defaultTokenPair,
  tokenPairs: [defaultTokenPair],
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
      state.liquiditySizeFilter = getPoolSizeFilterOptions(items, state.tokenPair);
      state.dateFilter.range = [
        addHours(dateNow(), -state.dateFilter.hours).toISOString(),
        new Date(dateNow()).toISOString(),
      ];
      state.priceRange = getPriceRange(items, state.tokenPair);
      state.revision = Date.now().toString();
    },
    changeDateFilter: (state, { payload }: PayloadAction<{ hours: number }>) => {
      const { hours } = payload;
      state.dateFilter = {
        hours: hours,
        range: [addHours(dateNow(), -hours).toISOString(), new Date(dateNow()).toISOString()],
      };
    },
    setTokenPairs: (state, { payload }: PayloadAction<{ tokenPairs: TokenPair[] }>) => {
      state.tokenPairs = payload.tokenPairs;
    },
    setTokenPair: (state, { payload }: PayloadAction<{ tokenPair: TokenPair }>) => {
      state.tokenPair = payload.tokenPair;
      state.liquiditySizeFilter = getPoolSizeFilterOptions(state.items, state.tokenPair);
      state.priceRange = getPriceRange(state.items, state.tokenPair);
      state.revision = Date.now().toString();
    },
    resetLiquidityPoolState: () => initialState(),
  },
});

export const liquidityPoolReducer = slice.reducer;
export const liquidityPoolState = (state: RootState) => state.liquidityPool;

export const { setLiquiditySize, setLiquidityPoolItems, changeDateFilter, setTokenPairs, setTokenPair } = slice.actions;

const prefix = slice.name;
export const loadData = createAction(`${prefix}/loadData`);
