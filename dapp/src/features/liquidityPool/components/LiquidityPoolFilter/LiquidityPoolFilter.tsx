import { LiquiditySizeFilterOptions } from '../../types';
import { PoolSizeFilter } from './PoolSizeFilter';

interface LiquidityPoolFilterProps {
  liquiditySize: LiquiditySizeFilterOptions;
  updatePoolSize: (value: [number, number]) => void;
}

export function LiquidityPoolFilter({ liquiditySize, updatePoolSize }: LiquidityPoolFilterProps) {
  return <PoolSizeFilter max={liquiditySize.max} value={liquiditySize.value} onChange={updatePoolSize} />;
}
