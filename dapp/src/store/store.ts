import { TypedStartListening, configureStore, createListenerMiddleware } from '@reduxjs/toolkit';
import { liquidityPoolReducer, registerLiquidityPoolListeners } from '@src/features';

const listenerMiddleware = createListenerMiddleware();

const createStore = () =>
  configureStore({
    reducer: { liquidityPool: liquidityPoolReducer },
    devTools: true, // @TODO check env
    middleware: (getDefaultMiddleware) => getDefaultMiddleware().prepend(listenerMiddleware.middleware),
  });

export const store = createStore();

type AppDispatch = typeof store.dispatch;
export type AppState = ReturnType<typeof store.getState>;
export type AppStartListening = TypedStartListening<AppState, AppDispatch>;
const listen = listenerMiddleware.startListening as AppStartListening;

registerLiquidityPoolListeners(listen);
