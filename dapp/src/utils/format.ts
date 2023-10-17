import { format, parseISO } from 'date-fns';
import { isMax, isMin } from './numbers';

export function truncateAddress(address: string, length = 6) {
  return address.slice(0, length) + '..' + address.slice(address.length - length, address.length);
}

export function truncateNumber(num: number, decimalPlaces = 2) {
  return num.toFixed(decimalPlaces);
}

export function formatDate(dateText: string) {
  const date = parseISO(dateText);
  return format(date, `yyyy-MM-dd HH:mm:ss`);
}

export function formatPoolLimit(num: number) {
  if (isMin(num)) {
    return 'MIN';
  }

  if (isMax(num)) {
    return 'MAX';
  }

  return truncateNumber(num);
}

export function formatUsd(value: number) {
  return Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(value);
}

export function formatUsdCompact(value: number) {
  return Intl.NumberFormat('en-US', { notation: 'compact', style: 'currency', currency: 'USD' }).format(value);
}
