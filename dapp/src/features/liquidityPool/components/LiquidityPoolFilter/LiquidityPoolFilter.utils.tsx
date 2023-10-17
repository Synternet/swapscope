import { poolSizeSteps } from '../../LiquidityPool.utils';

export function getStepValueText(value: number) {
  return poolSizeSteps[value].label;
}

export function getStepRange(valueRange: [number, number]): [number, number] {
  return [getStepFromValue(valueRange[0]), getStepFromValue(valueRange[1])];
}

export function getStepFromValue(value: number): number {
  return poolSizeSteps.findIndex((x) => x.value === value);
}

export function getValueRange(stepRange: [number, number]): [number, number] {
  return [poolSizeSteps[stepRange[0]].value, poolSizeSteps[stepRange[1]].value];
}
