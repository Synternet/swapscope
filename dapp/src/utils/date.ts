import { getMockedDateNow } from '@src/modules';
import { isMockedApi } from './env';

export function dateNow() {
  if (isMockedApi()) {
    return getMockedDateNow();
  }

  return Date.now();
}
