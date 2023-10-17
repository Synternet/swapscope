import { registerMessageListener } from './registerMessageListener';

export function registerLiquidityPoolListeners(startListening: ListenState) {
  registerMessageListener(startListening);
}
