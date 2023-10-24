import { isMockedApi } from './env';

export function dateNow() {
  if (isMockedApi()) {
    return new Date('2023-10-11T00:00:00.000Z').getTime();
  }

  return Date.now();
}

// @TODO change this when local time used instead of ISO/UTC
export function convertToIso(date: string) {
  if (isIsoDate(date)) {
    return date;
  }

  const isoDate = date.replace(' ', 'T') + 'Z';
  return isoDate;
}

function isIsoDate(date: string) {
  const isoDateRegex = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(\.\d{1,6})?Z$/;
  return isoDateRegex.test(date);
}
