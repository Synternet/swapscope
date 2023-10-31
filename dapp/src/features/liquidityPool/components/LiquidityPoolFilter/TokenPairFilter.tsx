import { ToggleButton, ToggleButtonGroup } from '@mui/material';
import { TokenPair } from '../../types';

interface TokenPairFilterProps {
  value: TokenPair;
  items: TokenPair[];
  onChange: (value: TokenPair) => void;
}

export function TokenPairFilter({ value, items, onChange }: TokenPairFilterProps) {
  const handleChange = (_event: React.MouseEvent<HTMLElement>, newValue: TokenPair) => {
    if (newValue !== null) {
      onChange(newValue);
    }
  };

  return (
    <ToggleButtonGroup value={value} exclusive onChange={handleChange} data-testid="TokenPairFilter">
      {items.map((tokenPair) => (
        <ToggleButton key={tokenPair.symbol1+tokenPair.symbol2} value={tokenPair}>
          {tokenPair.symbol1}/{tokenPair.symbol2}
        </ToggleButton>
      ))}
    </ToggleButtonGroup>
  );
}
