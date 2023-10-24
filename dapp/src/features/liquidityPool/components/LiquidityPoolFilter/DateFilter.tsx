import { ToggleButton, ToggleButtonGroup } from '@mui/material';

const values = [12, 24, 48];

interface DateFilterProps {
  value: number;
  onChange: (value: number) => void;
}

export function DateFilter({ value, onChange }: DateFilterProps) {
  const handleChange = (_event: React.MouseEvent<HTMLElement>, newValue: number) => {
    if (newValue !== null) {
      onChange(newValue);
    }
  };

  return (
    <ToggleButtonGroup value={value} exclusive onChange={handleChange} data-testid="DateFilter">
      {values.map((hours) => (
        <ToggleButton key={hours} value={hours}>
          {hours}H
        </ToggleButton>
      ))}
    </ToggleButtonGroup>
  );
}
