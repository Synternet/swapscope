import { PayloadAction, createAction, createSlice } from '@reduxjs/toolkit';
import { getDateFilter, getPoolSizeFilterOptions, getPriceRange, poolSizeSteps } from './LiquidityPool.utils';
import { LiquidityPoolItem, LiquiditySizeFilterOptions, OperationType, TokenPair } from './types';

interface LiquidityPoolState {
  items: LiquidityPoolItem[];
  dateFilter: { hours: number; range: [string, string] };
  priceRange: [number, number];
  liquiditySizeFilter: LiquiditySizeFilterOptions;
  tokenPair: TokenPair;
  tokenPairs: TokenPair[];
  operationType: OperationType;
  revision: string;
}

export const defaultTokenPair: TokenPair = { symbol1: 'WETH', symbol2: 'USDC' };

const initialState = (): LiquidityPoolState => ({
  items: [],
  dateFilter: getDateFilter(12),
  priceRange: [0, 100],
  liquiditySizeFilter: {
    max: poolSizeSteps[poolSizeSteps.length - 1].value,
    value: [0, poolSizeSteps[poolSizeSteps.length - 1].value],
  },
  tokenPair: defaultTokenPair,
  tokenPairs: [defaultTokenPair],
  operationType: 'all',
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
      state.dateFilter = getDateFilter(state.dateFilter.hours);
      state.priceRange = getPriceRange(items, state.tokenPair);
      state.revision = Date.now().toString();
    },
    changeDateFilter: (state, { payload }: PayloadAction<{ hours: number }>) => {
      const { hours } = payload;
      state.dateFilter = getDateFilter(hours);
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
    setOperationType: (state, { payload }: PayloadAction<{ operationType: OperationType }>) => {
      state.operationType = payload.operationType;
    },
    resetLiquidityPoolState: () => initialState(),
  },
});

export const liquidityPoolReducer = slice.reducer;
export const liquidityPoolState = (state: RootState) => state.liquidityPool;

export const {
  setLiquiditySize,
  setLiquidityPoolItems,
  changeDateFilter,
  setTokenPairs,
  setTokenPair,
  setOperationType,
} = slice.actions;

const prefix = slice.name;
export const loadData = createAction(`${prefix}/loadData`);
