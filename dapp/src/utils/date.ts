import { isMockedApi } from './env';

export function dateNow() {
  if (isMockedApi()) {
    return new Date('2023-10-11T00:00:00.000Z').getTime();
  }

  return Date.now();
}
