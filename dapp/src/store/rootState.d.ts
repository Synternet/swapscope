import { compose } from 'redux';
import type { AppStartListening, AppState } from './store';

declare global {
  interface Window {
    __REDUX_DEVTOOLS_EXTENSION_COMPOSE__?: typeof compose;
  }

  type RootState = AppState;
  type ListenState = AppStartListening;
}
